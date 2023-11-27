package oracleserver

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/helpers"
	pWrapper "autonity-oracle/plugin_wrapper"
	"autonity-oracle/types"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	tp "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
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
	PreSamplingRange = 15               // pre-sampling starts in 15blocks in advance.
	SaltRange        = new(big.Int).SetUint64(math.MaxInt64)
	AlertBalance     = new(big.Int).SetUint64(2000000000000) // 2000 Gwei, 0.000002 Ether
)

// OracleServer coordinates the plugin discovery, the data sampling, and do the health checking with L1 connectivity.
type OracleServer struct {
	logger        hclog.Logger
	doneCh        chan struct{}
	regularTicker *time.Ticker // the clock source to trigger the 10s interval job.
	psTicker      *time.Ticker // the pre-sampling ticker in 1s.

	pluginDIR string                             // the dir saves the plugins.
	pluginSet map[string]*pWrapper.PluginWrapper // the plugin clients that connect with different adapters.
	symbols   []string                           // the symbols for data fetching in oracle service.

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

	chRoundEvent  chan *contract.OracleNewRound
	subRoundEvent event.Subscription

	chSymbolsEvent  chan *contract.OracleNewSymbols
	subSymbolsEvent event.Subscription
	lastSampledTS   int64

	sampleEventFee event.Feed
	lostSync       bool // set to true if the connectivity with L1 Autonity network is dropped during runtime.
}

func NewOracleServer(conf *types.OracleServiceConfig, dialer types.Dialer, client types.Blockchain,
	oc contract.ContractAPI) *OracleServer {
	os := &OracleServer{
		dialer:         dialer,
		client:         client,
		oracleContract: oc,
		l1WSUrl:        conf.AutonityWSUrl,
		roundData:      make(map[uint64]*types.RoundData),
		key:            conf.Key,
		gasTipCap:      conf.GasTipCap,
		pluginConfFile: conf.PluginConfFile,
		symbols:        conf.Symbols,
		pluginDIR:      conf.PluginDIR,
		pluginSet:      make(map[string]*pWrapper.PluginWrapper),
		doneCh:         make(chan struct{}),
		regularTicker:  time.NewTicker(TenSecsInterval),
		psTicker:       time.NewTicker(OneSecInterval),
	}

	os.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(os).String(),
		Output: o.Stdout,
		Level:  hclog.Debug,
	})

	// load plugin configs before start them.
	plugConfs, err := config.LoadPluginsConfig(conf.PluginConfFile)
	if err != nil {
		os.logger.Error("cannot load plugin configuration", "error", err.Error())
		panic(err)
	}

	// discover plugins from plugin dir at startup.
	binaries, err := helpers.ListPlugins(conf.PluginDIR)
	if len(binaries) == 0 || err != nil {
		// to stop the service on the start once there is no plugin in the db.
		os.logger.Error("No plugins at plugin dir", "plugin-dir", os.pluginDIR)
		panic(fmt.Sprintf("No plugins at plugin dir: %s, please build the plugins", os.pluginDIR))
	}
	for _, file := range binaries {
		f := file
		pConf := plugConfs[f.Name()]
		os.tryLoadingNewPlugin(f, pConf)
	}

	os.logger.Info("Running data contract_binder", "rpc: ", conf.AutonityWSUrl, "voter", conf.Key.Address.String())
	err = os.syncStates()
	if err != nil {
		// stop the client on start up once the remote endpoint of autonity L1 network is not ready.
		panic(err)
	}
	os.lostSync = false
	return os
}

