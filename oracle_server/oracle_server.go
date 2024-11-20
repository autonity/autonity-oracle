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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	tp "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	"github.com/shopspring/decimal"
	"io/fs"
	"math"
	"math/big"
	o "os"
	"time"
)

var (
	TenSecsInterval  = 10 * time.Second // 10s ticker job to check health with l1, plugin discovery and regular data sampling.
	OneSecInterval   = 1 * time.Second  // 1s ticker job to check if we need to do pre-sampling.
	PreSamplingRange = 5                // pre-sampling starts in 5 blocks in advance.
	SaltRange        = new(big.Int).SetUint64(math.MaxInt64)
	AlertBalance     = new(big.Int).SetUint64(2000000000000) // 2000 Gwei, 0.000002 Ether
)

const (
	ATNUSD  = "ATN-USD"
	NTNUSD  = "NTN-USD"
	USDCUSD = "USDC-USD"
	ATNUSDC = "ATN-USDC"
	NTNUSDC = "NTN-USDC"
)

// OracleServer coordinates the plugin discovery, the data sampling, and do the health checking with L1 connectivity.
type OracleServer struct {
	logger        hclog.Logger
	doneCh        chan struct{}
	regularTicker *time.Ticker // the clock source to trigger the 10s interval job.
	psTicker      *time.Ticker // the pre-sampling ticker in 1s.

	pluginDIR       string                             // the dir saves the plugins.
	pluginSet       map[string]*pWrapper.PluginWrapper // the plugin clients that connect with different adapters.
	samplingSymbols []string                           // the symbols for data fetching in oracle service, can be different from the required protocol symbols.

	keyRequiredPlugins map[string]struct{} // saving those plugins which require a key granted by data provider

	// the reporting staffs
	dialer         types.Dialer
	oracleContract contract.ContractAPI
	client         types.Blockchain
	l1WSUrl        string

	curRound        uint64 //round ID.
	votePeriod      uint64 //vote period.
	curSampleTS     uint64 //the data sample TS of the current round.
	curSampleHeight uint64 //The block height on which the round rotation happens.

	protocolSymbols []string //symbols required for the voting on the oracle contract protocol.
	pricePrecision  decimal.Decimal
	roundData       map[uint64]*types.RoundData
	key             *keystore.Key

	pluginConfFile string

	gasTipCap uint64

	chPenalizedEvent  chan *contract.OraclePenalized
	subPenalizedEvent event.Subscription

	chRoundEvent  chan *contract.OracleNewRound
	subRoundEvent event.Subscription

	chSymbolsEvent  chan *contract.OracleNewSymbols
	subSymbolsEvent event.Subscription
	lastSampledTS   int64

	sampleEventFeed        event.Feed
	loggingLevel           hclog.Level
	confidenceStrategy     int
	lostSync               bool // set to true if the connectivity with L1 Autonity network is dropped during runtime.
	commitmentHashComputer *CommitmentHashComputer
}

