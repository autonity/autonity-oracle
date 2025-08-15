package oracleserver

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/helpers"
	"autonity-oracle/monitor"
	pWrapper "autonity-oracle/plugin_wrapper"
	common2 "autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/fs"
	"math"
	"math/big"
	o "os"
	"path/filepath"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	tp "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/fsnotify/fsnotify"
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
	tenSecsInterval = 10 * time.Second // ticker to check L2 connectivity and gc round data.
	oneSecsInterval = 1 * time.Second  // sampling interval during data pre-sampling period.
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
)

// OracleServer coordinates the plugin discovery, the data sampling, and do the health checking with L1 connectivity.
type OracleServer struct {
	logger hclog.Logger
	conf   *config.Config

	doneCh        chan struct{}
	regularTicker *time.Ticker // the clock source to trigger the 10s interval job.
	psTicker      *time.Ticker // the pre-sampling ticker in 1s.

	runningPlugins  map[string]*pWrapper.PluginWrapper // the plugin clients that connect with different adapters.
	samplingSymbols []string                           // the symbols for data fetching in oracle service, can be different from the required protocol symbols.

	keyRequiredPlugins map[string]struct{} // saving those plugins which require a key granted by data provider

	// the reporting staffs
	dialer         types.Dialer
	oracleContract contract.ContractAPI
	client         types.Blockchain

	curRound        uint64 //round ID.
	votePeriod      uint64 //vote period.
	curSampleTS     int64  //the data sample TS of the current round.
	curSampleHeight uint64 //The block height on which the last round rotation happens.

	protocolSymbols []string //symbols required for the voting on the oracle contract protocol.
	pricePrecision  decimal.Decimal
	roundData       map[uint64]*types.RoundData

	chInvalidVote  chan *contract.OracleInvalidVote
	subInvalidVote event.Subscription

	chVotedEvent  chan *contract.OracleSuccessfulVote
	subVotedEvent event.Subscription

	chRewardEvent  chan *contract.OracleTotalOracleRewards
	subRewardEvent event.Subscription

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

	memories Memories

	configWatcher  *fsnotify.Watcher // config file watcher which watches the config changes.
	pluginsWatcher *fsnotify.Watcher // plugins watcher which watches the changes of plugins and the plugins' configs.
	chainID        int64             // ChainID saves the L1 chain ID, it is used for plugin compatibility check.
}

func NewOracleServer(conf *config.Config, dialer types.Dialer, client types.Blockchain,
	oc contract.ContractAPI) *OracleServer {
	os := &OracleServer{
		conf:               conf,
		dialer:             dialer,
		client:             client,
		oracleContract:     oc,
		roundData:          make(map[uint64]*types.RoundData),
		runningPlugins:     make(map[string]*pWrapper.PluginWrapper),
		keyRequiredPlugins: make(map[string]struct{}),
		doneCh:             make(chan struct{}),
		regularTicker:      time.NewTicker(tenSecsInterval),
		psTicker:           time.NewTicker(oneSecsInterval),
		pricePrecision:     decimal.NewFromBigInt(common.Big1, int32(OracleDecimals)),
	}

	os.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(os).String() + conf.Key.Address.String(),
		Output: o.Stdout,
		Level:  conf.LoggingLevel,
	})

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		os.logger.Error("failed to get chain id", "error", err)
		o.Exit(1)
	}

	os.chainID = chainID.Int64()

	commitmentHashComputer, err := NewCommitmentHashComputer()
	if err != nil {
		os.logger.Error("cannot create commitment hash computer", "err", err)
		o.Exit(1)
	}
	os.commitmentHashComputer = commitmentHashComputer

	// load historic state, otherwise default initial state will be used.
	os.memories = Memories{dataDir: conf.ProfileDir}
	os.memories.init(os.logger)

	// discover plugins from plugin dir at startup.
	binaries, err := helpers.ListPlugins(conf.PluginDIR)
	if len(binaries) == 0 || err != nil {
		// to stop the service on the start once there is no plugin in the db.
		os.logger.Error("no plugin discovered", "plugin-dir", os.conf.PluginDIR)
		o.Exit(1)
	}

	// take the custom plugin configs and start plugins.
	for _, file := range binaries {
		f := file
		pConf := conf.PluginConfigs[f.Name()]
		if pConf.Disabled {
			continue
		}
		os.tryToLaunchPlugin(f, pConf)
	}

	os.logger.Info("running oracle contract listener at", "WS", conf.AutonityWSUrl, "ID", conf.Key.Address.String())
	err = os.syncStates()
	if err != nil {
		// stop the client on start up once the remote endpoint of autonity L1 network is not ready.
		os.logger.Error("Cannot synchronize oracle contract state from Autonity L1 Network", "error", err.Error(), "WS", conf.AutonityWSUrl)
		o.Exit(1)
	}
	os.lostSync = false

	// subscribe FS notifications of the watched plugins.
	pluginsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		os.logger.Error("cannot create fsnotify watcher", "error", err)
		o.Exit(1)
	}

	err = pluginsWatcher.Add(conf.PluginDIR)
	if err != nil {
		os.logger.Error("cannot watch plugin dir", "error", err)
		o.Exit(1)
	}
	os.pluginsWatcher = pluginsWatcher

	// subscribe FS notification of the watched config file.
	configWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		os.logger.Error("cannot create fsnotify watcher", "error", err)
		o.Exit(1)
	}

	dir := filepath.Dir(conf.ConfigFile)
	err = configWatcher.Add(dir) // Watch parent directory
	if err != nil {
		os.logger.Error("cannot watch oracle config directory", "error", err)
		o.Exit(1)
	}

	os.configWatcher = configWatcher
	return os
}

