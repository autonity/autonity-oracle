package oracleserver

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/helpers"
	pWrapper "autonity-oracle/plugin_wrapper"
	common2 "autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	tp "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	"github.com/shopspring/decimal"
	"io/fs"
	"math"
	"math/big"
	o "os"
	"path/filepath"
	"time"
)

var ForexCurrencies = map[string]struct{}{
	"AUD-USD": {},
	"CAD-USD": {},
	"EUR-USD": {},
	"GBP-USD": {},
	"JPY-USD": {},
	"SEK-USD": {},
}

var (
	saltRange        = new(big.Int).SetUint64(math.MaxInt64)
	alertBalance     = new(big.Int).SetUint64(2000000000000) // 2000 Gwei, 0.000002 Ether
	invalidPrice     = big.NewInt(0)
	invalidSalt      = big.NewInt(0)
	tenSecsInterval  = 10 * time.Second                             // ticker to check L2 connectivity and gc round data.
	oneSecsInterval  = 1 * time.Second                              // sampling interval during data pre-sampling period.
	preSamplingRange = uint64(5)                                    // pre-sampling starts in 5 blocks in advance.
	bridgerSymbols   = []string{"ATN-USDC", "NTN-USDC", "USDC-USD"} // used for value bridging to USD by USDC

	numOfPlugins       = metrics.GetOrRegisterGauge("oracle/plugins", nil)
	oracleRound        = metrics.GetOrRegisterGauge("oracle/round", nil)
	slashEventCounter  = metrics.GetOrRegisterCounter("oracle/slash", nil)
	l1ConnectivityErrs = metrics.GetOrRegisterCounter("oracle/l1/errs", nil)
	accountBalance     = metrics.GetOrRegisterGauge("oracle/balance", nil)
	isVoterFlag        = metrics.GetOrRegisterGauge("oracle/isVoter", nil)
)

const (
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
	serverStateDumpFile = "server_state_dump.json"
)

// ServerMemories is the state that to be flushed into the profiling report directory.
// For the time being, it just contains the last outlier record of the server. It is loaded on start up.
type ServerMemories struct {
	OutlierRecord
	LoggedAt string `json:"logged_at"`
}

type OutlierRecord struct {
	LastPenalizedAtBlock uint64         `json:"last_penalized_at_block"`
	Participant          common.Address `json:"participant"`
	Symbol               string         `json:"symbol"`
	Median               uint64         `json:"median"`
	Reported             uint64         `json:"reported"`
}

// flush dumps the ServerMemories into a JSON file in the specified profile directory.
func (s *ServerMemories) flush(profileDir string) error {
	if _, err := o.Stat(profileDir); o.IsNotExist(err) {
		return fmt.Errorf("profile directory does not exist: %s", profileDir)
	}

	fileName := filepath.Join(profileDir, serverStateDumpFile)

	// Create or open the file for writing
	file, err := o.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(s); err != nil {
		return fmt.Errorf("failed to encode ServerMemories to JSON: %v", err)
	}

	return nil
}

