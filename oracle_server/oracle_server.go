package oracleserver

import (
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/helpers"
	pWrapper "autonity-oracle/plugin_wrapper"
	"autonity-oracle/types"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	tp "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	"github.com/shopspring/decimal"
	"math/big"
	"math/rand"
	o "os"
	"sync"
	"time"
)

var (
	ValidDataAge     = 30 // 30 seconds, data fetched within 30s are valid to update the price.
	Version          = "v0.0.1"
	TenSecsInterval  = 10 * time.Second // 10s ticker job to check health with l1, plugin discovery and regular data sampling.
	OneSecInterval   = 1 * time.Second  // 1s ticker job to check if we need to do pre-sampling.
	PreSamplingRange = 15               // pre-sampling starts in 15blocks in advance.
	Timeout          = 5 * time.Second  // the timeout for fetching prices from plugin.
)

type OracleServer struct {
	logger        hclog.Logger
	doneCh        chan struct{}
	regularTicker *time.Ticker // the clock source to trigger the 10s interval job.
	psTicker      *time.Ticker // the pre-sampling ticker in 1s.

	pluginDIR  string                             // the dir saves the plugins.
	pluginLock sync.RWMutex                       // to prevent from race condition of pluginSet.
	pluginSet  map[string]*pWrapper.PluginWrapper // the plugin clients that connect with different adapters.

	symbols []string // the symbols for data fetching in oracle service.

	// the reporting staffs
	oracleContract contract.ContractAPI
	client         *ethclient.Client
	l1WSUrl        string

	curRound        uint64 //round ID.
	votePeriod      uint64 //vote period.
	curSampleTS     uint64 //the data sample TS of the current round.
	curSampleHeight uint64 //The block height on which the round rotation happens.

	currentSymbols []string
	pricePrecision decimal.Decimal
	roundData      map[uint64]*types.RoundData
	key            *keystore.Key

	chRoundEvent  chan *contract.OracleNewRound
	subRoundEvent event.Subscription

	chSymbolsEvent  chan *contract.OracleNewSymbols
	subSymbolsEvent event.Subscription
}

func NewOracleServer(symbols []string, pluginDir string, ws string, key *keystore.Key) *OracleServer {
	os := &OracleServer{
		l1WSUrl:       ws,
		roundData:     make(map[uint64]*types.RoundData),
		key:           key,
		symbols:       symbols,
		pluginDIR:     pluginDir,
		pluginSet:     make(map[string]*pWrapper.PluginWrapper),
		doneCh:        make(chan struct{}),
		regularTicker: time.NewTicker(TenSecsInterval),
		psTicker:      time.NewTicker(OneSecInterval),
	}

	os.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(os).String(),
		Output: o.Stdout,
		Level:  hclog.Debug,
	})

	// discover plugins from plugin dir at startup.
	binaries := os.listPluginDIR()
	if len(binaries) == 0 {
		// to stop the service on the start once there is no plugin in the db.
		panic(fmt.Sprintf("No plugins at plugin dir: %s, please build the plugins", os.pluginDIR))
	}
	for _, file := range binaries {
		os.createPlugin(file)
	}

	os.logger.Info("Running data contract_binder", "rpc: ", ws, "voter", key.Address.String())
	err := os.buildConnection()
	if err != nil {
		// stop the client on start up once the remote endpoint of autonity L1 network is not ready.
		panic(err)
	}
	return os
}

func (os *OracleServer) buildConnection() error {
	// connect to autonity node via web socket
	var err error
	os.client, err = ethclient.Dial(os.l1WSUrl)
	if err != nil {
		return err
	}

	// bind client with oracle contract address
	os.logger.Info("binding with oracle contract", "address", types.OracleContractAddress.String())
	os.oracleContract, err = contract.NewOracle(types.OracleContractAddress, os.client)
	if err != nil {
		return err
	}

	// get initial states from on-chain oracle contract.
	os.curRound, os.currentSymbols, os.pricePrecision, os.votePeriod, err = initStates(os.oracleContract)
	if err != nil {
		return err
	}

	os.logger.Info("buildConnection", "CurrentRound", os.curRound, "Num of Symbols", len(os.currentSymbols), "CurrentSymbols", os.currentSymbols)
	if len(os.currentSymbols) > 0 {
		os.UpdateSymbols(os.currentSymbols)
	}

	// subscribe on-chain round rotation event
	os.chRoundEvent = make(chan *contract.OracleNewRound)
	os.subRoundEvent, err = os.oracleContract.WatchNewRound(new(bind.WatchOpts), os.chRoundEvent)
	if err != nil {
		return err
	}

	// subscribe on-chain symbol update event
	os.chSymbolsEvent = make(chan *contract.OracleNewSymbols)
	os.subSymbolsEvent, err = os.oracleContract.WatchNewSymbols(new(bind.WatchOpts), os.chSymbolsEvent)
	if err != nil {
		return err
	}

	return nil
}