// syncStates is executed on client startup or after the L1 connection recovery to sync the on-chain oracle contract
// states, symbols, round id, precision, vote period, etc... to the oracle server. It also subscribes the on-chain
// events of oracle protocol: round event, symbol update event, etc...
func (os *OracleServer) syncStates() error {
	var err error
	// get initial states from on-chain oracle contract.
	os.curRound, os.protocolSymbols, os.votePeriod, err = os.initStates()
	if err != nil {
		os.logger.Error("synchronize oracle contract state", "error", err.Error())
		return err
	}

	// reset sampling symbols with the latest protocol symbols, it adds bridger symbols by according to the protocol symbols.
	os.ResetSamplingSymbols(os.protocolSymbols)
	os.logger.Info("syncStates", "CurrentRound", os.curRound, "protocol symbols", os.protocolSymbols, "sampling symbols", os.samplingSymbols)

	// subscribe on-chain round rotation event
	chRoundEvent := make(chan *contract.OracleNewRound)
	subRoundEvent, err := os.oracleContract.WatchNewRound(new(bind.WatchOpts), chRoundEvent)
	if err != nil {
		os.logger.Error("failed to subscribe round event", "error", err.Error())
		return err
	}
	os.chRoundEvent = chRoundEvent
	os.subRoundEvent = subRoundEvent

	// subscribe on-chain symbol update event
	chSymbolsEvent := make(chan *contract.OracleNewSymbols)
	subSymbolsEvent, err := os.oracleContract.WatchNewSymbols(new(bind.WatchOpts), chSymbolsEvent)
	if err != nil {
		os.logger.Error("failed to subscribe new symbol event", "error", err.Error())
		return err
	}
	os.chSymbolsEvent = chSymbolsEvent
	os.subSymbolsEvent = subSymbolsEvent

	// subscribe on-chain penalize event
	chPenalizedEvent := make(chan *contract.OraclePenalized)
	subPenalizedEvent, err := os.oracleContract.WatchPenalized(new(bind.WatchOpts), chPenalizedEvent, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe penalized event", "error", err.Error())
		return err
	}

	os.chPenalizedEvent = chPenalizedEvent
	os.subPenalizedEvent = subPenalizedEvent

	// subscribe voted event
	chVotedEvent := make(chan *contract.OracleSuccessfulVote)
	subVotedEvent, err := os.oracleContract.WatchSuccessfulVote(new(bind.WatchOpts), chVotedEvent, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe voted event", "error", err.Error())
		return err
	}

	os.chVotedEvent = chVotedEvent
	os.subVotedEvent = subVotedEvent

	// subscribe invalid vote event
	chInvalidVote := make(chan *contract.OracleInvalidVote)
	subInvalidVote, err := os.oracleContract.WatchInvalidVote(new(bind.WatchOpts), chInvalidVote, []common.Address{os.conf.Key.Address})
	if err != nil {
		os.logger.Error("failed to subscribe invalid vote event", "error", err.Error())
		return err
	}
	os.chInvalidVote = chInvalidVote
	os.subInvalidVote = subInvalidVote

	// subscribe reward event
	chRewardEvent := make(chan *contract.OracleTotalOracleRewards)
	subRewardEvent, err := os.oracleContract.WatchTotalOracleRewards(new(bind.WatchOpts), chRewardEvent)
	if err != nil {
		os.logger.Error("failed to subscribe reward event", "error", err.Error())
		return err
	}

	os.chRewardEvent = chRewardEvent
	os.subRewardEvent = subRewardEvent

	return nil
}

// initStates returns round id, symbols and committees on current chain, it is called on the startup of client.
func (os *OracleServer) initStates() (uint64, []string, uint64, error) {
	// on the startup, we need to sync the round id, symbols and committees from contract.
	currentRound, err := os.oracleContract.GetRound(nil)
	if err != nil {
		os.logger.Error("get round", "error", err.Error())
		return 0, nil, 0, err
	}

	symbols, err := os.oracleContract.GetSymbols(nil)
	if err != nil {
		os.logger.Error("get symbols", "error", err.Error())
		return 0, nil, 0, err
	}

	votePeriod, err := os.oracleContract.GetVotePeriod(nil)
	if err != nil {
		os.logger.Error("get vote period", "error", err.Error())
		return 0, nil, 0, nil
	}

	if len(symbols) == 0 {
		os.logger.Error("there are no symbols in Autonity L1 oracle contract")
		return currentRound.Uint64(), symbols, votePeriod.Uint64(), types.ErrNoSymbolsObserved
	}

	return currentRound.Uint64(), symbols, votePeriod.Uint64(), nil
}

func (os *OracleServer) gcExpiredSamples() {
	for _, plugin := range os.runningPlugins {
		plugin.GCExpiredSamples()
	}
}

func (os *OracleServer) gcRoundData() {
	if len(os.roundData) >= MaxBufferedRounds {
		offset := os.curRound - MaxBufferedRounds
		for k := range os.roundData {
			if k <= offset {
				delete(os.roundData, k)
			}
		}
	}
}

func (os *OracleServer) handleConnectivityError() {
	os.lostSync = true
}

