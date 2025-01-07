package types

import (
	"encoding/json"
	"errors"
	"math/big"

	contract "autonity-oracle/contract_binder/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

var (
	Deployer                = common.Address{}
	DefaultVolume           = new(big.Int).SetInt64(1000000)
	AutonityContractAddress = crypto.CreateAddress(Deployer, 0)
	OracleContractAddress   = crypto.CreateAddress(Deployer, 2)

	ErrPeerOnSync         = errors.New("l1 node is on peer sync")
	ErrNoAvailablePrice   = errors.New("no available prices collected yet")
	ErrNoSufficientPrices = errors.New("no sufficient num of prices were collected yet")
	ErrNoDataRound        = errors.New("no data collected at current round")
	ErrNoSymbolsObserved  = errors.New("no symbols observed from oracle contract")
	ErrMissingServiceKey  = errors.New("the key to access the data source is missing, please check the plugin config")
)

// Price is the structure contains the exchange rate of a symbol with a timestamp at which the sampling happens.
type Price struct {
	Timestamp        int64 // TS on when the data is being sampled in time's seconds since Jan 1 1970 (Unix time).
	Symbol           string
	Price            decimal.Decimal
	RecentVolInUsdcx *big.Int // recent trade volumes in usdcx for ATN or NTN assets, it is resolved by the plugins, oracle server use it for VWAP aggregation.
	Confidence       uint8    // confidence resolved by the server.
}

// PriceBySymbol group the price by symbols.
type PriceBySymbol map[string]Price

// RoundData contains the aggregated price by symbols for a round with those ordered symbols and a corresponding salt to
// compute the round commitment hash.
type RoundData struct {
	RoundID        uint64
	Tx             *types.Transaction
	Salt           *big.Int
	CommitmentHash common.Hash
	Prices         PriceBySymbol
	Symbols        []string
	Reports        []contract.IOracleReport
	MissingData    bool
}

// JSONRPCMessage is the JSON spec to carry those data response from the binance data simulator.
type JSONRPCMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// SampleEvent carry the symbols and TS for the data sampling.
type SampleEvent struct {
	Symbols []string
	TS      int64
}