// initStates returns round id, symbols and committees on current chain, it is called on the startup of client.
func initStates(oc contract.ContractAPI) (uint64, []string, decimal.Decimal, uint64, error) {
	var precision decimal.Decimal
	// on the startup, we need to sync the round id, symbols and committees from contract.
	currentRound, err := oc.GetRound(nil)
	if err != nil {
		return 0, nil, precision, 0, err
	}

	symbols, err := oc.GetSymbols(nil)
	if err != nil {
		return 0, nil, precision, 0, err
	}

	p, err := oc.GetPrecision(nil)
	if err != nil {
		return 0, nil, precision, 0, err
	}

	votePeriod, err := oc.GetVotePeriod(nil)
	if err != nil {
		return 0, nil, precision, 0, nil
	}

	return currentRound.Uint64(), symbols, decimal.NewFromInt(p.Int64()), votePeriod.Uint64(), nil
}

func (os *OracleServer) gcDataSamples() {
	os.pluginLock.RLock()
	defer os.pluginLock.RUnlock()
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
	if os.client != nil {
		os.client.Close()
		os.client = nil
	}
	os.subRoundEvent.Unsubscribe()
	os.subSymbolsEvent.Unsubscribe()
}

func (os *OracleServer) checkHealth() {
	// if the web socket was drops my remote peer, the client will be reset into nil.
	if os.client == nil {
		// rebuild the connection with autonity L1 node.
		err := os.buildConnection()
		if err != nil {
			os.logger.Info("rebuilding connectivity with autonity L1 node", "error", err)
		}
		return
	}

	h, err := os.client.BlockNumber(context.Background())
	if err != nil {
		return
	}

	r, err := os.oracleContract.GetRound(nil)
	if err != nil {
		return
	}
	os.logger.Info("checking heart beat", "current height", h, "current round", r.Uint64())
}