func (os *OracleServer) syncStates() error {
	var err error
	// get initial states from on-chain oracle contract.
	os.curRound, os.protocolSymbols, os.pricePrecision, os.votePeriod, err = os.initStates()
	if err != nil {
		os.logger.Error("init state", "error", err.Error())
		return err
	}

	os.logger.Info("syncStates", "CurrentRound", os.curRound, "Num of AvailableSymbols", len(os.protocolSymbols), "CurrentSymbols", os.protocolSymbols)
	os.UpdateSymbols(os.protocolSymbols)

	// subscribe on-chain round rotation event
	os.chRoundEvent = make(chan *contract.OracleNewRound)
	os.subRoundEvent, err = os.oracleContract.WatchNewRound(new(bind.WatchOpts), os.chRoundEvent)
	if err != nil {
		os.logger.Error("WatchNewRound", "error", err.Error())
		return err
	}

	// subscribe on-chain symbol update event
	os.chSymbolsEvent = make(chan *contract.OracleNewSymbols)
	os.subSymbolsEvent, err = os.oracleContract.WatchNewSymbols(new(bind.WatchOpts), os.chSymbolsEvent)
	if err != nil {
		os.logger.Error("WatchNewSymbols", "error", err.Error())
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
		os.logger.Error("Get round", "error", err.Error())
		return 0, nil, precision, 0, err
	}

	symbols, err := os.oracleContract.GetSymbols(nil)
	if err != nil {
		os.logger.Error("Get symbols", "error", err.Error())
		return 0, nil, precision, 0, err
	}

	p, err := os.oracleContract.GetPrecision(nil)
	if err != nil {
		os.logger.Error("Get precision", "error", err.Error())
		return 0, nil, precision, 0, err
	}

	votePeriod, err := os.oracleContract.GetVotePeriod(nil)
	if err != nil {
		os.logger.Error("Get vote period", "error", err.Error())
		return 0, nil, precision, 0, nil
	}

	if len(symbols) == 0 {
		os.logger.Error("No symbols set by operator")
		return currentRound.Uint64(), symbols, decimal.NewFromInt(p.Int64()), votePeriod.Uint64(), types.ErrNoSymbolsObserved
	}

	return currentRound.Uint64(), symbols, decimal.NewFromInt(p.Int64()), votePeriod.Uint64(), nil
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
			os.logger.Info("rebuilding connectivity with autonity L1 node", "error", err)
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
	os.logger.Info("checking heart beat", "current height", h, "current round", r.Uint64())
}

func (os *OracleServer) isVoter() (bool, error) {
	voters, err := os.oracleContract.GetVoters(nil)
	if err != nil {
		os.logger.Error("Get voters", "error", err.Error())
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
			os.logger.Error("GetRoundData", "error", err.Error())
			return
		}

		os.logger.Info("GetRoundPrice", "round", newRound-1, "symbol", s, "Price",
			rd.Price.String(), "status", rd.Status.String())
	}

	for _, s := range os.protocolSymbols {
		rd, err := os.oracleContract.LatestRoundData(nil, s)
		if err != nil {
			os.logger.Error("GetLatestRoundPrice", "error", err.Error())
			return
		}

		price, err := decimal.NewFromString(rd.Price.String())
		if err != nil {
			continue
		}

		os.logger.Info("LatestRoundPrice", "round", rd.Round.Uint64(), "symbol", s, "price",
			price.Div(os.pricePrecision).String(), "status", rd.Status.String())
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
	os.logger.Debug("Data pre-sampling", "on height", curHeight, "TS", preSampleTS)
	os.samplePrice(os.symbols, preSampleTS)

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
		os.logger.Info("Cannot find last round's data, reports with commitment hash and no data")
	}

	// if node is no longer a validator, and it doesn't have last round data, skip reporting.
	if !isVoter && !ok {
		os.logger.Debug("skip reporting since client is no long a voter, and have no last round data.")
		return nil
	}

	if isVoter {
		// report with last round data and with current round commitment hash.
		return os.reportWithCommitment(os.curRound, lastRoundData)
	}

	// report with last round data but without current round commitment.
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
	os.logger.Info("report with commitment", "TX hash", curRoundData.Tx.Hash(), "Nonce", curRoundData.Tx.Nonce(), "Cost", curRoundData.Tx.Cost())

	// alert in case of balance reach the warning value.
	balance, err := os.client.BalanceAt(context.Background(), os.key.Address, nil)
	if err != nil {
		os.logger.Error("cannot get account balance", "error", err.Error())
		return err
	}

	os.logger.Info("oracle client left fund", "balance", balance.String())
	if balance.Cmp(AlertBalance) <= 0 {
		os.logger.Warn("oracle account has too less balance left for data reporting", "balance", balance.String())
	}

	return nil
}

// report with last round data but without current round commitment.
func (os *OracleServer) reportWithoutCommitment(lastRoundData *types.RoundData) error {

	tx, err := os.doReport(common.Hash{}, lastRoundData)
	if err != nil {
		os.logger.Error("do report", "error", err.Error())
		return err
	}
	os.logger.Info("report without commitment", "TX hash", tx.Hash(), "Nonce", tx.Nonce())
	return nil
}