// loadState loads the ServerMemories from a JSON file in the specified profile directory.
func (s *ServerMemories) loadState(profileDir string) error {
	fileName := filepath.Join(profileDir, serverStateDumpFile)

	file, err := o.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(s); err != nil {
		return fmt.Errorf("failed to decode JSON into ServerMemories: %v", err)
	}

	return nil
}

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

	serverMemories *ServerMemories // server memories to be flushed.

	fsWatcher *fsnotify.Watcher // FS watcher watches the changes of plugins and the plugins' configs.
	chainID   int64             // ChainID saves the L1 chain ID, it is used for plugin compatibility check.
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
	state := &ServerMemories{}
	err = state.loadState(os.conf.ProfileDir)
	if err == nil {
		os.logger.Info("run oracle server with historical flushed state", "state", state)
		os.serverMemories = state
	}

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

	// subscribe FS notifications of the watched plugins and config file.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		os.logger.Error("cannot create fsnotify watcher", "error", err)
		o.Exit(1)
	}

	err = watcher.Add(conf.PluginDIR)
	if err != nil {
		os.logger.Error("cannot watch plugin dir", "error", err)
		o.Exit(1)
	}

	err = watcher.Add(conf.ConfigFile)
	if err != nil {
		os.logger.Error("cannot watch config file", "error", err)
		o.Exit(1)
	}
	os.fsWatcher = watcher
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

	os.logger.Info("syncStates", "CurrentRound", os.curRound, "Num of AvailableSymbols", len(os.protocolSymbols), "CurrentSymbols", os.protocolSymbols)
	os.AddNewSymbols(os.protocolSymbols)
	os.logger.Info("syncStates", "CurrentRound", os.curRound, "Num of bridgerSymbols", len(bridgerSymbols), "bridgerSymbols", bridgerSymbols)
	os.AddNewSymbols(bridgerSymbols)

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

func (os *OracleServer) gcDataSamples() {
	for _, plugin := range os.runningPlugins {
		plugin.GCSamples()
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
				l1ConnectivityErrs.Inc(1)
			}
			return
		}
		os.lostSync = false
		return
	}

	os.logger.Debug("checking heart beat", "current round", os.curRound)
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
	if nextRoundHeight-curHeight > preSamplingRange {
		return nil
	}

	// do the data pre-sampling.
	os.logger.Debug("data pre-sampling", "on height", curHeight, "TS", preSampleTS)
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
		os.logger.Debug("no last round data, client is no longer a voter or it will report with commitment hash")
	}

	// if node is no longer a validator, and it doesn't have last round data, skip reporting.
	if !isVoter && !ok {
		os.logger.Debug("skip data reporting since client is no longer a voter")
		if metrics.Enabled {
			isVoterFlag.Update(0)
		}
		return nil
	}

	// check with the vote buffer from the last penalty event.
	if os.serverMemories != nil && os.curSampleHeight-os.serverMemories.LastPenalizedAtBlock <= os.conf.VoteBuffer {
		left := os.conf.VoteBuffer - (os.curSampleHeight - os.serverMemories.LastPenalizedAtBlock)
		os.logger.Warn("due to the outlier penalty, we postpone your next vote from slashing", "next vote block", left)
		os.logger.Warn("your last outlier report was", "report", os.serverMemories.OutlierRecord)
		os.logger.Warn("during this period, you can: 1. check your data source infra; 2. restart your oracle-client; 3. contact Autonity team for help;")
		return nil
	}

	if isVoter {
		if metrics.Enabled {
			isVoterFlag.Update(1)
		}
		// report with last round data and with current round commitment hash.
		return os.reportWithCommitment(os.curRound, lastRoundData)
	}

	// voter reports with last round data but without current round commitment since it is not committee member now.
	return os.reportWithoutCommitment(lastRoundData)
}

