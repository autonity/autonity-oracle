// Package chain_adaptor implement the client connect to autonity client via web socket, it also implements a data
// reporting module, which not only construct the oracle contract binder on the start, and discovery the latest symbols from the
// oracle contract for oracle service, but also subscribe the chain head event, create event handler routine to coordinate
// the oracle data reporting workflows as well.
package chain_adaptor

import (
	"autonity-oracle/types"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	o "os"
)

type DataReporter struct {
	logger           hclog.Logger
	client           *ethclient.Client
	autonityWSUrl    string
	currentRound     uint64
	roundData        map[uint64]*types.RoundData
	privateKey       *ecdsa.PrivateKey
	validatorAccount common.Address
}

func NewDataReporter(ws string, privateKey *ecdsa.PrivateKey, validatorAccount common.Address) *DataReporter {
	client, err := ethclient.Dial(ws)
	if err != nil {
		panic(err)
	}

	dp := &DataReporter{
		validatorAccount: validatorAccount,
		client:           client,
		autonityWSUrl:    ws,
		roundData:        make(map[uint64]*types.RoundData),
		privateKey:       privateKey,
	}
	dp.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(dp).String(),
		Output: o.Stdout,
		Level:  hclog.Debug,
	})

	dp.logger.Info("Running data reporter", "rpc: ", ws)
	return dp
}

// Start starts the data reporter, it prepares the ws connection, build contract binders, subscribe the chain head event,
// discover the latest symbols from oracle contract for oracle service, and start the chain head event handler routine to
// coordinate the data reporting workflows.
func (dp *DataReporter) Start() error {
	return nil
}

func (dp *DataReporter) Stop() {
	dp.client.Close()
	// todo: close the chain head event handler routine.
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

// todo: subscribe the chain head event
func (dp *DataReporter) subscribeChainHeadEvent() error {
	return nil
}

// todo: a chain head event handler that handles the chain head event to coordinate the oracle data reporting.
func (dp *DataReporter) chainHeadEventHandler() error {
	return nil
}

// todo: coordinate the data reporting workflow.
func (dp *DataReporter) handleNewBlockEvent() error {
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
