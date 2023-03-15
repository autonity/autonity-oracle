// Package reporter implement the client connect to autonity client via web socket, it also implements a data
// reporting module, which not only construct the oracle contract binder on the start, and discovery the latest symbols from the
// oracle contract for oracle service, but also subscribe the chain head event, create event handler routine to coordinate
// the oracle data reporting workflows as well.
package reporter

import (
	contract "autonity-oracle/reporter/contract"
	"autonity-oracle/types"
	"context"
	"errors"
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
	"os"
	"time"
)

var Deployer = common.Address{}
var ContractAddress = crypto.CreateAddress(Deployer, 1)
var PricePrecision = decimal.RequireFromString("10000000")

var ErrPeerOnSync = errors.New("l1 node is on peer sync")
var ErrNoAvailablePrice = errors.New("no available prices collected yet")
var HealthCheckerInterval = 2 * time.Minute // ws connectivity checker interval.

const MaxBufferedRounds = 10

type DataReporter struct {
	logger          hclog.Logger
	oracleContract  contract.ContractAPI
	client          *ethclient.Client
	autonityWSUrl   string
	currentRound    uint64
	currentSymbols  []string
	roundData       map[uint64]*types.RoundData
	key             *keystore.Key
	oracleService   types.OracleService
	chRoundEvent    chan *contract.OracleNewRound
	subRoundEvent   event.Subscription
	chSymbolsEvent  chan *contract.OracleNewSymbols
	subSymbolsEvent event.Subscription
	liveTicker      *time.Ticker
}

func NewDataReporter(ws string, key *keystore.Key, oracleService types.OracleService) *DataReporter {
	dp := &DataReporter{
		autonityWSUrl: ws,
		roundData:     make(map[uint64]*types.RoundData),
		key:           key,
		oracleService: oracleService,
		liveTicker:    time.NewTicker(HealthCheckerInterval),
	}
	dp.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(dp).String(),
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	dp.logger.Info("Running data reporter", "rpc: ", ws, "voter", key.Address.String())
	err := dp.buildConnection()
	if err != nil {
		// stop the client on start up once the remote endpoint of autonity L1 network is not ready.
		panic(err)
	}
	return dp
}

func (dp *DataReporter) buildConnection() error {
	// connect to autonity node via web socket
	var err error
	dp.client, err = ethclient.Dial(dp.autonityWSUrl)
	if err != nil {
		return err
	}

	// bind client with oracle contract address
	dp.logger.Info("binding with oracle contract", "address", ContractAddress.String())
	dp.oracleContract, err = contract.NewOracle(ContractAddress, dp.client)
	if err != nil {
		return err
	}

	// get initial states from on-chain oracle contract.
	dp.currentRound, dp.currentSymbols, err = getStartingStates(dp.oracleContract)
	if err != nil {
		return err
	}

	if len(dp.currentSymbols) > 0 {
		dp.oracleService.UpdateSymbols(dp.currentSymbols)
	}

	// subscribe on-chain round rotation event
	dp.chRoundEvent = make(chan *contract.OracleNewRound)
	dp.subRoundEvent, err = dp.oracleContract.WatchNewRound(new(bind.WatchOpts), dp.chRoundEvent)
	if err != nil {
		return err
	}

	// subscribe on-chain symbol update event
	dp.chSymbolsEvent = make(chan *contract.OracleNewSymbols)
	dp.subSymbolsEvent, err = dp.oracleContract.WatchNewSymbols(new(bind.WatchOpts), dp.chSymbolsEvent)
	if err != nil {
		return err
	}
	return nil
}

// getStartingStates returns round id, symbols and committees on current chain, it is called on the startup of client.
func getStartingStates(oc contract.ContractAPI) (uint64, []string, error) {
	// on the startup, we need to sync the round id, symbols and committees from contract.
	currentRound, err := oc.GetRound(nil)
	if err != nil {
		return 0, nil, err
	}

	symbols, err := oc.GetSymbols(nil)
	if err != nil {
		return 0, nil, err
	}

	return currentRound.Uint64(), symbols, nil
}

// Start starts the event loop to handle the on-chain events, we have 3 events to be processed.
func (dp *DataReporter) Start() {
	for {
		select {
		case err := <-dp.subSymbolsEvent.Err():
			dp.logger.Info("reporter routine is shutting down ", err)
		case err := <-dp.subRoundEvent.Err():
			dp.logger.Info("reporter routine is shutting down ", err)
		case round := <-dp.chRoundEvent:
			dp.logger.Info("handle new round", "round", round.Round.Uint64())
			err := dp.handleRoundChangeEvent(round.Round.Uint64())
			if err != nil {
				dp.logger.Error("Handling round change event", "err", err.Error())
			}
		case symbols := <-dp.chSymbolsEvent:
			dp.handleNewSymbolsEvent(symbols.Symbols)
		case <-dp.liveTicker.C:
			dp.checkHealth()
			dp.gcRoundData()
		}
	}
}

func (dp *DataReporter) gcRoundData() {
	if len(dp.roundData) >= MaxBufferedRounds {
		offset := dp.currentRound - MaxBufferedRounds
		for k := range dp.roundData {
			if k <= offset {
				delete(dp.roundData, k)
			}
		}
	}
}

func (dp *DataReporter) checkHealth() {

	r, err := dp.oracleContract.GetRound(nil)
	if err == nil {
		dp.logger.Warn("checking heart beat", "current voting round", r.Uint64())
		return
	}

	// release the legacy resources if the connectivity was lost.
	dp.client.Close()
	dp.subRoundEvent.Unsubscribe()
	dp.subSymbolsEvent.Unsubscribe()

	// rebuild the connection with autonity L1 node.
	err = dp.buildConnection()
	if err != nil {
		dp.logger.Info("rebuilding connectivity with autonity L1 node", "error", err)
	}
}

