// Package chain_adaptor implement the client connect to autonity client via web socket, it also implements a data
// reporting module, which not only construct the oracle contract binder on the start, and discovery the latest symbols from the
// oracle contract for oracle service, but also subscribe the chain head event, create event handler routine to coordinate
// the oracle data reporting workflows as well.
package chain_adaptor

import (
	contract "autonity-oracle/chain_adaptor/contract"
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
var PricePrecision = decimal.RequireFromString("1000000000")

var ErrPeerOnSync = errors.New("l1 node is on peer sync")
var HealthCheckerInterval = 2 * time.Minute // ws connectivity checker interval.

const MaxBufferedRounds = 10

type DataReporter struct {
	logger           hclog.Logger
	oracleContract   *contract.Oracle
	client           *ethclient.Client
	autonityWSUrl    string
	currentRound     uint64
	currentSymbols   []string
	roundData        map[uint64]*types.RoundData
	key              *keystore.Key
	validatorAccount common.Address
	oracleService    types.OracleService
	chRoundEvent     chan *contract.OracleUpdatedRound
	subRoundEvent    event.Subscription
	chSymbolsEvent   chan *contract.OracleUpdatedSymbols
	subSymbolsEvent  event.Subscription
	liveTicker       *time.Ticker
}

func NewDataReporter(ws string, key *keystore.Key, validatorAccount common.Address, oracleService types.OracleService) *DataReporter {
	dp := &DataReporter{
		validatorAccount: validatorAccount,
		autonityWSUrl:    ws,
		roundData:        make(map[uint64]*types.RoundData),
		key:              key,
		oracleService:    oracleService,
		liveTicker:       time.NewTicker(HealthCheckerInterval),
	}

	err := dp.buildConnection()
	if err != nil {
		// stop the client on start up once the remote endpoint of autonity L1 network is not ready.
		panic(err)
	}

	dp.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(dp).String(),
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	dp.logger.Info("Running data reporter", "rpc: ", ws)
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
	oc, err := contract.NewOracle(ContractAddress, dp.client)
	if err != nil {
		return err
	}

	// get initial states from on-chain oracle contract.
	dp.currentRound, dp.currentSymbols, err = getStartingStates(oc)
	if err != nil {
		return err
	}

	if len(dp.currentSymbols) > 0 {
		dp.oracleService.UpdateSymbols(dp.currentSymbols)
	}

	// subscribe on-chain round rotation event
	dp.chRoundEvent = make(chan *contract.OracleUpdatedRound)
	dp.subRoundEvent, err = oc.WatchUpdatedRound(new(bind.WatchOpts), dp.chRoundEvent)
	if err != nil {
		return err
	}

	// subscribe on-chain symbol update event
	dp.chSymbolsEvent = make(chan *contract.OracleUpdatedSymbols)
	dp.subSymbolsEvent, err = oc.WatchUpdatedSymbols(new(bind.WatchOpts), dp.chSymbolsEvent)
	if err != nil {
		return err
	}
	return nil
}

// getStartingStates returns round id, symbols and committees on current chain, it is called on the startup of client.
func getStartingStates(oc *contract.Oracle) (uint64, []string, error) {
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
			err := dp.handleRoundChange(round.Round.Uint64())
			if err != nil {
				dp.logger.Error("Handling round change event", "err", err.Error())
			}
		case symbols := <-dp.chSymbolsEvent:
			dp.handleNewSymbols(symbols.Symbols)
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
			if k < offset {
				delete(dp.roundData, k)
			}
		}
	}
}

// get the latest chain height at the L1 autonity node, if it encounters any failure, do the connection rebuilding.
func (dp *DataReporter) checkHealth() {

	height, err := dp.client.BlockNumber(context.Background())
	if err == nil {
		dp.logger.Info("L1 autonity client health check is okay!", "height", height)
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
	return
}

func (dp *DataReporter) isCommitteeMember() (bool, error) {
	committee, err := dp.oracleContract.GetCommittee(nil)
	if err != nil {
		return false, err
	}

	for _, c := range committee {
		if c == dp.validatorAccount {
			return true, nil
		}
	}
	return false, nil
}

func (dp *DataReporter) handleRoundChange(newRound uint64) error {
	dp.currentRound = newRound

	// if the autonity node is on peer synchronization state, just skip the reporting.
	sync, err := dp.client.SyncProgress(context.Background())
	if err != nil {
		return err
	}
	if sync != nil {
		return ErrPeerOnSync
	}

	// if client is not a committee member, just skip reporting.
	isCommittee, err := dp.isCommitteeMember()
	if err != nil {
		return err
	}

	if !isCommittee {
		return nil
	}

	symbols, err := dp.oracleContract.GetSymbols(nil)
	if err != nil {
		return err
	}

	//todo, if there is prices not available, shall we wait for a while(10s) until we get the all the prices to be ready.
	prices := dp.oracleService.GetPricesBySymbols(symbols)
	curRoundData := dp.computeCommitment(prices)

	// query last round's prices, its random salt which will reveal last round's report.
	lastRoundData, ok := dp.roundData[newRound-1]
	if !ok {
		dp.logger.Info("Cannot find last round's data")
	}

	// prepare the transaction which carry current round's commitment, and last round's data.
	curRoundData.Tx, err = dp.doReport(curRoundData.Hash, lastRoundData, symbols)
	if err != nil {
		return err
	}

	// save current round's commitment, prices and the random salt.
	dp.roundData[newRound] = curRoundData
	return nil
}

func (dp *DataReporter) doReport(commitment common.Hash, lastRoundData *types.RoundData, symbols []string) (*tp.Transaction, error) {
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

	noPrice := big.NewInt(0)
	var votes []*big.Int
	for _, s := range symbols {
		_, ok := lastRoundData.Prices[s]
		if !ok {
			votes = append(votes, noPrice)
		} else {
			price := lastRoundData.Prices[s].Price.Mul(PricePrecision).BigInt()
			votes = append(votes, price)
		}
	}

	// append the salt at the end of the slice, to reveal the report.
	votes = append(votes, lastRoundData.Salt)

	tx, err := dp.oracleContract.Vote(auth, new(big.Int).SetBytes(commitment.Bytes()), votes)
	if err != nil {
		return tx, err
	}

	return tx, nil
}

func (dp *DataReporter) computeCommitment(prices types.PriceBySymbol) *types.RoundData {
	var roundData = &types.RoundData{
		Salt:   new(big.Int).SetUint64(rand.Uint64()),
		Hash:   common.Hash{},
		Prices: prices,
	}

	salt := roundData.Salt
	sum := salt

	for _, p := range roundData.Prices {
		pr := p
		sum.Add(sum, pr.Price.Mul(PricePrecision).BigInt())
	}

	roundData.Hash = crypto.Keccak256Hash(sum.FillBytes(make([]byte, 32)))
	return roundData
}

func (dp *DataReporter) handleNewSymbols(symbols []string) {
	dp.logger.Info("handleNewSymbols", "symbols", symbols)
	dp.currentSymbols = symbols
	dp.oracleService.UpdateSymbols(symbols)
}

func (dp *DataReporter) Stop() {
	dp.client.Close()
	dp.subRoundEvent.Unsubscribe()
	dp.subSymbolsEvent.Unsubscribe()
	dp.liveTicker.Stop()
}
