package server

import (
	"autonity-oracle/contract_binder/contract"
	"autonity-oracle/monitor"
	"autonity-oracle/types"
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/metrics"
)

// sync is executed on client startup or after the L1 connection recovery to sync the on-chain oracle contract
// states, symbols, round id, precision, vote period, etc... to the oracle server. It also subscribes the on-chain
// events of oracle protocol: round event, symbol update event, etc...
func (os *Server) sync() error {
	var err error
	// get initial states from oracle contract.
	os.curRoundHeight, os.curRound, os.protocolSymbols, os.votePeriod, err = os.syncRoundState()
	if err != nil {
		os.logger.Error("synchronize oracle contract state", "error", err.Error())
		return err
	}

	// reset sampling symbols with the latest protocol symbols, it adds bridger symbols by according to the protocol symbols.
	os.resetSamplingSymbols(os.protocolSymbols)

	// subscribe protocol events
	if err = os.subscribeEvents(); err != nil {
		return err
	}
	os.logger.Info("synced", "CurrentRoundHeight", os.curRoundHeight, "CurrentRound", os.curRound,
		"protocol symbols", os.protocolSymbols, "sampling symbols", os.samplingSymbols)
	return nil
}

func (os *Server) subscribeEvents() error {
	// subscribe on-chain round rotation event
	chRoundEvent := make(chan *oracle.OracleNewRound)
	subRoundEvent, err := os.oracleContract.WatchNewRound(new(bind.WatchOpts), chRoundEvent)
	if err != nil {
		os.logger.Error("failed to subscribe round event", "error", err.Error())
		return err
	}
	os.chRoundEvent = chRoundEvent
	os.subRoundEvent = subRoundEvent

	// subscribe on-chain symbol update event
	chSymbolsEvent := make(chan *oracle.OracleNewSymbols)
	subSymbolsEvent, err := os.oracleContract.WatchNewSymbols(new(bind.WatchOpts), chSymbolsEvent)
	if err != nil {
		os.logger.Error("failed to subscribe new symbol event", "error", err.Error())
		return err
	}
	os.chSymbolsEvent = chSymbolsEvent
	os.subSymbolsEvent = subSymbolsEvent

	// subscribe on-chain no-reveal event
	chNoRevealEvent := make(chan *oracle.OracleNoRevealPenalty)
	subNoRevealEvent, err := os.oracleContract.WatchNoRevealPenalty(new(bind.WatchOpts), chNoRevealEvent, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe no reveal event", "error", err.Error())
		return err
	}
	os.chNoRevealEvent = chNoRevealEvent
	os.subNoRevealEvent = subNoRevealEvent

	// subscribe on-chain penalize event
	chPenalizedEvent := make(chan *oracle.OraclePenalized)
	subPenalizedEvent, err := os.oracleContract.WatchPenalized(new(bind.WatchOpts), chPenalizedEvent, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe penalized event", "error", err.Error())
		return err
	}
	os.chPenalizedEvent = chPenalizedEvent
	os.subPenalizedEvent = subPenalizedEvent

	// subscribe voted event
	chVotedEvent := make(chan *oracle.OracleSuccessfulVote)
	subVotedEvent, err := os.oracleContract.WatchSuccessfulVote(new(bind.WatchOpts), chVotedEvent, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe voted event", "error", err.Error())
		return err
	}
	os.chVotedEvent = chVotedEvent
	os.subVotedEvent = subVotedEvent

	// subscribe invalid vote event
	chInvalidVote := make(chan *oracle.OracleInvalidVote)
	subInvalidVote, err := os.oracleContract.WatchInvalidVote(new(bind.WatchOpts), chInvalidVote, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe invalid vote event", "error", err.Error())
		return err
	}
	os.chInvalidVote = chInvalidVote
	os.subInvalidVote = subInvalidVote

	// subscribe reward event
	chRewardEvent := make(chan *oracle.OracleTotalOracleRewards)
	subRewardEvent, err := os.oracleContract.WatchTotalOracleRewards(new(bind.WatchOpts), chRewardEvent)
	if err != nil {
		os.logger.Error("failed to subscribe reward event", "error", err.Error())
		return err
	}
	os.chRewardEvent = chRewardEvent
	os.subRewardEvent = subRewardEvent
	return nil
}

