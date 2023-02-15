// Package chain_adaptor implement the client connect to autonity client via web socket, it also implements a data
// reporting module, which not only construct the oracle contract binder on the start, and discovery the latest symbols from the
// oracle contract for oracle service, but also subscribe the chain head event, create event handler routine to coordinate
// the oracle data reporting workflows as well.
package chain_adaptor

import (
	"autonity-oracle/types"
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	tp "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	o "os"
)

type DataReporter struct {
	logger           hclog.Logger
	client           *ethclient.Client
	headers          chan *tp.Header
	sub              ethereum.Subscription
	autonityWSUrl    string
	currentRound     uint64
	roundData        map[uint64]*types.RoundData
	privateKey       *ecdsa.PrivateKey
	validatorAccount common.Address
}

func NewDataReporter(ws string, privateKey *ecdsa.PrivateKey, validatorAccount common.Address) *DataReporter {
	// connect to autonity node via web socket
	client, err := ethclient.Dial(ws)
	if err != nil {
		panic(err)
	}

	dp := &DataReporter{
		headers:          make(chan *tp.Header),
		validatorAccount: validatorAccount,
		client:           client,
		autonityWSUrl:    ws,
		roundData:        make(map[uint64]*types.RoundData),
		privateKey:       privateKey,
	}

	// subscribe chain head event for the round data reporting coordination.
	sub, err := dp.client.SubscribeNewHead(context.Background(), dp.headers)
	if err != nil {
		panic(err)
	}

	dp.sub = sub
	dp.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(dp).String(),
		Output: o.Stdout,
		Level:  hclog.Debug,
	})

	dp.logger.Info("Running data reporter", "rpc: ", ws)
	return dp
}

// Start starts the event loop to handle the chain head events.
func (dp *DataReporter) Start() {
	for {
		select {
		case err := <-dp.sub.Err():
			dp.logger.Info("reporter routine is shutting down ", err)
		case header := <-dp.headers:
			dp.logger.Info("new block: ", header.Number.Uint64())
			err := dp.handleNewBlockEvent(header)
			dp.logger.Info("do data reporting ", err.Error())
		}
	}
}

func (dp *DataReporter) Stop() {
	dp.client.Close()
	dp.sub.Unsubscribe()
}

// todo: construct the oracle contract binder instance
func (dp *DataReporter) oracleContract() error {
	return nil
}

// todo: construct the autonity contract binder instance
func (dp *DataReporter) autonityContract() error {
	return nil
}

// todo: get symbols from oracle contract
func (dp *DataReporter) latestSymbols() ([]string, error) {
	return nil, nil
}

func (dp *DataReporter) handleNewBlockEvent(header *tp.Header) error {
	// todo: use the chain head event to coordinate the data reporting.
	// getLatestSymbols, update oracle servers' symbols with the latest one.

	// getCommittee, if client is not committee member skip.

	// GetLastBlockEpoch, GetRoundLength, to resolve the latest round ID, if already reported, skip the reporting.

	// do the data reporting.
	return nil
}

// todo: do the data reporting.
// 1. collect current round's data commitment hash with a random salt.
// 2. save current round's data values.
// 3. send the report with last round's values with salts, and current rounds' commitment hash.
func (dp *DataReporter) report(round uint64) error {
	return nil
}