func NewOracleServer(conf *types.OracleServiceConfig, dialer types.Dialer, client types.Blockchain,
	oc contract.ContractAPI) *OracleServer {
	os := &OracleServer{
		dialer:             dialer,
		client:             client,
		oracleContract:     oc,
		l1WSUrl:            conf.AutonityWSUrl,
		roundData:          make(map[uint64]*types.RoundData),
		key:                conf.Key,
		gasTipCap:          conf.GasTipCap,
		pluginConfFile:     conf.PluginConfFile,
		pluginDIR:          conf.PluginDIR,
		pluginSet:          make(map[string]*pWrapper.PluginWrapper),
		keyRequiredPlugins: make(map[string]struct{}),
		doneCh:             make(chan struct{}),
		regularTicker:      time.NewTicker(TenSecsInterval),
		psTicker:           time.NewTicker(OneSecInterval),
		loggingLevel:       conf.LoggingLevel,
		confidenceStrategy: conf.ConfidenceStrategy,
	}

	os.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(os).String() + conf.Key.Address.String(),
		Output: o.Stdout,
		Level:  conf.LoggingLevel,
	})

	commitmentHashComputer, err := NewCommitmentHashComputer()
	if err != nil {
		os.logger.Error("cannot create commitment hash computer", "err", err)
		o.Exit(1)
	}
	os.commitmentHashComputer = commitmentHashComputer

	// load plugin configs before start them.
	plugConfs, err := config.LoadPluginsConfig(conf.PluginConfFile)
	if err != nil {
		os.logger.Error("cannot load plugin configuration", "error", err.Error(), "path", conf.PluginConfFile)
		helpers.PrintUsage()
		o.Exit(1)
	}

	// discover plugins from plugin dir at startup.
	binaries, err := helpers.ListPlugins(conf.PluginDIR)
	if len(binaries) == 0 || err != nil {
		// to stop the service on the start once there is no plugin in the db.
		os.logger.Error("no plugin discovered", "plugin-dir", os.pluginDIR)
		helpers.PrintUsage()
		o.Exit(1)
	}
	for _, file := range binaries {
		f := file
		pConf := plugConfs[f.Name()]
		os.loadNewPlugin(f, pConf)
	}

	os.logger.Info("running oracle contract listener at", "WS", conf.AutonityWSUrl, "ID", conf.Key.Address.String())
	err = os.syncStates()
	if err != nil {
		// stop the client on start up once the remote endpoint of autonity L1 network is not ready.
		os.logger.Error("Cannot synchronize oracle contract state from Autonity L1 Network", "error", err.Error(), "WS", conf.AutonityWSUrl)
		helpers.PrintUsage()
		o.Exit(1)
	}
	os.lostSync = false
	return os
}

// syncStates is executed on client startup or after the L1 connection recovery to sync the on-chain oracle contract
// states, symbols, round id, precision, vote period, etc... to the oracle server. It also subscribes the on-chain
// events of oracle protocol: round event, symbol update event, etc...
func (os *OracleServer) syncStates() error {
	var err error
	// get initial states from on-chain oracle contract.
	os.curRound, os.protocolSymbols, os.pricePrecision, os.votePeriod, err = os.initStates()
	if err != nil {
		os.logger.Error("synchronize oracle contract state", "error", err.Error())
		return err
	}

	os.logger.Info("syncStates", "CurrentRound", os.curRound, "Num of AvailableSymbols", len(os.protocolSymbols), "CurrentSymbols", os.protocolSymbols)
	os.AddNewSymbols(os.protocolSymbols)
	os.logger.Info("syncStates", "CurrentRound", os.curRound, "Num of BridgerSymbols", len(config.BridgerSymbols), "BridgerSymbols", config.BridgerSymbols)
	os.AddNewSymbols(config.BridgerSymbols)

	// subscribe on-chain round rotation event
	os.chRoundEvent = make(chan *contract.OracleNewRound)
	os.subRoundEvent, err = os.oracleContract.WatchNewRound(new(bind.WatchOpts), os.chRoundEvent)
	if err != nil {
		os.logger.Error("failed to subscribe round event", "error", err.Error())
		return err
	}

	// subscribe on-chain symbol update event
	os.chSymbolsEvent = make(chan *contract.OracleNewSymbols)
	os.subSymbolsEvent, err = os.oracleContract.WatchNewSymbols(new(bind.WatchOpts), os.chSymbolsEvent)
	if err != nil {
		os.logger.Error("failed to subscribe new symbol event", "error", err.Error())
		return err
	}

	// subscribe on-chain penalize event
	os.chPenalizedEvent = make(chan *contract.OraclePenalized)
	os.subPenalizedEvent, err = os.oracleContract.WatchPenalized(new(bind.WatchOpts), os.chPenalizedEvent, nil)
	if err != nil {
		os.logger.Error("failed to subscribe penalized event", "error", err.Error())
		return err
	}

	return nil
}