// syncRoundState returns round id, symbols and vote period on oracle contract, it is called on the startup of client.
// Since below steps are not atomic get operation from blockchain, thus they are just being used at the initial phase
// for data presampling, the correctness of voting is promised by the synchronization triggered by the round event before
// the voting.
func (os *Server) syncRoundState() (uint64, uint64, []string, uint64, error) {
	// on the startup, we need to sync the round block, round id, symbols and committees from contract.
	currentRoundHeight, err := os.oracleContract.GetLastRoundBlock(nil)
	if err != nil {
		os.logger.Error("get round block", "error", err.Error())
		return 0, 0, nil, 0, err
	}

	currentRound, err := os.oracleContract.GetRound(nil)
	if err != nil {
		os.logger.Error("get round", "error", err.Error())
		return 0, 0, nil, 0, err
	}

	symbols, err := os.oracleContract.GetSymbols(nil)
	if err != nil {
		os.logger.Error("get symbols", "error", err.Error())
		return 0, 0, nil, 0, err
	}

	votePeriod, err := os.oracleContract.GetVotePeriod(nil)
	if err != nil {
		os.logger.Error("get vote period", "error", err.Error())
		return 0, 0, nil, 0, nil
	}

	if len(symbols) == 0 {
		os.logger.Error("there are no symbols in Autonity L1 oracle contract")
		return currentRoundHeight.Uint64(), currentRound.Uint64(), symbols, votePeriod.Uint64(), types.ErrNoSymbolsObserved
	}

	return currentRoundHeight.Uint64(), currentRound.Uint64(), symbols, votePeriod.Uint64(), nil
}

func (os *Server) handleConnectivityError() {
	os.lostSync = true
}

func (os *Server) checkHealth() {
	if os.lostSync {
		err := os.sync()
		if err != nil && !errors.Is(err, types.ErrNoSymbolsObserved) {
			os.logger.Info("rebuilding WS connectivity with Autonity L1 node", "error", err)
			if metrics.Enabled {
				metrics.GetOrRegisterCounter(monitor.L1ConnectivityMetric, nil).Inc(1)
			}
			return
		}
		os.lostSync = false
		return
	}
}

func (os *Server) isBlockchainSynced() bool {
	// if the autonity node is on peer synchronization state, just skip the reporting.
	syncing, err := os.client.SyncProgress(context.Background())
	if err != nil {
		os.logger.Error("vote get SyncProgress", "error", err.Error())
		return false
	}

	if syncing != nil {
		os.logger.Warn("skip round event since the Autonity L1 node is doing block synchronization")
		return false
	}

	return true
}

func (os *Server) syncProtocolSymbols() error {
	// get latest symbols from oracle.
	var err error
	os.protocolSymbols, err = os.oracleContract.GetSymbols(nil)
	if err != nil {
		os.logger.Error("vote get symbols", "error", err.Error())
		return err
	}

	return nil
}

// resetSamplingSymbols reset the latest sampling symbol set with the protocol symbol set.
func (os *Server) resetSamplingSymbols(protocolSymbols []string) {
	os.samplingSymbols = protocolSymbols

	// check if we need to add bridger symbols on demand.
	bridged := false
	for _, s := range protocolSymbols {
		if bridger, ok := BridgedSymbols[s]; ok {
			bridged = true
			os.samplingSymbols = append(os.samplingSymbols, bridger)
		}
	}

	if bridged {
		os.samplingSymbols = append(os.samplingSymbols, USDCUSD)
	}
}

// addNewSymbols adds new symbols to the local symbol set for data fetching, duplicated one is not added.
func (os *Server) addNewSymbols(newSymbols []string) {
	var symbolsMap = make(map[string]struct{})
	for _, s := range os.samplingSymbols {
		symbolsMap[s] = struct{}{}
	}

	// check if we need to add bridger symbols on demand.
	bridged := false
	for _, newS := range newSymbols {
		if _, ok := symbolsMap[newS]; !ok {
			os.samplingSymbols = append(os.samplingSymbols, newS)
			// if the new symbol requires a bridger symbol, add it too.
			if bridger, ok := BridgedSymbols[newS]; ok {
				bridged = true
				if _, ok := symbolsMap[bridger]; !ok {
					os.samplingSymbols = append(os.samplingSymbols, bridger)
				}
			}
		}
	}

	if _, ok := symbolsMap[USDCUSD]; !ok && bridged {
		os.samplingSymbols = append(os.samplingSymbols, USDCUSD)
	}
}

func (os *Server) handleNewSymbolsEvent(symbols []string) {
	// just add symbols to oracle service's symbol pool, thus the oracle service can start to prepare the data.
	os.addNewSymbols(symbols)
}
