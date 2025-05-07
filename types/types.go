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

const MaxConfidence = 100
const BaseConfidence = 40

var (
	Deployer                = common.Address{}
	NoVolumeData            = new(big.Int).SetInt64(0) // used by those plugins which cannot get volume info from data source.
	AutonityContractAddress = crypto.CreateAddress(Deployer, 0)
	OracleContractAddress   = crypto.CreateAddress(Deployer, 2)

	ErrPeerOnSync        = errors.New("l1 node is on peer sync")
	ErrNoAvailablePrice  = errors.New("no available prices collected yet")
	ErrNoDataRound       = errors.New("no data collected at current round")
	ErrNoSymbolsObserved = errors.New("no symbols observed from oracle contract")
	ErrMissingDataPoint  = errors.New("missing data point")
	ErrMissingServiceKey = errors.New("the key to access the data source is missing, please check the plugin config")
)

// Price is the structure contains the exchange rate of a symbol with a timestamp at which the sampling happens.
type Price struct {
	Timestamp  int64 // TS on when the data is being sampled in time's seconds since Jan 1 1970 (Unix time).
	Symbol     string
	Price      decimal.Decimal
	Confidence uint8 // confidence of the data point is resolved by the plugin by according to config.
	// Below field is reserved for data providers which can provide recent trade volumes of the pair,
	Volume *big.Int // recent trade volume in quote of USDCx for on-chain AMM marketplace.
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