// initStates returns round id, symbols and committees on current chain, it is called on the startup of client.
func (os *OracleServer) initStates() (uint64, []string, decimal.Decimal, uint64, error) {
	var precision decimal.Decimal
	// on the startup, we need to sync the round id, symbols and committees from contract.
	currentRound, err := os.oracleContract.GetRound(nil)
	if err != nil {
		os.logger.Error("get round", "error", err.Error())
		return 0, nil, precision, 0, err
	}

	symbols, err := os.oracleContract.GetSymbols(nil)
	if err != nil {
		os.logger.Error("get symbols", "error", err.Error())
		return 0, nil, precision, 0, err
	}

	decimals, err := os.oracleContract.GetDecimals(nil)
	if err != nil {
		os.logger.Error("get precision", "error", err.Error())
		return 0, nil, precision, 0, err
	}

	votePeriod, err := os.oracleContract.GetVotePeriod(nil)
	if err != nil {
		os.logger.Error("get vote period", "error", err.Error())
		return 0, nil, precision, 0, nil
	}

	if len(symbols) == 0 {
		os.logger.Error("there are no symbols in Autonity L1 oracle contract")
		return currentRound.Uint64(), symbols, decimal.NewFromBigInt(common.Big1, int32(decimals)), votePeriod.Uint64(), types.ErrNoSymbolsObserved
	}

	return currentRound.Uint64(), symbols, decimal.NewFromBigInt(common.Big1, int32(decimals)), votePeriod.Uint64(), nil
}

func (os *OracleServer) gcDataSamples() {
	for _, plugin := range os.pluginSet {
		plugin.GCSamples()
	}
}

func (os *OracleServer) gcRoundData() {
	if len(os.roundData) >= types.MaxBufferedRounds {
		offset := os.curRound - types.MaxBufferedRounds
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
		if err != nil && err != types.ErrNoSymbolsObserved {
			os.logger.Info("rebuilding WS connectivity with Autonity L1 node", "error", err)
			return
		}
		os.lostSync = false
		return
	}

	h, err := os.client.BlockNumber(context.Background())
	if err != nil {
		os.logger.Error("get block number", "error", err.Error())
		return
	}

	r, err := os.oracleContract.GetRound(nil)
	if err != nil {
		os.logger.Error("get round", "error", err.Error())
		return
	}
	os.logger.Debug("checking heart beat", "current height", h, "current round", r.Uint64())
}

