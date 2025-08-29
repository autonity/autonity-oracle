package server

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/monitor"
	"autonity-oracle/types"
	"context"
	"math"
	"math/big"
	o "os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	"github.com/shopspring/decimal"
)

var BridgedSymbols = map[string]string{
	ATNUSD: ATNUSDC,
	NTNUSD: NTNUSDC,
}

var (
	saltRange       = new(big.Int).SetUint64(math.MaxInt64)
	alertBalance    = new(big.Int).SetUint64(2000000000000) // 2000 Gwei, 0.000002 Ether
	invalidPrice    = big.NewInt(0)
	invalidSalt     = big.NewInt(0)
	tenSecsInterval = 10 * time.Second // ticker to gc round data.
	oneSecsInterval = 1 * time.Second  // ticker for pre-sampling interval and for L1 reconnecting.
)

const (
	FirstRound          = uint64(1)
	ATNUSD              = "ATN-USD"
	NTNUSD              = "NTN-USD"
	USDCUSD             = "USDC-USD"
	ATNUSDC             = "ATN-USDC"
	NTNUSDC             = "NTN-USDC"
	MaxConfidence       = 100
	BaseConfidence      = 40
	OracleDecimals      = uint8(18)
	MaxBufferedRounds   = 10
	SourceScalingFactor = uint64(10)
	penalizeEventName   = "Penalized"
)

// Server coordinates the plugin discovery, the data sampling, and do the health checking with L1 connectivity.
type Server struct {
	logger hclog.Logger
	conf   *config.Config

	doneCh        chan struct{}
	regularTicker *time.Ticker // the clock source to trigger the 10s interval job.
	psTicker      *time.Ticker // the pre-sampling ticker in 1s.

	samplingSymbols []string // the symbols for data fetching in oracle service, can be different from the required protocol symbols.

	// the reporting staffs
	dialer         types.Dialer
	oracleContract contract.ContractAPI
	client         types.Blockchain
	abi            abi.ABI

	curRound       uint64 //round ID.
	votePeriod     uint64 //vote period.
	curSampleTS    int64  //the data sample TS of the current round.
	curRoundHeight uint64 //The block height on which the last round rotation happens.

	protocolSymbols []string //symbols required for the voting on the oracle contract protocol.
	pricePrecision  decimal.Decimal

	voteRecords VoteRecords

	chInvalidVote  chan *contract.OracleInvalidVote
	subInvalidVote event.Subscription

	chVotedEvent  chan *contract.OracleSuccessfulVote
	subVotedEvent event.Subscription

	chRewardEvent  chan *contract.OracleTotalOracleRewards
	subRewardEvent event.Subscription

	chNoRevealEvent  chan *contract.OracleNoRevealPenalty
	subNoRevealEvent event.Subscription

	chPenalizedEvent  chan *contract.OraclePenalized
	subPenalizedEvent event.Subscription

	chRoundEvent  chan *contract.OracleNewRound
	subRoundEvent event.Subscription

	chSymbolsEvent  chan *contract.OracleNewSymbols
	subSymbolsEvent event.Subscription
	lastSampledTS   int64

	sampleEventFeed        event.Feed
	lostSync               bool // set to true if the connectivity with L1 Autonity network is dropped during runtime.
	commitmentHashComputer *CommitmentHashComputer

	memories      Memories
	pluginManager *PluginManager
}