func (os *OracleServer) isVoter() (bool, error) {
	voters, err := os.oracleContract.GetVoters(nil)
	if err != nil {
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
	for _, s := range os.currentSymbols {
		rd, err := os.oracleContract.GetRoundData(nil, new(big.Int).SetUint64(newRound-1), s)
		if err != nil {
			os.logger.Error("GetRoundData", "error", err.Error())
		}

		os.logger.Info("GetRoundPrice", "round", newRound-1, "symbol", s, "Price",
			rd.Price.String(), "status", rd.Status.String())
	}

	for _, s := range os.currentSymbols {
		rd, err := os.oracleContract.LatestRoundData(nil, s)
		if err != nil {
			os.logger.Error("GetLatestRoundPrice", "error", err.Error())
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
		os.logger.Info("skip pre-sampling on start up round")
		return nil
	}

	// if it is not a good timing to start sampling then return.
	nSampleHeight := os.curSampleHeight + os.votePeriod
	curHeight, err := os.client.BlockNumber(context.Background())
	if err != nil {
		return err
	}
	if nSampleHeight-curHeight > uint64(PreSamplingRange) {
		os.logger.Info("pre-sampling window does not approach yet")
		return nil
	}

	// do the data pre-sampling.
	os.samplePrice(os.symbols, preSampleTS)

	return nil
}

func (os *OracleServer) handleRoundVote() error {
	// if the autonity node is on peer synchronization state, just skip the reporting.
	syncing, err := os.client.SyncProgress(context.Background())
	if err != nil {
		return err
	}

	if syncing != nil {
		return types.ErrPeerOnSync
	}

	// get latest symbols from oracle.
	os.currentSymbols, err = os.oracleContract.GetSymbols(nil)
	if err != nil {
		return err
	}

	os.printLatestRoundData(os.curRound)

	// if client is not a voter, just skip reporting.
	isVoter, err := os.isVoter()
	if err != nil {
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
		return err
	}

	// prepare the transaction which carry current round's commitment, and last round's data.
	curRoundData.Tx, err = os.doReport(curRoundData.Hash, lastRoundData)
	if err != nil {
		return err
	}

	// save current round data.
	os.roundData[newRound] = curRoundData
	os.logger.Info("report with commitment", "TX hash", curRoundData.Tx.Hash(), "Nonce", curRoundData.Tx.Nonce())
	return nil
}

// report with last round data but without current round commitment.
func (os *OracleServer) reportWithoutCommitment(lastRoundData *types.RoundData) error {

	tx, err := os.doReport(common.Hash{}, lastRoundData)
	if err != nil {
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
	from := os.key.Address

	nonce, err := os.client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return nil, err
	}

	gasPrice, err := os.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	chainID, err := os.client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(os.key.PrivateKey, chainID)
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	// if there is no last round data, then we just submit the curRndCommitHash hash of current round.
	var votes []*big.Int
	if lastRoundData == nil {
		for i := 0; i < len(os.currentSymbols); i++ {
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
	// get symbols of the latest round.
	symbols, err := os.oracleContract.GetSymbols(nil)
	if err != nil {
		return nil, err
	}

	if len(symbols) == 0 {
		return nil, types.ErrNoSymbolsObserved
	}

	var prices types.PriceBySymbol
	for _, s := range symbols {
		p, err := os.aggregatePrice(s, int64(os.curSampleTS))
		if err != nil {
			os.logger.Warn("get sample", "error", err.Error())
			continue
		}
		prices[s] = *p
	}

	if len(prices) == 0 {
		return nil, types.ErrNoAvailablePrice
	}

	os.logger.Info("sampled prices", "round", round, "prices", prices)

	seed := time.Now().UnixNano()
	var roundData = &types.RoundData{
		RoundID: round,
		Symbols: symbols,
		Salt:    new(big.Int).SetUint64(rand.New(rand.NewSource(seed)).Uint64()), // nolint
		Hash:    common.Hash{},
		Prices:  prices,
	}

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
	roundData.Hash = crypto.Keccak256Hash(source)
	os.logger.Info("build round data", "current round", round, "commitment hash", roundData.Hash.String())
	return roundData, nil
}

func (os *OracleServer) handleNewSymbolsEvent(symbols []string) {
	os.logger.Info("handleNewSymbolsEvent", "symbols", symbols)
	// just add symbols to oracle service's symbol pool, thus the oracle service can start to prepare the data.
	os.UpdateSymbols(symbols)
}

func (os *OracleServer) aggregatePrice(s string, target int64) (*types.Price, error) {
	var prices []decimal.Decimal

	os.pluginLock.RLock()
	os.pluginLock.RUnlock()
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
	os.pluginLock.RLock()
	os.pluginLock.RUnlock()
	for _, p := range os.pluginSet {
		plugin := p
		go func() {
			err := plugin.FetchPrices(symbols, ts)
			if err != nil {
				os.logger.Warn("FetchPrices routine error", "error", err.Error())
			}
			os.logger.Debug("FetchPrices routine done successfully")
		}()
	}
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
			}
		case err := <-os.subRoundEvent.Err():
			if err != nil {
				os.logger.Info("subscription error of new rEvent event", err)
				os.handleConnectivityError()
			}
		case <-os.psTicker.C:
			preSampleTS := time.Now().Unix()
			err := os.handlePreSampling(preSampleTS)
			if err != nil {
				os.logger.Warn("handle pre-sampling", "error", err.Error())
			}
		case rEvent := <-os.chRoundEvent:
			os.logger.Info("handle new round", "round", rEvent.Round.Uint64(), "current sampling TS",
				rEvent.Timestamp.Uint64(), "vote period", rEvent.VotePeriod.Uint64())

			// save the round rotation info to coordinate the pre-sampling.
			os.curRound = rEvent.Round.Uint64()
			os.votePeriod = rEvent.VotePeriod.Uint64()
			os.curSampleHeight = rEvent.Height.Uint64()
			os.curSampleTS = rEvent.Timestamp.Uint64()

			err := os.handleRoundVote()
			if err != nil {
				os.logger.Warn("Handling round vote", "err", err.Error())
			}
			os.gcDataSamples()
		case symbols := <-os.chSymbolsEvent:
			os.logger.Info("handle new symbols", "symbols", symbols.Symbols, "activated at rEvent", symbols.Round)
			os.handleNewSymbolsEvent(symbols.Symbols)
		case <-os.regularTicker.C:
			// start the regular price sampling for oracle service on each 10s.
			os.samplePrice(os.symbols, time.Now().Unix())
			os.checkHealth()
			os.PluginRuntimeDiscovery()
			os.gcRoundData()
		}
	}
}

func (os *OracleServer) Stop() {
	if os.client != nil {
		os.client.Close()
	}
	os.subRoundEvent.Unsubscribe()
	os.subSymbolsEvent.Unsubscribe()

	os.doneCh <- struct{}{}
	os.pluginLock.RLock()
	defer os.pluginLock.RUnlock()
	for _, c := range os.pluginSet {
		p := c
		p.Close()
	}
}