func (os *OracleServer) reportWithCommitment(newRound uint64, lastRoundData *types.RoundData) error {
	curRoundData, err := os.buildRoundData(newRound)
	if err != nil {
		os.logger.Error("build round data", "error", err)
		return err
	}

	// save current round data.
	os.roundData[newRound] = curRoundData

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
		accountBalance.Update(balance.Int64())
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

// AddNewSymbols adds new symbols to the local symbol set for data fetching, duplicated one is not added.
func (os *OracleServer) AddNewSymbols(newSymbols []string) {
	var symbolsMap = make(map[string]struct{})
	for _, s := range os.samplingSymbols {
		symbolsMap[s] = struct{}{}
	}

	for _, newS := range newSymbols {
		if _, ok := symbolsMap[newS]; !ok {
			os.samplingSymbols = append(os.samplingSymbols, newS)
		}
	}
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

	auth.Value = big.NewInt(0)
	auth.GasTipCap = new(big.Int).SetUint64(os.conf.GasTipCap)
	auth.GasLimit = uint64(3000000)

	// if there is no last round data or there were missing datapoint in last round data, then we just submit the
	// commitment hash of current round as data might be available at current round. This vote will be reimbursed by the
	// protocol, however it won't be abused as it is limited by the 1 vote per round rule.
	if lastRoundData == nil || lastRoundData.MissingData {
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
	usdcPrice, err := os.aggregatePrice(USDCUSD, os.curSampleTS)
	if err != nil {
		os.logger.Error("aggregate USDC-USD price", "error", err.Error())
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
				os.logger.Info("zero price measured from market", "symbol", s)
				missingData = true
			}
			reports = append(reports, contract.IOracleReport{
				Price:      price,
				Confidence: pr.Confidence,
			})
		} else {
			// logging the missing of data points for all symbols
			missingData = true
			os.logger.Info("round report miss data point for symbol", "symbol", s)
			reports = append(reports, contract.IOracleReport{
				Price: invalidPrice,
			})
		}
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

	roundData.MissingData = missingData
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

func (os *OracleServer) aggregatePrice(s string, target int64) (*types.Price, error) {
	var prices []decimal.Decimal
	for _, plugin := range os.runningPlugins {
		p, err := plugin.GetSample(s, target)
		if err != nil {
			continue
		}
		prices = append(prices, p.Price)
	}

	if len(prices) == 0 {
		return nil, types.ErrNoDataRound
	}

	// compute confidence of the symbol from the num of plugins' samples of it.
	confidence := ComputeConfidence(s, len(prices), os.conf.ConfidenceStrategy)

	price := &types.Price{
		Timestamp:  target,
		Price:      prices[0],
		Symbol:     s,
		Confidence: confidence,
	}

	// we have multiple provider provide prices for this symbol, we have to aggregate it.
	if len(prices) > 1 {
		p, err := helpers.Median(prices)
		if err != nil {
			return nil, err
		}
		price.Price = p
	}

	return price, nil
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
			os.logger.Info("oracle service is stopped")
			return

		case err, ok := <-os.fsWatcher.Errors:
			if !ok {
				os.logger.Error("failed to watch filesystem")
				return
			}
			os.logger.Error("fs-watcher errors", "err", err.Error())

		case err := <-os.subSymbolsEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new symbols event", err)
				os.handleConnectivityError()
				os.subSymbolsEvent.Unsubscribe()
			}
		case err := <-os.subRoundEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new roundEvent", err)
				os.handleConnectivityError()
				os.subRoundEvent.Unsubscribe()
			}
		case err := <-os.subPenalizedEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new roundEvent", err)
				os.handleConnectivityError()
				os.subPenalizedEvent.Unsubscribe()
			}
		case <-os.psTicker.C:
			preSampleTS := time.Now().Unix()
			err := os.handlePreSampling(preSampleTS)
			if err != nil {
				os.logger.Error("handle pre-sampling", "error", err.Error())
			}
			os.lastSampledTS = preSampleTS
		case penalizeEvent := <-os.chPenalizedEvent:

			os.logger.Warn("Oracle client get penalized as an outlier", "node", penalizeEvent.Participant,
				"currency symbol", penalizeEvent.Symbol, "median value", penalizeEvent.Median.String(),
				"reported value", penalizeEvent.Reported.String(), "block", penalizeEvent.Raw.BlockNumber)
			os.logger.Warn("your next vote will be postponed", "in blocks", os.conf.VoteBuffer)

			if metrics.Enabled {
				slashEventCounter.Inc(1)
			}

			newState := &ServerMemories{
				OutlierRecord: OutlierRecord{
					LastPenalizedAtBlock: penalizeEvent.Raw.BlockNumber,
					Participant:          penalizeEvent.Participant,
					Symbol:               penalizeEvent.Symbol,
					Median:               penalizeEvent.Median.Uint64(),
					Reported:             penalizeEvent.Reported.Uint64(),
				},
				LoggedAt: time.Now().Format(time.RFC3339),
			}

			os.serverMemories = newState
			if err := os.serverMemories.flush(os.conf.ProfileDir); err != nil {
				os.logger.Error("failed to flush oracle state", "error", err.Error())
			}
		case fsEvent, ok := <-os.fsWatcher.Events:
			if !ok {
				os.logger.Error("fs watcher has been closed")
			}

			os.logger.Info("watched new fs event", "file", fsEvent.Name, "event", fsEvent.Op.String())
			// updates on the watched config and plugin directory will trigger plugin management.
			os.PluginRuntimeManagement()

		case roundEvent := <-os.chRoundEvent:
			os.logger.Info("handle new round", "round", roundEvent.Round.Uint64(), "required sampling TS",
				roundEvent.Timestamp.Uint64(), "height", roundEvent.Height.Uint64(), "round period", roundEvent.VotePeriod.Uint64())

			if metrics.Enabled {
				oracleRound.Update(roundEvent.Round.Int64())
			}

			// save the round rotation info to coordinate the pre-sampling.
			os.curRound = roundEvent.Round.Uint64()
			os.votePeriod = roundEvent.VotePeriod.Uint64()
			os.curSampleHeight = roundEvent.Height.Uint64()
			os.curSampleTS = roundEvent.Timestamp.Int64()

			err := os.handleRoundVote()
			if err != nil {
				continue
			}
			os.gcDataSamples()
			// after vote finished, gc useless symbols by protocol required symbols.
			os.samplingSymbols = os.protocolSymbols
			// attach the bridger symbols too once the sampling symbols is replaced by protocol symbols.
			os.AddNewSymbols(bridgerSymbols)
		case newSymbolEvent := <-os.chSymbolsEvent:
			os.logger.Info("handle new symbols", "new symbols", newSymbolEvent.Symbols, "activate at round", newSymbolEvent.Round)
			os.handleNewSymbolsEvent(newSymbolEvent.Symbols)
		case <-os.regularTicker.C:
			os.checkHealth()
			os.gcRoundData()
		}
	}
}