func (os *OracleServer) checkHealth() {
	if os.lostSync {
		err := os.syncStates()
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

func (os *OracleServer) isVoter() (bool, error) {
	voters, err := os.oracleContract.GetVoters(nil)
	if err != nil {
		os.logger.Error("get voters", "error", err.Error())
		return false, err
	}

	for _, c := range voters {
		if c == os.conf.Key.Address {
			return true, nil
		}
	}
	return false, nil
}

func (os *OracleServer) printLatestRoundData(newRound uint64) {
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

func (os *OracleServer) handlePreSampling(preSampleTS int64) error {

	// start to sample data point for the 1st round as the round period could be longer than 30s, we don't want to
	// wait for another round to get the data be available on-chain.
	if os.curRound == FirstRound {
		nextRoundHeight := os.votePeriod
		curHeight, err := os.client.BlockNumber(context.Background())
		if err != nil {
			os.logger.Error("handle pre-sampling", "error", err.Error())
			return err
		}

		if curHeight > nextRoundHeight {
			return nil
		}

		if nextRoundHeight-curHeight > uint64(config.PreSamplingRange) { //nolint
			return nil
		}

		// do the data pre-sampling.
		os.logger.Debug("data pre-sampling", "round", os.curRound, "on height", curHeight, "TS", preSampleTS)
		os.samplePrice(os.samplingSymbols, preSampleTS)

		return nil
	}

	// taking the 1st round and the round after a node recover from a disaster as a special case, to skip the
	// pre-sampling. In this special case, the regular 10s samples will be used for data reporting.
	if os.curSampleTS == 0 || os.curSampleHeight == 0 {
		return nil
	}

	// if it is not a good timing to start sampling then return.
	nextRoundHeight := os.curSampleHeight + os.votePeriod
	curHeight, err := os.client.BlockNumber(context.Background())
	if err != nil {
		os.logger.Error("handle pre-sampling", "error", err.Error())
		return err
	}
	if nextRoundHeight-curHeight > uint64(config.PreSamplingRange) { //nolint
		return nil
	}

	// do the data pre-sampling.
	os.logger.Debug("data pre-sampling", "round", os.curRound, "on height", curHeight, "TS", preSampleTS)
	os.samplePrice(os.samplingSymbols, preSampleTS)

	return nil
}

func (os *OracleServer) handleRoundVote() error {
	// if the autonity node is on peer synchronization state, just skip the reporting.
	syncing, err := os.client.SyncProgress(context.Background())
	if err != nil {
		os.logger.Error("handleRoundVote get SyncProgress", "error", err.Error())
		return err
	}

	if syncing != nil {
		os.logger.Warn("skip round event since the Autonity L1 node is doing block synchronization")
		return types.ErrPeerOnSync
	}

	// get latest symbols from oracle.
	os.protocolSymbols, err = os.oracleContract.GetSymbols(nil)
	if err != nil {
		os.logger.Error("handleRoundVote get symbols", "error", err.Error())
		return err
	}

	os.printLatestRoundData(os.curRound)

	// if client is not a voter, just skip reporting.
	isVoter, err := os.isVoter()
	if err != nil {
		os.logger.Error("handleRoundVote isVoter", "error", err.Error())
		return err
	}

	// query last round's prices, its random salt which will reveal last round's report.
	lastRoundData, ok := os.roundData[os.curRound-1]
	if !ok {
		// no last round data was addressed from the buffer, then try to find it from the persistence memory.
		if os.memories.voteRecord != nil && os.memories.voteRecord.RoundID == os.curRound-1 {
			os.logger.Info("load last round data from server persistence")
			lastRoundData = os.memories.voteRecord.ToRoundData()
		} else {
			os.logger.Debug("no last round data, client is no longer a voter or it reports just current round commitment")
		}
	}

	// if node is no longer a validator, and it doesn't have last round data, skip reporting.
	if !isVoter && !ok {
		os.logger.Debug("skip data reporting since client is no longer a voter")
		if metrics.Enabled {
			metrics.GetOrRegisterGauge(monitor.IsVoterMetric, nil).Update(0)
		}
		return nil
	}

	// check with the vote buffer from the last penalty event.
	if os.memories.outlierRecord != nil && os.curSampleHeight-os.memories.outlierRecord.LastPenalizedAtBlock <= os.conf.VoteBuffer {
		left := os.conf.VoteBuffer - (os.curSampleHeight - os.memories.outlierRecord.LastPenalizedAtBlock)
		os.logger.Warn("due to the outlier penalty, we postpone your next vote from slashing", "next vote block", left)
		os.logger.Warn("your last outlier report was", "report", os.memories.outlierRecord)
		os.logger.Warn("during this period, you can: 1. check your data source infra; 2. restart your oracle-client; 3. contact Autonity team for help;")
		return nil
	}

	if isVoter {
		// a voter need to assemble current round data to report it.
		curRoundData, err := os.buildRoundData(os.curRound)
		// if the voter failed to assemble current round data, but it has last round data available, then reveal it.
		if err != nil {
			if lastRoundData != nil {
				return os.reportWithoutCommitment(lastRoundData)
			}
			return err
		}

		// round data was successfully assembled, save current round data.
		if err = os.memories.flushRecord(toVoteRecord(curRoundData)); err != nil {
			os.logger.Warn("failed to vote record to persistence", "error", err.Error())
			os.logger.Warn("IMPORTANT: please check your profile data dir, the server need to flush vote record into it.")
		}

		os.roundData[os.curRound] = curRoundData
		if metrics.Enabled {
			metrics.GetOrRegisterGauge(monitor.IsVoterMetric, nil).Update(1)
		}
		// report with last round data and with current round commitment hash.
		return os.reportWithCommitment(curRoundData, lastRoundData)
	}

	// edge case, voter is no longer a committee member, it has to reveal the last round data that it committed to.
	if lastRoundData != nil {
		return os.reportWithoutCommitment(lastRoundData)
	}
	return nil
}

// reportWithCommitment reports the commitment of current round, and with last round data if the last round data is available.
// if the input last round data is nil, we just need to report the commitment of current round without last round data.
func (os *OracleServer) reportWithCommitment(curRoundData, lastRoundData *types.RoundData) error {
	var err error
	// prepare the transaction which carry current round's commitment, and last round's data.
	curRoundData.Tx, err = os.doReport(curRoundData.CommitmentHash, lastRoundData)
	if err != nil {
		os.logger.Error("do report", "error", err.Error())
		return err
	}

	os.logger.Info("reported last round data and with current round commitment", "TX hash", curRoundData.Tx.Hash(), "Nonce", curRoundData.Tx.Nonce(), "Cost", curRoundData.Tx.Cost())

	// alert in case of balance reach the warning value.
	balance, err := os.client.BalanceAt(context.Background(), os.conf.Key.Address, nil)
	if err != nil {
		os.logger.Error("cannot get account balance", "error", err.Error())
		return err
	}

	if metrics.Enabled {
		metrics.GetOrRegisterGauge(monitor.BalanceMetric, nil).Update(balance.Int64())
	}

	os.logger.Info("oracle server account", "address", os.conf.Key.Address, "remaining balance", balance.String())
	if balance.Cmp(alertBalance) <= 0 {
		os.logger.Warn("oracle account has too less balance left for data reporting", "balance", balance.String())
	}

	return nil
}

// report with last round data but without current round commitment, voter is leaving from the committee.
func (os *OracleServer) reportWithoutCommitment(lastRoundData *types.RoundData) error {

	// report with no commitment of current round as voter is leaving from the committee.
	tx, err := os.doReport(common.Hash{}, lastRoundData)
	if err != nil {
		os.logger.Error("do report", "error", err.Error())
		return err
	}
	os.logger.Info("reported last round data and without current round commitment", "TX hash", tx.Hash(), "Nonce", tx.Nonce())
	return nil
}

// ResetSamplingSymbols reset the latest sampling symbol set with the protocol symbol set.
func (os *OracleServer) ResetSamplingSymbols(protocolSymbols []string) {
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

// AddNewSymbols adds new symbols to the local symbol set for data fetching, duplicated one is not added.
func (os *OracleServer) AddNewSymbols(newSymbols []string) {
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

func (os *OracleServer) resolveGasTipCap() *big.Int {
	configured := new(big.Int).SetUint64(os.conf.GasTipCap)
	suggested, err := os.client.SuggestGasTipCap(context.Background())
	if err != nil {
		os.logger.Warn("cannot get fee history, using configured gas tip cap", "error", err.Error())
		return configured
	}

	// take the max one to let the report get mine with higher priority.
	if suggested.Cmp(configured) > 0 {
		return suggested
	}
	return configured
}

func (os *OracleServer) doReport(curRoundCommitmentHash common.Hash, lastRoundData *types.RoundData) (*tp.Transaction, error) {
	chainID, err := os.client.ChainID(context.Background())
	if err != nil {
		os.logger.Error("get chain id", "error", err.Error())
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(os.conf.Key.PrivateKey, chainID)
	if err != nil {
		os.logger.Error("new keyed transactor with chain ID", "error", err)
		return nil, err
	}

	gasTipCap := os.resolveGasTipCap()
	// Get base fee for pending block
	header, err := os.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		os.logger.Error("get header by number", "error", err.Error())
		return nil, err
	}

	// Calculate MaxFeePerGas (baseFee * 2 + gasTipCap)
	maxFeePerGas := new(big.Int).Mul(header.BaseFee, big.NewInt(2))
	maxFeePerGas.Add(maxFeePerGas, gasTipCap)

	auth.Value = big.NewInt(0)
	auth.GasTipCap = gasTipCap
	auth.GasFeeCap = maxFeePerGas
	auth.GasLimit = uint64(3000000)

	// if there is no last round data, then we just submit the commitment hash of current round.
	if lastRoundData == nil {
		var reports []contract.IOracleReport
		return os.oracleContract.Vote(auth, new(big.Int).SetBytes(curRoundCommitmentHash.Bytes()), reports, invalidSalt, config.Version)
	}

	// there is last round data, report with current round commitment, and the last round reports and salt to be revealed.
	return os.oracleContract.Vote(auth, new(big.Int).SetBytes(curRoundCommitmentHash.Bytes()), lastRoundData.Reports, lastRoundData.Salt, config.Version)
}

func (os *OracleServer) buildRoundData(round uint64) (*types.RoundData, error) {
	if len(os.protocolSymbols) == 0 {
		return nil, types.ErrNoSymbolsObserved
	}

	prices, err := os.aggregateProtocolSymbolPrices()
	if err != nil {
		return nil, err
	}

	// assemble round data with reports, salt and commitment hash.
	roundData, err := os.assembleReportData(round, os.protocolSymbols, prices)
	if err != nil {
		os.logger.Error("failed to assemble round report data", "error", err.Error())
		return nil, err
	}
	os.logger.Info("assembled round report data", "current round", round, "prices", roundData)
	return roundData, nil
}

func (os *OracleServer) aggregateProtocolSymbolPrices() (types.PriceBySymbol, error) {
	prices := make(types.PriceBySymbol)

	// if we need a bridger pair USDC-USD to convert ATN-USD or NTN-USD from ATN-USDC or NTN-USDC,
	// then we need to aggregate USDC-USD data point first.
	var usdcPrice *types.Price
	var err error
	if slices.Contains(os.protocolSymbols, ATNUSD) || slices.Contains(os.protocolSymbols, NTNUSD) {
		usdcPrice, err = os.aggregatePrice(USDCUSD, os.curSampleTS)
		if err != nil {
			os.logger.Error("aggregate USDC-USD price", "error", err.Error())
		}
	}

	for _, s := range os.protocolSymbols {
		// aggregate bridged symbols
		if s == ATNUSD || s == NTNUSD {
			if usdcPrice == nil {
				continue
			}

			p, e := os.aggregateBridgedPrice(s, os.curSampleTS, usdcPrice)
			if e != nil {
				os.logger.Error("aggregate bridged price", "error", e.Error(), "symbol", s)
				continue
			}
			prices[s] = *p
			continue
		}

		// aggregate none bridged symbols
		p, e := os.aggregatePrice(s, os.curSampleTS)
		if e != nil {
			os.logger.Debug("no data for aggregation", "reason", e.Error(), "symbol", s)
			continue
		}
		prices[s] = *p
	}

	// edge case: if NTN-ATN price was not computable from inside plugin,
	// try to compute it from NTNprice and ATNprice across from different plugins.
	if _, ok := prices[common2.NTNATNSymbol]; !ok {
		ntnPrice, ntnExist := prices[NTNUSD]
		atnPrice, atnExist := prices[ATNUSD]
		if ntnExist && atnExist {
			ntnATNPrice, err := common2.ComputeDerivedPrice(ntnPrice.Price.String(), atnPrice.Price.String()) //nolint
			if err == nil {
				p, err := decimal.NewFromString(ntnATNPrice.Price) // nolint
				if err == nil {
					prices[common2.NTNATNSymbol] = types.Price{
						Timestamp:  time.Now().Unix(),
						Price:      p,
						Symbol:     common2.NTNATNSymbol,
						Confidence: ntnPrice.Confidence,
					}
				} else {
					os.logger.Error("cannot parse NTN-ATN price in decimal", "error", err.Error())
				}
			} else {
				os.logger.Error("failed to compute NTN-ATN price", "error", err.Error())
			}
		}
	}

	return prices, nil
}

// assemble the final reports, salt and commitment hash.
func (os *OracleServer) assembleReportData(round uint64, symbols []string, prices types.PriceBySymbol) (*types.RoundData, error) {
	var roundData = &types.RoundData{
		RoundID: round,
		Symbols: symbols,
		Prices:  prices,
	}

	var missingData bool
	var reports []contract.IOracleReport
	for _, s := range symbols {
		if pr, ok := prices[s]; ok {
			// This is an edge case, which means there is no liquidity in the market for this symbol.
			price := pr.Price.Mul(os.pricePrecision).BigInt()
			if price.Cmp(invalidPrice) == 0 {
				os.logger.Info("please check your data source, zero data point measured from market", "symbol", s)
				missingData = true
			}
			reports = append(reports, contract.IOracleReport{
				Price:      price,
				Confidence: pr.Confidence,
			})
		} else {
			// logging the missing of data points for symbols
			missingData = true
			os.logger.Info("please check your data source, missing data point for symbol", "symbol", s)
		}
	}

	// we won't assemble the round data if any data point is missing.
	if missingData {
		return nil, types.ErrMissingDataPoint
	}

	salt, err := rand.Int(rand.Reader, saltRange)
	if err != nil {
		os.logger.Error("generate rand salt", "error", err.Error())
		return nil, err
	}

	commitmentHash, err := os.commitmentHashComputer.CommitmentHash(reports, salt, os.conf.Key.Address)
	if err != nil {
		os.logger.Error("failed to compute commitment hash", "error", err.Error())
		return nil, err
	}

	roundData.Reports = reports
	roundData.Salt = salt
	roundData.CommitmentHash = commitmentHash
	return roundData, nil
}

func (os *OracleServer) handleNewSymbolsEvent(symbols []string) {
	// just add symbols to oracle service's symbol pool, thus the oracle service can start to prepare the data.
	os.AddNewSymbols(symbols)
}

// aggregateBridgedPrice ATN-USD or NTN-USD from bridged ATN-USDC or NTN-USDC with USDC-USD price,
// it assumes the input usdcPrice is not nil.
func (os *OracleServer) aggregateBridgedPrice(srcSymbol string, target int64, usdcPrice *types.Price) (*types.Price, error) {
	var bridgedSymbol string
	if srcSymbol == ATNUSD {
		bridgedSymbol = ATNUSDC
	}

	if srcSymbol == NTNUSD {
		bridgedSymbol = NTNUSDC
	}

	p, err := os.aggregatePrice(bridgedSymbol, target)
	if err != nil {
		os.logger.Error("aggregate bridged price", "error", err.Error(), "symbol", bridgedSymbol)
		return nil, err
	}

	// reset the symbol with source symbol,
	// and update price with: ATN-USD=ATN-USDC*USDC-USD / NTN-USD=NTN-USDC*USDC-USD
	// the confidence of ATN-USD and NTN-USD are inherit from ATN-USDC and NTN-USDC.
	p.Symbol = srcSymbol
	p.Price = p.Price.Mul(usdcPrice.Price)
	return p, nil
}

// aggregatePrice takes the symbol's aggregated data points from all the supported plugins, if there are multiple
// markets' datapoint, it will do a final VWAP aggregation to form the final reporting value.
func (os *OracleServer) aggregatePrice(s string, target int64) (*types.Price, error) {
	var prices []decimal.Decimal
	var volumes []*big.Int
	for _, plugin := range os.runningPlugins {
		p, err := plugin.AggregatedPrice(s, target)
		if err != nil {
			continue
		}
		prices = append(prices, p.Price)
		volumes = append(volumes, p.Volume)
	}

	if len(prices) == 0 {
		historicRoundPrice, err := os.queryHistoricRoundPrice(s)
		if err != nil {
			return nil, err
		}

		return confidenceAdjustedPrice(&historicRoundPrice, target)
	}

	// compute confidence of the symbol from the num of plugins' samples of it.
	confidence := ComputeConfidence(s, len(prices), os.conf.ConfidenceStrategy)
	price := &types.Price{
		Timestamp:  target,
		Price:      prices[0],
		Volume:     volumes[0],
		Symbol:     s,
		Confidence: confidence,
	}

	_, isForex := common2.ForexCurrencies[s]

	// we have multiple markets' data for this forex symbol, update the price with median value.
	if len(prices) > 1 && isForex {
		p, err := helpers.Median(prices)
		if err != nil {
			return nil, err
		}
		price.Price = p
		price.Volume = types.DefaultVolume
		return price, nil
	}

	// we have multiple markets' data for this crypto symbol, update the price with VWAP.
	if len(prices) > 1 && !isForex {
		p, vol, err := helpers.VWAP(prices, volumes)
		if err != nil {
			return nil, err
		}
		price.Price = p
		price.Volume = vol
	}

	return price, nil
}

// queryHistoricRoundPrice queries the last available price for a given symbol from the historic rounds.
func (os *OracleServer) queryHistoricRoundPrice(symbol string) (types.Price, error) {

	if len(os.roundData) == 0 {
		return types.Price{}, types.ErrNoDataRound
	}

	numOfRounds := len(os.roundData)
	// Iterate from the current round backward
	for i := 0; i < numOfRounds; i++ {
		roundID := os.curRound - uint64(i) - 1 //nolint
		// Get the round data for the current round ID
		roundData, exists := os.roundData[roundID]
		if !exists {
			continue
		}

		if roundData == nil {
			continue
		}

		// Check if the symbol exists in the Prices map
		if price, found := roundData.Prices[symbol]; found {
			return price, nil
		}
	}

	// If no price was found after checking all rounds, return an error
	return types.Price{}, types.ErrNoDataRound
}

func (os *OracleServer) samplePrice(symbols []string, ts int64) {
	if os.lastSampledTS == ts {
		return
	}
	cpSymbols := make([]string, len(symbols))
	copy(cpSymbols, symbols)
	e := &types.SampleEvent{
		Symbols: cpSymbols,
		TS:      ts,
	}
	nListener := os.sampleEventFeed.Send(e)
	os.logger.Debug("sample event is sent to", "num of plugins", nListener)
}

func (os *OracleServer) Start() {
	for {
		select {
		case <-os.doneCh:
			os.regularTicker.Stop()
			os.psTicker.Stop()
			if os.pluginsWatcher != nil {
				os.pluginsWatcher.Close() //nolint
			}

			if os.configWatcher != nil {
				os.configWatcher.Close() //nolint
			}
			os.logger.Info("oracle service is stopped")
			return
		case err := <-os.configWatcher.Errors:
			if err != nil {
				os.logger.Error("oracle config file watcher err", "err", err.Error())
			}
		case err := <-os.pluginsWatcher.Errors:
			if err != nil {
				os.logger.Error("plugin watcher errors", "err", err.Error())
			}
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
			if metrics.Enabled {
				metrics.GetOrRegisterCounter(monitor.InvalidVoteMetric, nil).Inc(1)
			}

		case votedEvent := <-os.chVotedEvent:
			os.logger.Info("received voted event", "height", votedEvent.Raw.BlockNumber, "txn", votedEvent.Raw.TxHash)
			if metrics.Enabled {
				metrics.GetOrRegisterCounter(monitor.SuccessfulVoteMetric, nil).Inc(1)
			}

		case rewardEvent := <-os.chRewardEvent:
			os.logger.Info("received reward distribution event", "height", rewardEvent.Raw.BlockNumber,
				"total distributed ATN", rewardEvent.AtnReward.Uint64(), "total distributed NTN", rewardEvent.NtnReward.Uint64())

		case penalizeEvent := <-os.chPenalizedEvent:

			// As the OutlierDetectionThreshold is set to low level, e.g. 3% against the median, and the OutlierSlashingThreshold
			// is configured at (10%, 15%) which is much higher, a penalization event may occur with zero slashing amount.
			// This indicates that the current client has been identified as an outlier but is not penalized, as its data
			// point falls below the OutlierSlashingThreshold when compared to the median price. To ensure a broader participation
			// of nodes within the oracle network and maintain its operational liveness, we continue to allow these
			// non-slashed outliers to contribute data samples to the network.

			if metrics.Enabled {
				gap := new(big.Int).Abs(new(big.Int).Sub(penalizeEvent.Reported, penalizeEvent.Median))
				gapPercent := new(big.Int).Div(new(big.Int).Mul(gap, big.NewInt(100)), penalizeEvent.Median)
				metrics.GetOrRegisterGauge(monitor.OutlierDistancePercentMetric, nil).Update(gapPercent.Int64())
			}

			if penalizeEvent.SlashingAmount.Cmp(common.Big0) == 0 {
				os.logger.Warn("Client addressed as an outlier, the last vote won't be counted for reward distribution, "+
					"please use high quality data source.", "symbol", penalizeEvent.Symbol, "median value",
					penalizeEvent.Median.String(), "reported value", penalizeEvent.Reported.String())
				os.logger.Warn("IMPORTANT: please double check your data source setup before getting penalized")
				if metrics.Enabled {
					metrics.GetOrRegisterCounter(monitor.OutlierNoSlashTimesMetric, nil).Inc(1)
				}
				continue
			}

			os.logger.Warn("Client get penalized as an outlier", "node", penalizeEvent.Participant,
				"currency symbol", penalizeEvent.Symbol, "median value", penalizeEvent.Median.String(),
				"reported value", penalizeEvent.Reported.String(), "block", penalizeEvent.Raw.BlockNumber, "slashed amount", penalizeEvent.SlashingAmount.Uint64())
			os.logger.Warn("your next vote will be postponed", "in blocks", os.conf.VoteBuffer)
			os.logger.Warn("IMPORTANT: please repair your data setups for data precision before getting penalized again")

			if metrics.Enabled {
				metrics.GetOrRegisterCounter(monitor.OutlierSlashTimesMetric, nil).Inc(1)
				baseUnitsPerNTN := new(big.Float).SetInt(big.NewInt(1e18))
				amount := new(big.Float).SetUint64(penalizeEvent.SlashingAmount.Uint64())
				ntnFloat, _ := new(big.Float).Quo(amount, baseUnitsPerNTN).Float64()
				metrics.GetOrRegisterGaugeFloat64(monitor.OutlierPenaltyMetric, nil).Update(ntnFloat)
			}

			if err := os.memories.flushRecord(&OutlierRecord{
				LastPenalizedAtBlock: penalizeEvent.Raw.BlockNumber,
				Participant:          penalizeEvent.Participant,
				Symbol:               penalizeEvent.Symbol,
				Median:               penalizeEvent.Median.Uint64(),
				Reported:             penalizeEvent.Reported.Uint64(),
				SlashingAmount:       penalizeEvent.SlashingAmount.Uint64(),
				LoggedAt:             time.Now().Format(time.RFC3339),
			}); err != nil {
				os.logger.Error("failed to flush outlier record", "error", err.Error())
			}
		case fsEvent, ok := <-os.configWatcher.Events:
			if !ok {
				os.logger.Error("config watcher channel has been closed")
				return
			}
			// filter unwatched files in the dir.
			if fsEvent.Name != os.conf.ConfigFile {
				continue
			}

			switch {
			// tools like sed issues write event for the updates.
			case fsEvent.Op&fsnotify.Write > 0:
				// apply plugin config changes.
				os.logger.Info("config file content changed", "file", fsEvent.Name)
				os.PluginRuntimeManagement()

			// tools like vim or vscode issues rename, chmod and remove events for the update for an atomic change mode.
			case fsEvent.Op&fsnotify.Rename > 0:
				os.logger.Info("config file changed", "file", fsEvent.Name)
				os.PluginRuntimeManagement()
			}

		case fsEvent, ok := <-os.pluginsWatcher.Events:
			if !ok {
				os.logger.Error("plugin watcher channel has been closed")
				return
			}

			os.logger.Info("watched plugins fs event", "file", fsEvent.Name, "event", fsEvent.Op.String())
			// updates on the watched plugin directory will trigger plugin management.
			os.PluginRuntimeManagement()

		case roundEvent := <-os.chRoundEvent:
			os.logger.Info("handle new round", "round", roundEvent.Round.Uint64(), "required sampling TS",
				roundEvent.Timestamp.Uint64(), "height", roundEvent.Raw.BlockNumber, "round period", roundEvent.VotePeriod.Uint64())

			if metrics.Enabled {
				metrics.GetOrRegisterGauge(monitor.RoundMetric, nil).Update(roundEvent.Round.Int64())
			}

			// save the round rotation info to coordinate the pre-sampling.
			os.curRound = roundEvent.Round.Uint64()
			os.votePeriod = roundEvent.VotePeriod.Uint64()
			os.curSampleHeight = roundEvent.Raw.BlockNumber
			os.curSampleTS = roundEvent.Timestamp.Int64()

			// sync protocol symbols, and submit round vote by according to node's membership.
			err := os.handleRoundVote()
			if err != nil {
				os.logger.Error("round voting failed", "error", err.Error())
			}
			// at the end of each round, gc expired samples of per plugin.
			os.gcExpiredSamples()
			// after vote finished, reset sampling symbols by the latest synced protocol symbols.
			os.ResetSamplingSymbols(os.protocolSymbols)
		case newSymbolEvent := <-os.chSymbolsEvent:
			// New symbols are added, add them into the sampling set to prepare data in advance for the coming round's vote.
			os.logger.Info("handle new symbols", "new symbols", newSymbolEvent.Symbols, "activate at round", newSymbolEvent.Round)
			os.handleNewSymbolsEvent(newSymbolEvent.Symbols)
		case <-os.regularTicker.C:
			os.gcRoundData()
			if metrics.Enabled {
				metrics.GetOrRegisterGauge(monitor.PluginMetric, nil).Update(int64(len(os.runningPlugins)))
			}
			os.logger.Debug("current round ID", "round", os.curRound)
		}
	}
}

func (os *OracleServer) Stop() {
	os.client.Close()
	os.subRoundEvent.Unsubscribe()
	os.subSymbolsEvent.Unsubscribe()
	os.subPenalizedEvent.Unsubscribe()
	os.subVotedEvent.Unsubscribe()
	os.subInvalidVote.Unsubscribe()
	os.subRewardEvent.Unsubscribe()

	os.doneCh <- struct{}{}
	for _, c := range os.runningPlugins {
		p := c
		p.Close()
	}
}

func (os *OracleServer) PluginRuntimeManagement() {
	// load plugin configs before start them.
	newConfs, err := config.LoadPluginsConfig(os.conf.ConfigFile)
	if err != nil {
		os.logger.Error("cannot load plugin configuration", "error", err.Error())
		return
	}

	// load plugin binaries
	binaries, err := helpers.ListPlugins(os.conf.PluginDIR)
	if err != nil {
		os.logger.Error("list plugin", "error", err.Error())
		return
	}

	// shutdown the plugins which are removed, disabled or with config update.
	for name, plugin := range os.runningPlugins {
		// shutdown the plugins that were removed.
		if _, ok := binaries[name]; !ok {
			os.logger.Info("removing plugin", "name", name)
			plugin.Close()
			delete(os.runningPlugins, name)
			continue
		}

		// shutdown the plugins that are runtime disabled.
		newConf := newConfs[name]
		if newConf.Disabled {
			os.logger.Info("disabling plugin", "name", name)
			plugin.Close()
			delete(os.runningPlugins, name)
			continue
		}

		// shutdown the plugins that with config updates, they will be reloaded after the shutdown.
		if plugin.Config().Diff(&newConf) {
			os.logger.Info("updating plugin", "name", name)
			plugin.Close()
			delete(os.runningPlugins, name)
		}
	}

	// try to load new plugins.
	for _, file := range binaries {
		f := file
		pConf := newConfs[f.Name()]

		if pConf.Disabled {
			continue
		}

		// skip to set up plugins until there is a service key is presented at plugin-confs.yml
		if _, ok := os.keyRequiredPlugins[f.Name()]; ok && pConf.Key == "" {
			continue
		}

		os.tryToLaunchPlugin(f, pConf)
	}

	if metrics.Enabled {
		metrics.GetOrRegisterGauge("oracle/plugins", nil).Update(int64(len(os.runningPlugins)))
	}
}

func (os *OracleServer) tryToLaunchPlugin(f fs.FileInfo, plugConf config.PluginConfig) {
	plugin, ok := os.runningPlugins[f.Name()]
	if !ok {
		os.logger.Info("new plugin discovered, going to setup it: ", f.Name(), f.Mode().String())
		pluginWrapper, err := os.setupNewPlugin(f.Name(), &plugConf)
		if err != nil {
			return
		}
		os.runningPlugins[f.Name()] = pluginWrapper
		return
	}

	if f.ModTime().After(plugin.StartTime()) || plugin.Exited() {
		os.logger.Info("replacing legacy plugin with new one: ", f.Name(), f.Mode().String())
		// stop the legacy plugin
		plugin.Close()
		delete(os.runningPlugins, f.Name())

		pluginWrapper, err := os.setupNewPlugin(f.Name(), &plugConf)
		if err != nil {
			return
		}
		os.runningPlugins[f.Name()] = pluginWrapper
	}
}

func (os *OracleServer) setupNewPlugin(name string, conf *config.PluginConfig) (*pWrapper.PluginWrapper, error) {
	if err := os.ApplyPluginConf(name, conf); err != nil {
		os.logger.Error("apply plugin config", "error", err.Error())
		return nil, err
	}

	pluginWrapper := pWrapper.NewPluginWrapper(os.conf.LoggingLevel, name, os.conf.PluginDIR, os, conf)
	if err := pluginWrapper.Initialize(os.chainID); err != nil {
		// if the plugin states that a service key is missing, then we mark it down, thus the runtime discovery can
		// skip those plugins without a key configured.
		if errors.Is(err, types.ErrMissingServiceKey) {
			os.keyRequiredPlugins[name] = struct{}{}
		}
		os.logger.Error("cannot run plugin", "name", name, "error", err.Error())
		pluginWrapper.CleanPluginProcess()
		return nil, err
	}

	return pluginWrapper, nil
}

func (os *OracleServer) WatchSampleEvent(sink chan<- *types.SampleEvent) event.Subscription {
	return os.sampleEventFeed.Subscribe(sink)
}

func (os *OracleServer) ApplyPluginConf(name string, plugConf *config.PluginConfig) error {
	// set the plugin configuration via system env, thus the plugin can load it on startup.
	conf, err := json.Marshal(plugConf)
	if err != nil {
		os.logger.Error("cannot marshal plugin's configuration", "error", err.Error())
		return err
	}
	if err = o.Setenv(name, string(conf)); err != nil {
		os.logger.Error("cannot set plugin configuration via system ENV")
		return err
	}
	return nil
}

// ComputeConfidence calculates the confidence weight based on the number of data samples. Note! Cryptos take
// fixed strategy as we have very limited number of data sources at the genesis phase. Thus, the confidence
// computing is just for forex currencies for the time being.
func ComputeConfidence(symbol string, numOfSamples, strategy int) uint8 {

	// Todo: once the community have more extensive AMM and DEX markets, we will remove this to enable linear
	//  strategy as well for cryptos.
	if _, is := common2.ForexCurrencies[symbol]; !is {
		return MaxConfidence
	}

	// Forex currencies with fixed strategy.
	if strategy == config.ConfidenceStrategyFixed {
		return MaxConfidence
	}

	// Forex currencies with "linear" strategy. Labeled "linear" but uses exponential scaling (1.75^n) since we
	// are at the network bootstrapping phase with very limited number of data sources.
	weight := BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, float64(numOfSamples)))

	if weight > MaxConfidence {
		weight = MaxConfidence
	}

	return uint8(weight) //nolint
}

// by according to the spreading of price timestamp from the target timestamp,
// we reduce the confidence of the price, set the lowest confidence as 1.
func confidenceAdjustedPrice(historicRoundPrice *types.Price, target int64) (*types.Price, error) {
	// Calculate the time difference between the target timestamp and the historic price timestamp
	timeDifference := target - historicRoundPrice.Timestamp

	var reducedConfidence uint8
	if timeDifference < 60 { // Less than 1 minute
		reducedConfidence = historicRoundPrice.Confidence // Keep original confidence
	} else if timeDifference < 3600 { // Less than 1 hour
		reducedConfidence = historicRoundPrice.Confidence / 2 // Reduce confidence by half
	} else {
		reducedConfidence = 1 // Set the lowest confidence to 1 if more than 1 hour old
	}

	if reducedConfidence == 0 {
		return nil, types.ErrNoAvailablePrice
	}

	historicRoundPrice.Confidence = reducedConfidence
	return historicRoundPrice, nil
}