func (dp *DataReporter) isVoter() (bool, error) {
	voters, err := dp.oracleContract.GetVoters(nil)
	if err != nil {
		return false, err
	}

	for _, c := range voters {
		if c == dp.key.Address {
			return true, nil
		}
	}
	return false, nil
}

func (dp *DataReporter) handleRoundChangeEvent(newRound uint64) error {
	dp.currentRound = newRound

	// if the autonity node is on peer synchronization state, just skip the reporting.
	sync, err := dp.client.SyncProgress(context.Background())
	if err != nil {
		return err
	}
	if sync != nil {
		return ErrPeerOnSync
	}

	// if client is not a voter, just skip reporting.
	isVoter, err := dp.isVoter()
	if err != nil {
		return err
	}

	// query last round's prices, its random salt which will reveal last round's report.
	lastRoundData, ok := dp.roundData[newRound-1]
	if !ok {
		dp.logger.Info("Cannot find last round's data, oracle will just report current round commitment hash.")
	}

	// if node is no longer a validator, and it doesn't have last round data, skip reporting.
	if !isVoter && !ok {
		dp.logger.Debug("skip reporting since client is no long a voter, and have no last round data.")
		return nil
	}

	if isVoter {
		// report with last round data and with current round commitment hash.
		return dp.reportWithCommitment(newRound, lastRoundData)
	}

	// report with last round data but without current round commitment.
	return dp.reportWithoutCommitment(lastRoundData)
}

func (dp *DataReporter) reportWithCommitment(newRound uint64, lastRoundData *types.RoundData) error {
	curRoundData, err := dp.buildRoundData()
	if err != nil {
		return err
	}

	// prepare the transaction which carry current round's commitment, and last round's data.
	curRoundData.Tx, err = dp.doReport(curRoundData.Hash, lastRoundData)
	if err != nil {
		return err
	}

	// save current round data.
	dp.roundData[newRound] = curRoundData
	dp.logger.Info("report with commitment", "TX hash", curRoundData.Tx.Hash(), "Nonce", curRoundData.Tx.Nonce())
	return nil
}

// report with last round data but without current round commitment.
func (dp *DataReporter) reportWithoutCommitment(lastRoundData *types.RoundData) error {

	tx, err := dp.doReport(common.Hash{}, lastRoundData)
	if err != nil {
		return err
	}
	dp.logger.Info("report with commitment", "TX hash", tx.Hash(), "Nonce", tx.Nonce())
	return nil
}

func (dp *DataReporter) doReport(curRndCommitHash common.Hash, lastRoundData *types.RoundData) (*tp.Transaction, error) {
	from := dp.key.Address

	nonce, err := dp.client.PendingNonceAt(context.Background(), from)
	if err != nil {
		return nil, err
	}

	gasPrice, err := dp.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	chainID, err := dp.client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(dp.key.PrivateKey, chainID)
	if err != nil {
		return nil, err
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	// if there is no last round data, then we just submit the curRndCommitHash hash of current round.
	if lastRoundData == nil {
		var votes []*big.Int
		return dp.oracleContract.Vote(auth, new(big.Int).SetBytes(curRndCommitHash.Bytes()), votes, types.InvalidPrice)
	}

	var votes []*big.Int
	for _, s := range lastRoundData.Symbols {
		_, ok := lastRoundData.Prices[s]
		if !ok {
			votes = append(votes, types.InvalidPrice)
		} else {
			price := lastRoundData.Prices[s].Price.Mul(PricePrecision).BigInt()
			votes = append(votes, price)
		}
	}

	return dp.oracleContract.Vote(auth, new(big.Int).SetBytes(curRndCommitHash.Bytes()), votes, lastRoundData.Salt)
}

func (dp *DataReporter) buildRoundData() (*types.RoundData, error) {
	// get symbols of the latest round.
	symbols, err := dp.oracleContract.GetSymbols(nil)
	if err != nil {
		return nil, err
	}

	prices := dp.oracleService.GetPricesBySymbols(symbols)
	if len(prices) == 0 {
		return nil, ErrNoAvailablePrice
	}

	seed := time.Now().UnixNano()
	var roundData = &types.RoundData{
		Symbols: symbols,
		Salt:    new(big.Int).SetUint64(rand.New(rand.NewSource(seed)).Uint64()), // nolint
		Hash:    common.Hash{},
		Prices:  prices,
	}

	var source []byte
	for _, s := range symbols {
		if pr, ok := roundData.Prices[s]; ok {
			source = append(source, common.LeftPadBytes(pr.Price.Mul(PricePrecision).BigInt().Bytes(), 32)...)
		} else {
			source = append(source, common.LeftPadBytes(types.InvalidPrice.Bytes(), 32)...)
		}
	}
	// append the salt at the tail of votes
	source = append(source, common.LeftPadBytes(roundData.Salt.Bytes(), 32)...)
	roundData.Hash = crypto.Keccak256Hash(source)
	return roundData, nil
}

func (dp *DataReporter) handleNewSymbolsEvent(symbols []string) {
	dp.logger.Info("handleNewSymbolsEvent", "symbols", symbols)
	dp.currentSymbols = symbols
	dp.oracleService.UpdateSymbols(symbols)
}

func (dp *DataReporter) Stop() {
	dp.client.Close()
	dp.subRoundEvent.Unsubscribe()
	dp.subSymbolsEvent.Unsubscribe()
	dp.liveTicker.Stop()
}
