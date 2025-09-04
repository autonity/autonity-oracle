package types

import (
	"encoding/json"
	"errors"
	"math/big"

	contract "autonity-oracle/contract_binder/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

var (
	Deployer                = common.Address{}
	DefaultVolume           = new(big.Int).SetInt64(1000000) // used by forex currency which does not have trade volumes.
	AutonityContractAddress = crypto.CreateAddress(Deployer, 0)
	OracleContractAddress   = crypto.CreateAddress(Deployer, 2)

	ErrOnOutlierSlashing = errors.New("client is on outlier Slashing")
	ErrPeerOnSync        = errors.New("l1 node is on peer sync")
	ErrNoAvailablePrice  = errors.New("no available prices collected yet")
	ErrNoDataRound       = errors.New("no data collected at current round")
	ErrNoSymbolsObserved = errors.New("no symbols observed from oracle contract")
	ErrMissingDataPoint  = errors.New("missing data point")
	ErrMissingServiceKey = errors.New("the key to access the data source is missing, please check the plugin config")
)

// Price is the structure contains the exchange rate of a symbol with a timestamp at which the sampling happens.
type Price struct {
	Timestamp  int64           `json:"timestamp"` // TS on when the data is being sampled in time's seconds since Jan 1 1970 (Unix time).
	Symbol     string          `json:"symbol"`
	Price      decimal.Decimal `json:"price"`
	Confidence uint8           `json:"confidence"` // confidence of the data point is resolved by the oracle server.
	// Below field is reserved for data providers which can provide recent trade volumes of the pair,
	// otherwise it will be resolved by oracle server.
	Volume *big.Int `json:"volume"` // recent trade volume in quote of USDCx for on-chain AMM marketplace.
}

// PriceBySymbol group the price by symbols.
type PriceBySymbol map[string]Price

// VoteRecord contains the aggregated price by symbols for a round with those ordered symbols and a corresponding salt to
// compute the round commitment hash.
type VoteRecord struct {
	// Round meta data.
	RoundID     uint64 `json:"round_id"`
	RoundHeight uint64 `json:"round_height"`
	VotePeriod  uint64 `json:"vote_period"`

	// TXN meta data.
	Mined   bool        `json:"mined"`
	TxHash  common.Hash `json:"tx_hash"`
	TxNonce uint64      `json:"tx_nonce"`
	TxCost  *big.Int    `json:"tx_cost"`

	// Report meta data.
	Salt           *big.Int                 `json:"salt"`
	CommitmentHash common.Hash              `json:"commitment_hash"`
	Prices         PriceBySymbol            `json:"prices"`
	Symbols        []string                 `json:"symbols"`
	Reports        []contract.IOracleReport `json:"reports"`
	Error          string                   `json:"error"`
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