func NewServer(conf *config.Config, dialer types.Dialer, client types.Blockchain,
	oc contract.ContractAPI) *Server {
	os := &Server{
		conf:           conf,
		dialer:         dialer,
		client:         client,
		oracleContract: oc,
		voteRecords:    make(map[uint64]*types.VoteRecord),
		doneCh:         make(chan struct{}),
		regularTicker:  time.NewTicker(tenSecsInterval),
		psTicker:       time.NewTicker(oneSecsInterval),
		pricePrecision: decimal.NewFromBigInt(common.Big1, int32(OracleDecimals)),
	}

	os.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(os).String() + conf.Key.Address.String(),
		Output: o.Stdout,
		Level:  conf.LoggingLevel,
	})

	abi, err := abi.JSON(strings.NewReader(contract.OracleMetaData.ABI))
	if err != nil {
		os.logger.Error("failed to load ABI", "error", err)
		o.Exit(1)
	}
	os.abi = abi

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		os.logger.Error("failed to get chain id", "error", err)
		o.Exit(1)
	}

	commitmentHashComputer, err := NewCommitmentHashComputer()
	if err != nil {
		os.logger.Error("cannot create commitment hash computer", "err", err)
		o.Exit(1)
	}
	os.commitmentHashComputer = commitmentHashComputer

	// load memories from persistence.
	os.memories = Memories{dataDir: conf.ProfileDir}
	os.memories.init(os.logger)
	if os.memories.voteRecords != nil {
		os.voteRecords = *os.memories.voteRecords
		os.logger.Info("loaded vote records from persistence", "records", len(os.voteRecords))
	}

	os.logger.Info("running oracle contract listener at", "WS", conf.AutonityWSUrl, "ID", conf.Key.Address.String())
	err = os.sync()
	if err != nil {
		// stop the client on start up once the remote endpoint of autonity L1 network is not ready.
		os.logger.Error("Cannot synchronize oracle contract state from Autonity L1 Network", "error", err.Error(), "WS", conf.AutonityWSUrl)
		o.Exit(1)
	}
	os.lostSync = false

	os.pluginManager = NewPluginManager(os.conf.ConfigFile, os.conf.PluginDir, os.conf.LoggingLevel, os, chainID.Int64(),
		os.conf.Key.Address, os.conf.PluginConfigs)
	return os
}

func (os *Server) Start() {
	go os.pluginManager.Start()

	for {
		select {
		case <-os.doneCh:
			os.regularTicker.Stop()
			os.psTicker.Stop()
			os.logger.Info("server is stopping ...")
			return
		case err := <-os.subSymbolsEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new symbols event", err)
				os.handleConnectivityError()
				os.subSymbolsEvent.Unsubscribe()
			}
		case err := <-os.subRoundEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new round event", err)
				os.handleConnectivityError()
				os.subRoundEvent.Unsubscribe()
			}
		case err := <-os.subRewardEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of reward event", err)
				os.handleConnectivityError()
				os.subRewardEvent.Unsubscribe()
			}
		case err := <-os.subInvalidVote.Err():
			if err != nil {
				os.logger.Info("subscription error of invalid vote", err)
				os.handleConnectivityError()
				os.subInvalidVote.Unsubscribe()
			}
		case err := <-os.subVotedEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of voted event", err)
				os.handleConnectivityError()
				os.subVotedEvent.Unsubscribe()
			}
		case err := <-os.subNoRevealEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of no reveal event", err)
				os.handleConnectivityError()
				os.subNoRevealEvent.Unsubscribe()
			}
		case err := <-os.subPenalizedEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of penalty event", err)
				os.handleConnectivityError()
				os.subPenalizedEvent.Unsubscribe()
			}
		case <-os.psTicker.C:
			// shorten the health checker, if we have L1 connectivity issue, try to repair it before pre sampling starts.
			os.checkHealth()
			preSampleTS := time.Now().Unix()
			err := os.handlePreSampling(preSampleTS)
			if err != nil {
				os.logger.Error("handle pre-sampling", "error", err.Error())
			}
			os.lastSampledTS = preSampleTS

		case invalidVote := <-os.chInvalidVote:
			os.logger.Info("received invalid vote", "cause", invalidVote.Cause, "expected",
				invalidVote.ExpValue.String(), "actual", invalidVote.ActualValue.String(), "txn", invalidVote.Raw.TxHash)
			os.setVoteMined(invalidVote.Raw.TxHash, invalidVote.Cause)
			if metrics.Enabled {
				metrics.GetOrRegisterCounter(monitor.InvalidVoteMetric, nil).Inc(1)
			}

		case votedEvent := <-os.chVotedEvent:
			os.logger.Info("received voted event", "height", votedEvent.Raw.BlockNumber, "txn", votedEvent.Raw.TxHash)
			os.setVoteMined(votedEvent.Raw.TxHash, "")
			if metrics.Enabled {
				metrics.GetOrRegisterCounter(monitor.SuccessfulVoteMetric, nil).Inc(1)
			}

		case rewardEvent := <-os.chRewardEvent:
			os.logger.Info("received reward distribution event", "height", rewardEvent.Raw.BlockNumber,
				"total distributed ATN", rewardEvent.AtnReward.Uint64(), "total distributed NTN", rewardEvent.NtnReward.Uint64())

		case noRevealEvent := <-os.chNoRevealEvent:
			os.logger.Info("received no reveal event", "height", noRevealEvent.Raw.BlockNumber, "round", noRevealEvent.Round,
				"missed", noRevealEvent.MissedReveal.Uint64())
			if metrics.Enabled {
				metrics.GetOrRegisterGauge(monitor.NoRevealVoteMetric, nil).Update(noRevealEvent.MissedReveal.Int64())
			}

		case penalizeEvent := <-os.chPenalizedEvent:
			if err := os.handlePenaltyEvent(penalizeEvent); err != nil {
				os.logger.Error("handle penalty event", "error", err.Error())
			}

		case roundEvent := <-os.chRoundEvent:
			os.logger.Info("handle new round", "round", roundEvent.Round.Uint64(), "required sampling TS",
				roundEvent.Timestamp.Uint64(), "height", roundEvent.Raw.BlockNumber, "round period", roundEvent.VotePeriod.Uint64())

			if metrics.Enabled {
				metrics.GetOrRegisterGauge(monitor.RoundMetric, nil).Update(roundEvent.Round.Int64())
			}

			// IMPORTANT! sync the round and vote period carries by the round event, and store the reference points for
			// next presampling and the target for sample selection.
			os.curRound = roundEvent.Round.Uint64()
			os.votePeriod = roundEvent.VotePeriod.Uint64()
			os.curRoundHeight = roundEvent.Raw.BlockNumber
			os.curSampleTS = roundEvent.Timestamp.Int64()

			// vote for latest protocol symbols.
			err := os.vote()
			if err != nil {
				os.logger.Error("round voting failed", "error", err.Error())
			}
			// after vote, reset sampling symbols with the latest protocol symbols.
			os.resetSamplingSymbols(os.protocolSymbols)
			os.printLatestRoundData(os.curRound)
		case newSymbolEvent := <-os.chSymbolsEvent:
			// New symbols are added, add them into the sampling set to prepare data in advance for the coming round's vote.
			os.logger.Info("handle new symbols", "new symbols", newSymbolEvent.Symbols, "activate at round", newSymbolEvent.Round)
			os.handleNewSymbolsEvent(newSymbolEvent.Symbols)
		case <-os.regularTicker.C:
			os.trackVoteState()
			os.gcVoteRecords()
			os.logger.Debug("current round ID", "round", os.curRound)
		}
	}
}