func (os *OracleServer) UpdateSymbols(newSymbols []string) {
	var symbolsMap = make(map[string]struct{})
	for _, s := range os.symbols {
		symbolsMap[s] = struct{}{}
	}

	for _, newS := range newSymbols {
		if _, ok := symbolsMap[newS]; !ok {
			os.symbols = append(os.symbols, newS)
		}
	}
}

func (os *OracleServer) doReport(curRndCommitHash common.Hash, lastRoundData *types.RoundData) (*tp.Transaction, error) {
	chainID, err := os.client.ChainID(context.Background())
	if err != nil {
		os.logger.Error("get chain id", "error", err.Error())
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(os.key.PrivateKey, chainID)
	if err != nil {
		os.logger.Error("NewKeyedTransactorWithChainID", "error", err)
		return nil, err
	}

	auth.Value = big.NewInt(0)
	auth.GasTipCap = new(big.Int).SetUint64(os.gasTipCap)
	auth.GasLimit = uint64(3000000)

	// if there is no last round data, then we just submit the curRndCommitHash hash of current round.
	var votes []*big.Int
	if lastRoundData == nil {
		for i := 0; i < len(os.protocolSymbols); i++ {
			votes = append(votes, types.InvalidPrice)
		}
		return os.oracleContract.Vote(auth, new(big.Int).SetBytes(curRndCommitHash.Bytes()), votes, types.InvalidSalt)
	}

	for _, s := range lastRoundData.Symbols {
		_, ok := lastRoundData.Prices[s]
		if !ok {
			votes = append(votes, types.InvalidPrice)
		} else {
			price := lastRoundData.Prices[s].Price.Mul(os.pricePrecision).BigInt()
			votes = append(votes, price)
		}
	}

	return os.oracleContract.Vote(auth, new(big.Int).SetBytes(curRndCommitHash.Bytes()), votes, lastRoundData.Salt)
}

func (os *OracleServer) buildRoundData(round uint64) (*types.RoundData, error) {
	if len(os.protocolSymbols) == 0 {
		return nil, types.ErrNoSymbolsObserved
	}

	prices := make(types.PriceBySymbol)
	for _, s := range os.protocolSymbols {
		p, err := os.aggregatePrice(s, int64(os.curSampleTS))
		if err != nil {
			os.logger.Error("aggregatePrice", "error", err.Error(), "symbol", s)
			continue
		}
		prices[s] = *p
	}

	if len(prices) == 0 {
		return nil, types.ErrNoAvailablePrice
	}

	os.logger.Info("sampled prices", "round", round, "prices", prices)

	salt, err := rand.Int(rand.Reader, SaltRange)
	if err != nil {
		os.logger.Error("generate rand salt", "error", err.Error())
		return nil, err
	}

	var roundData = &types.RoundData{
		RoundID:        round,
		Symbols:        os.protocolSymbols,
		Salt:           salt,
		CommitmentHash: common.Hash{},
		Prices:         prices,
	}
	roundData.CommitmentHash = os.commitmentHash(roundData, os.protocolSymbols)
	os.logger.Info("build round data", "current round", round, "commitment hash", roundData.CommitmentHash.String())
	return roundData, nil
}

func (os *OracleServer) commitmentHash(roundData *types.RoundData, symbols []string) common.Hash {
	var source []byte
	for _, s := range symbols {
		if pr, ok := roundData.Prices[s]; ok {
			source = append(source, common.LeftPadBytes(pr.Price.Mul(os.pricePrecision).BigInt().Bytes(), 32)...)
		} else {
			source = append(source, common.LeftPadBytes(types.InvalidPrice.Bytes(), 32)...)
		}
	}
	// append the salt at the tail of votes
	source = append(source, common.LeftPadBytes(roundData.Salt.Bytes(), 32)...)
	// append the sender address at the tail for commitment hash computing as well
	source = append(source, os.key.Address.Bytes()...)
	return crypto.Keccak256Hash(source)
}

func (os *OracleServer) handleNewSymbolsEvent(symbols []string) {
	os.logger.Info("handleNewSymbolsEvent", "symbols", symbols)
	// just add symbols to oracle service's symbol pool, thus the oracle service can start to prepare the data.
	os.UpdateSymbols(symbols)
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
		return nil, types.ErrNoAvailablePrice
	}

	price := &types.Price{
		Timestamp: target,
		Price:     prices[0],
		Symbol:    s,
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
	nListener := os.sampleEventFee.Send(e)
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
		case <-os.psTicker.C:
			preSampleTS := time.Now().Unix()
			err := os.handlePreSampling(preSampleTS)
			if err != nil {
				os.logger.Error("handle pre-sampling", "error", err.Error())
			}
			os.lastSampledTS = preSampleTS
		case rEvent := <-os.chRoundEvent:
			os.logger.Info("handle new round", "round", rEvent.Round.Uint64(), "required sampling TS",
				rEvent.Timestamp.Uint64(), "height", rEvent.Height.Uint64(), "vote period", rEvent.VotePeriod.Uint64())

			// save the round rotation info to coordinate the pre-sampling.
			os.curRound = rEvent.Round.Uint64()
			os.votePeriod = rEvent.VotePeriod.Uint64()
			os.curSampleHeight = rEvent.Height.Uint64()
			os.curSampleTS = rEvent.Timestamp.Uint64()

			err := os.handleRoundVote()
			if err != nil {
				os.logger.Error("Handling round vote", "err", err.Error())
				continue
			}
			os.gcDataSamples()
			// after vote finished, gc useless symbols by protocol required symbols.
			os.symbols = os.protocolSymbols
		case symbols := <-os.chSymbolsEvent:
			os.logger.Info("handle new symbols", "symbols", symbols.Symbols, "activate at round", symbols.Round)
			os.handleNewSymbolsEvent(symbols.Symbols)
		case <-os.regularTicker.C:
			// start the regular price sampling for oracle service on each 10s.
			now := time.Now().Unix()
			os.logger.Debug("regular 10s data sampling", "ts", now)
			os.samplePrice(os.symbols, now)
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
		os.logger.Error("PluginRuntimeDiscovery", "error", err.Error())
		return
	}
	for _, file := range binaries {
		f := file
		pConf := plugConfs[f.Name()]
		os.tryLoadingNewPlugin(f, pConf)
	}
}