func (os *OracleServer) isVoter() (bool, error) {
	voters, err := os.oracleContract.GetVoters(nil)
	if err != nil {
		os.logger.Error("get voters", "error", err.Error())
		return false, err
	}

	for _, c := range voters {
		if c == os.key.Address {
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
	nSampleHeight := os.curSampleHeight + os.votePeriod
	curHeight, err := os.client.BlockNumber(context.Background())
	if err != nil {
		os.logger.Error("handle pre-sampling", "error", err.Error())
		return err
	}
	if nSampleHeight-curHeight > uint64(PreSamplingRange) {
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
		return nil
	}

	if isVoter {
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

	// prepare the transaction which carry current round's commitment, and last round's data.
	curRoundData.Tx, err = os.doReport(curRoundData.CommitmentHash, lastRoundData)
	if err != nil {
		os.logger.Error("do report", "error", err.Error())
		return err
	}

	// save current round data.
	os.roundData[newRound] = curRoundData
	os.logger.Info("reported last round data and with current round commitment", "TX hash", curRoundData.Tx.Hash(), "Nonce", curRoundData.Tx.Nonce(), "Cost", curRoundData.Tx.Cost())

	// alert in case of balance reach the warning value.
	balance, err := os.client.BalanceAt(context.Background(), os.key.Address, nil)
	if err != nil {
		os.logger.Error("cannot get account balance", "error", err.Error())
		return err
	}

	os.logger.Info("oracle server account", "address", os.key.Address, "remaining balance", balance.String())
	if balance.Cmp(AlertBalance) <= 0 {
		os.logger.Warn("oracle account has too less balance left for data reporting", "balance", balance.String())
	}

	return nil
}

// report with last round data but without current round commitment, voter is leaving from the committee.
func (os *OracleServer) reportWithoutCommitment(lastRoundData *types.RoundData) error {

	// report with no commitment
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

	auth, err := bind.NewKeyedTransactorWithChainID(os.key.PrivateKey, chainID)
	if err != nil {
		os.logger.Error("new keyed transactor with chain ID", "error", err)
		return nil, err
	}

	auth.Value = big.NewInt(0)
	auth.GasTipCap = new(big.Int).SetUint64(os.gasTipCap)
	auth.GasLimit = uint64(3000000)

	// if there is no last round data, then we just submit the curRndCommitHash hash of current round.
	if lastRoundData == nil {
		var reports []contract.IOracleReport
		return os.oracleContract.Vote(auth, new(big.Int).SetBytes(curRoundCommitmentHash.Bytes()), reports, types.InvalidSalt, config.Version)
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
	usdcPrice, err := os.aggregatePrice(USDCUSD, int64(os.curSampleTS))
	if err != nil {
		os.logger.Error("aggregate USDC-USD price", "error", err.Error())
	}

	for _, s := range os.protocolSymbols {
		// aggregate bridged symbols
		if s == ATNUSD || s == NTNUSD {
			if usdcPrice == nil {
				continue
			}

			p, e := os.aggregateBridgedPrice(s, int64(os.curSampleTS), usdcPrice)
			if e != nil {
				os.logger.Error("aggregate bridged price", "error", e.Error(), "symbol", s)
				continue
			}
			prices[s] = *p
			continue
		}

		// aggregate none bridged symbols
		p, e := os.aggregatePrice(s, int64(os.curSampleTS))
		if e != nil {
			os.logger.Debug("no data for aggregation", "reason", e.Error(), "symbol", s)
			continue
		}
		prices[s] = *p
	}

	if len(prices) == 0 {
		return nil, types.ErrNoAvailablePrice
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

	var reports []contract.IOracleReport
	for _, s := range symbols {
		if pr, ok := prices[s]; ok {
			reports = append(reports, contract.IOracleReport{
				Price:      pr.Price.Mul(os.pricePrecision).BigInt(),
				Confidence: pr.Confidence,
			})
		} else {
			reports = append(reports, contract.IOracleReport{
				Price: types.InvalidPrice,
			})
		}
	}

	salt, err := rand.Int(rand.Reader, SaltRange)
	if err != nil {
		os.logger.Error("generate rand salt", "error", err.Error())
		return nil, err
	}

	commitmentHash, err := os.commitmentHashComputer.CommitmentHash(reports, salt, os.key.Address)
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

func (os *OracleServer) aggregatePrice(s string, target int64) (*types.Price, error) {
	var prices []decimal.Decimal
	for _, plugin := range os.pluginSet {
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
	confidence := config.ComputeConfidence(s, len(prices), os.confidenceStrategy)

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

		case err := <-os.subSymbolsEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new symbols event", err)
				os.handleConnectivityError()
				os.subSymbolsEvent.Unsubscribe()
			}
		case err := <-os.subRoundEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new rEvent event", err)
				os.handleConnectivityError()
				os.subRoundEvent.Unsubscribe()
			}
		case err := <-os.subPenalizedEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new rEvent event", err)
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
			if penalizeEvent.Participant == os.key.Address {
				os.logger.Info("Your oracle client get penalized as an outlier")
				os.logger.Info("observed oracle penalize event", "participant", penalizeEvent.Participant,
					"symbol", penalizeEvent.Symbol, "median", penalizeEvent.Median.String(), "reported", penalizeEvent.Reported.String())
				if os.confidenceStrategy == config.ConfidenceStrategyFixed {
					os.logger.Info("confidence strategy switch to linear from fixed to reduce the penalty risk")
					os.confidenceStrategy = config.ConfidenceStrategyLinear
				}

				if os.confidenceStrategy == config.ConfidenceStrategyLinear && config.SourceScalingFactor > 0 {
					config.SourceScalingFactor--
					os.logger.Info("reduce the source scaling factor to reduce the penalty risk", "scaling factor", config.SourceScalingFactor)
				}
			}
		case rEvent := <-os.chRoundEvent:
			if os.curRound == rEvent.Round.Uint64() {
				os.logger.Debug("skip duplicated round event", "round", rEvent.Round)
				continue
			}

			os.logger.Info("handle new round", "round", rEvent.Round.Uint64(), "required sampling TS",
				rEvent.Timestamp.Uint64(), "height", rEvent.Height.Uint64(), "round period", rEvent.VotePeriod.Uint64())

			// save the round rotation info to coordinate the pre-sampling.
			os.curRound = rEvent.Round.Uint64()
			os.votePeriod = rEvent.VotePeriod.Uint64()
			os.curSampleHeight = rEvent.Height.Uint64()
			os.curSampleTS = rEvent.Timestamp.Uint64()

			err := os.handleRoundVote()
			if err != nil {
				continue
			}
			os.gcDataSamples()
			// after vote finished, gc useless symbols by protocol required symbols.
			os.samplingSymbols = os.protocolSymbols
			// attach the bridger symbols too once the sampling symbols is replaced by protocol symbols.
			os.AddNewSymbols(config.BridgerSymbols)
		case symbols := <-os.chSymbolsEvent:
			os.logger.Info("handle new symbols", "new symbols", symbols.Symbols, "activate at round", symbols.Round)
			os.handleNewSymbolsEvent(symbols.Symbols)
		case <-os.regularTicker.C:
			// start the regular price sampling for oracle service on each 10s.
			now := time.Now().Unix()
			os.logger.Debug("regular 10s data sampling", "ts", now)
			os.samplePrice(os.samplingSymbols, now)
			os.lastSampledTS = now
			os.checkHealth()
			os.PluginRuntimeDiscovery()
			os.gcRoundData()
		}
	}
}

func (os *OracleServer) Stop() {
	os.client.Close()
	os.subRoundEvent.Unsubscribe()
	os.subSymbolsEvent.Unsubscribe()
	os.subPenalizedEvent.Unsubscribe()

	os.doneCh <- struct{}{}
	for _, c := range os.pluginSet {
		p := c
		p.Close()
	}
}

func (os *OracleServer) PluginRuntimeDiscovery() {
	// load plugin configs before start them.
	plugConfs, err := config.LoadPluginsConfig(os.pluginConfFile)
	if err != nil {
		os.logger.Error("cannot load plugin configuration", "error", err.Error())
		return
	}

	binaries, err := helpers.ListPlugins(os.pluginDIR)
	if err != nil {
		os.logger.Error("list plugin", "error", err.Error())
		return
	}
	for _, file := range binaries {
		f := file
		pConf := plugConfs[f.Name()]

		// skip to set up plugins until there is a service key is presented at plugin-confs.yml
		if _, ok := os.keyRequiredPlugins[f.Name()]; ok && pConf.Key == "" {
			continue
		}

		os.loadNewPlugin(f, pConf)
	}
}

func (os *OracleServer) loadNewPlugin(f fs.FileInfo, plugConf types.PluginConfig) {
	plugin, ok := os.pluginSet[f.Name()]
	if !ok {
		os.logger.Info("new plugin discovered, going to setup it: ", f.Name(), f.Mode().String())
		pluginWrapper, err := os.setupNewPlugin(f.Name(), &plugConf)
		if err != nil {
			return
		}
		os.pluginSet[f.Name()] = pluginWrapper
		return
	}

	if f.ModTime().After(plugin.StartTime()) || plugin.Exited() {
		os.logger.Info("replacing legacy plugin with new one: ", f.Name(), f.Mode().String())
		// stop the legacy plugin
		plugin.Close()
		delete(os.pluginSet, f.Name())

		pluginWrapper, err := os.setupNewPlugin(f.Name(), &plugConf)
		if err != nil {
			return
		}
		os.pluginSet[f.Name()] = pluginWrapper
	}
}

func (os *OracleServer) setupNewPlugin(name string, conf *types.PluginConfig) (*pWrapper.PluginWrapper, error) {
	if err := os.ApplyPluginConf(name, conf); err != nil {
		os.logger.Error("apply plugin config", "error", err.Error())
		return nil, err
	}

	pluginWrapper := pWrapper.NewPluginWrapper(os.loggingLevel, name, os.pluginDIR, os, conf)
	if err := pluginWrapper.Initialize(); err != nil {
		// if the plugin states that a service key is missing, then we mark it down, thus the runtime discovery can
		// skip those plugins without a key configured.
		if err == types.ErrMissingServiceKey {
			os.keyRequiredPlugins[name] = struct{}{}
		}
		os.logger.Error("cannot setup plugin", "name", name, "error", err.Error())
		pluginWrapper.CleanPluginProcess()
		return nil, err
	}

	return pluginWrapper, nil
}

func (os *OracleServer) WatchSampleEvent(sink chan<- *types.SampleEvent) event.Subscription {
	return os.sampleEventFeed.Subscribe(sink)
}

func (os *OracleServer) ApplyPluginConf(name string, plugConf *types.PluginConfig) error {
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