func (os *OracleServer) Stop() {
	os.client.Close()
	os.subRoundEvent.Unsubscribe()
	os.subSymbolsEvent.Unsubscribe()
	os.subPenalizedEvent.Unsubscribe()
	if os.fsWatcher != nil {
		os.fsWatcher.Close() //nolint
	}

	os.doneCh <- struct{}{}
	for _, c := range os.runningPlugins {
		p := c
		p.Close()
	}
}

func (os *OracleServer) PluginRuntimeManagement() {
	// load plugin configs before start them.
	plugConfs, err := config.LoadPluginsConfig(os.conf.ConfigFile)
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

	// shutdown the plugins which are removed from the plugin directory, or disabled from config.
	for name, plugin := range os.runningPlugins {
		if _, ok := binaries[name]; !ok {
			os.logger.Info("removing plugin", "name", name)
			plugin.Close()
			delete(os.runningPlugins, name)
			continue
		}

		// shutdown the plugin that are runtime disabled.
		pConf := plugConfs[name]
		if pConf.Disabled {
			os.logger.Info("disabling plugin", "name", name)
			plugin.Close()
			delete(os.runningPlugins, name)
		}
	}

	// try to load new plugins.
	for _, file := range binaries {
		f := file
		pConf := plugConfs[f.Name()]

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
		numOfPlugins.Update(int64(len(os.runningPlugins)))
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
	if _, is := ForexCurrencies[symbol]; !is {
		return MaxConfidence
	}

	// Forex currencies with fixed strategy.
	if strategy == config.ConfidenceStrategyFixed {
		return MaxConfidence
	}

	// Forex currencies with linear strategy.
	weight := BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, float64(numOfSamples)))

	if weight > MaxConfidence {
		weight = MaxConfidence
	}

	return uint8(weight) //nolint
}