func (os *OracleServer) tryLoadingNewPlugin(f fs.FileInfo, plugConf types.PluginConfig) {
	plugin, ok := os.pluginSet[f.Name()]
	if !ok {
		os.logger.Info("** New plugin discovered, going to setup it: ", f.Name(), f.Mode().String())

		if err := os.ApplyPluginConf(f.Name(), plugConf); err != nil {
			os.logger.Error("Apply plugin config", "error", err.Error())
			return
		}

		pluginWrapper := pWrapper.NewPluginWrapper(f.Name(), os.pluginDIR, os)
		if err := pluginWrapper.Initialize(); err != nil {
			os.logger.Error("** Cannot initialize plugin", "name", f.Name(), "error", err.Error())
			return
		}
		os.pluginSet[f.Name()] = pluginWrapper
		os.logger.Info("** New plugin on ready: ", f.Name())
		return
	}

	if f.ModTime().After(plugin.StartTime()) || plugin.Exited() {
		if err := os.ApplyPluginConf(f.Name(), plugConf); err != nil {
			os.logger.Error("Apply plugin config", "error", err.Error())
			return
		}

		// stop the legacy plugin
		plugin.Close()
		delete(os.pluginSet, f.Name())

		os.logger.Info("*** Replacing legacy plugin with new one: ", f.Name(), f.Mode().String())
		pluginWrapper := pWrapper.NewPluginWrapper(f.Name(), os.pluginDIR, os)
		if err := pluginWrapper.Initialize(); err != nil {
			os.logger.Error("** Cannot initialize plugin", "name", f.Name(), "error", err.Error())
			return
		}
		os.pluginSet[f.Name()] = pluginWrapper
		os.logger.Info("*** Finnish the replacement of plugin: ", f.Name())
	}
}

func (os *OracleServer) WatchSampleEvent(sink chan<- *types.SampleEvent) event.Subscription {
	return os.sampleEventFee.Subscribe(sink)
}

func (os *OracleServer) ApplyPluginConf(name string, plugConf types.PluginConfig) error {
	// set the plugin configuration via system env, thus the plugin can load it on startup.
	conf, err := json.Marshal(plugConf)
	if err != nil {
		os.logger.Error("** Cannot marshal plugin's configuration", "error", err.Error())
		return err
	}
	if err = o.Setenv(name, string(conf)); err != nil {
		os.logger.Error("** Cannot set plugin configuration via system ENV")
		return err
	}
	return nil
}