func (os *Server) Stop() {
	os.client.Close()
	os.subRoundEvent.Unsubscribe()
	os.subSymbolsEvent.Unsubscribe()
	os.subPenalizedEvent.Unsubscribe()
	os.subNoRevealEvent.Unsubscribe()
	os.subVotedEvent.Unsubscribe()
	os.subInvalidVote.Unsubscribe()
	os.subRewardEvent.Unsubscribe()

	os.doneCh <- struct{}{}
	os.pluginManager.Stop()
	os.logger.Info("server is stopped")
}

func (os *Server) printLatestRoundData(newRound uint64) {
	for _, s := range os.protocolSymbols {
		rd, err := os.oracleContract.GetRoundData(nil, new(big.Int).SetUint64(newRound-1), s)
		if err != nil {
			os.logger.Error("get round data", "error", err.Error())
			return
		}

		os.logger.Debug("get round price", "round", newRound-1, "symbol", s, "Price",
			rd.Price.String(), "success", rd.Success)
	}

	for _, s := range os.protocolSymbols {
		rd, err := os.oracleContract.LatestRoundData(nil, s)
		if err != nil {
			os.logger.Error("get latest round price", "error", err.Error())
			return
		}

		price, err := decimal.NewFromString(rd.Price.String())
		if err != nil {
			continue
		}

		os.logger.Debug("latest round price", "round", rd.Round.Uint64(), "symbol", s, "price",
			price.Div(os.pricePrecision).String(), "success", rd.Success)
	}
}
