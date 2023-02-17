// Package chain_adaptor implement the client connect to autonity client via web socket, it also implements a data
// reporting module, which not only construct the oracle contract binder on the start, and discovery the latest symbols from the
// oracle contract for oracle service, but also subscribe the chain head event, create event handler routine to coordinate
// the oracle data reporting workflows as well.
package chain_adaptor

import (
	contract "autonity-oracle/chain_adaptor/contract"
	"autonity-oracle/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	"github.com/shopspring/decimal"
	"math/big"
	"math/rand"
	"os"
)

var Deployer = common.Address{}
var ContractAddress = crypto.CreateAddress(Deployer, 1)
var PricePrecision = decimal.RequireFromString("1000000000")

type DataReporter struct {
	logger           hclog.Logger
	oracleContract   *contract.Oracle
	client           *ethclient.Client
	autonityWSUrl    string
	currentRound     uint64
	roundData        map[uint64]*types.RoundData
	key              *keystore.Key
	validatorAccount common.Address
	oracleService    types.OracleService
	chRoundEvent     chan *contract.OracleUpdatedRound
	subRoundEvent    event.Subscription
	chSymbolsEvent   chan *contract.OracleUpdatedSymbols
	subSymbolsEvent  event.Subscription
}

func NewDataReporter(ws string, key *keystore.Key, validatorAccount common.Address, symbolsWriter types.OracleService) *DataReporter {
	// connect to autonity node via web socket
	client, err := ethclient.Dial(ws)
	if err != nil {
		panic(err)
	}

	// bind client with oracle contract address
	oc, err := contract.NewOracle(ContractAddress, client)
	if err != nil {
		panic(err)
	}

	// get initial states from on-chain oracle contract.
	curRound, curSymbols := getStartingStates(oc)
	if len(curSymbols) > 0 {
		symbolsWriter.UpdateSymbols(curSymbols)
	}

	// todo: watch the vote accepted event to let the client to know their reporting is accepted.

	// subscribe on-chain round rotation event
	chanRoundEvent := make(chan *contract.OracleUpdatedRound)
	subRoundEvent, err := oc.WatchUpdatedRound(new(bind.WatchOpts), chanRoundEvent)
	if err != nil {
		panic(err)
	}

	// subscribe on-chain symbol update event
	chanSymbolsEvent := make(chan *contract.OracleUpdatedSymbols)
	subSymbolEvent, err := oc.WatchUpdatedSymbols(new(bind.WatchOpts), chanSymbolsEvent)
	if err != nil {
		panic(err)
	}

	dp := &DataReporter{
		currentRound:     curRound,
		oracleContract:   oc,
		validatorAccount: validatorAccount,
		client:           client,
		autonityWSUrl:    ws,
		roundData:        make(map[uint64]*types.RoundData),
		key:              key,
		oracleService:    symbolsWriter,
		chRoundEvent:     chanRoundEvent,
		chSymbolsEvent:   chanSymbolsEvent,
		subSymbolsEvent:  subSymbolEvent,
		subRoundEvent:    subRoundEvent,
	}

	dp.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(dp).String(),
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	dp.logger.Info("Running data reporter", "rpc: ", ws)
	return dp
}

// getStartingStates returns round id, symbols and committees on current chain, it is called on the startup of client.
func getStartingStates(oc *contract.Oracle) (uint64, []string) {
	// on the startup, we need to sync the round id, symbols and committees from contract.
	currentRound, err := oc.GetRound(nil)
	if err != nil {
		panic(err)
	}

	symbols, err := oc.GetSymbols(nil)
	if err != nil {
		panic(err)
	}

	return currentRound.Uint64(), symbols
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
		}
	}
}

func (dp *DataReporter) isCommitteeMember() bool {
	committee, err := dp.oracleContract.GetCommittee(nil)
	if err != nil {
		return false
	}

	for _, c := range committee {
		if c == dp.validatorAccount {
			return true
		}
	}
	return false
}

func (dp *DataReporter) handleRoundChange(newRound uint64) error {
	dp.currentRound = newRound
	// todo, if the autonity node is on peer synchronization state, just skip the reporting.
	// if client is not a committee member, just skip reporting.
	if !dp.isCommitteeMember() {
		return nil
	}

	symbols, err := dp.oracleContract.GetSymbols(nil)
	if err != nil {
		return err
	}

	//todo, if there is prices not available, shall we wait for a while(10s) until we get the all the prices to be ready.
	prices := dp.oracleService.GetPricesBySymbols(symbols)
	curRoundData := dp.computeCommitment(prices)
	// save current round's commitment, prices and the random salt.
	dp.roundData[newRound] = curRoundData

	// query last round's prices, its random salt which will reveal last round's report.
	lastRoundData, ok := dp.roundData[newRound-1]
	if !ok {
		dp.logger.Info("Cannot find last round's data")
	}

	// prepare the transaction which carry current round's commitment, and last round's data.
	err = dp.doReport(curRoundData.Hash, lastRoundData, symbols)
	if err != nil {
		return err
	}
	return nil
}

func (dp *DataReporter) doReport(commitment common.Hash, lastRoundData *types.RoundData, symbols []string) error {
	if lastRoundData == nil {
		// cannot find last round data, just submit commitment hash.
	}
	// prepare transaction and do the reporting.
	return nil
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

	// todo, double check the endian problem and the packing of 0 bits of big integer.
	roundData.Hash = crypto.Keccak256Hash(sum.Bytes())
	return roundData
}

func (dp *DataReporter) handleNewSymbols(symbols []string) {
	dp.oracleService.UpdateSymbols(symbols)
}

func (dp *DataReporter) Stop() {
	dp.client.Close()
	dp.subRoundEvent.Unsubscribe()
	dp.subSymbolsEvent.Unsubscribe()
}
