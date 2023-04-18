package types

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

var (
	EnvCryptoSymbols = "ORACLE_CRYPTO_SYMBOLS"
	EnvPluginDIR     = "ORACLE_PLUGIN_DIR"
	EnvKeyFile       = "ORACLE_KEY_FILE"
	EnvKeyFilePASS   = "ORACLE_KEY_PASSWORD"
	SimulatedPrice   = decimal.RequireFromString("11.11")
	InvalidPrice     = new(big.Int).Sub(math.BigPow(2, 255), big.NewInt(1))
	InvalidSalt      = big.NewInt(0)
)

var Deployer = common.Address{}
var AutonityContractAddress = crypto.CreateAddress(Deployer, 0)
var OracleContractAddress = crypto.CreateAddress(Deployer, 1)

var ErrPeerOnSync = errors.New("l1 node is on peer sync")
var ErrNoAvailablePrice = errors.New("no available prices collected yet")
var ErrNoSymbolsObserved = errors.New("no symbols observed from oracle contract")

const MaxBufferedRounds = 10

type Aggregator interface {
	Mean(prices []decimal.Decimal) (decimal.Decimal, error)
	Median(prices []decimal.Decimal) (decimal.Decimal, error)
}

type DataPool interface {
	AddSample(prices []Price, ts int64)
	GCSamples()
}

type PluginWrapper interface {
	Name() string
	Version() string
	FetchPrices(symbols []string, ts int64) error
	GCSamples()
	Close()
	StartTime() time.Time
}

type PluginPriceReport struct {
	Prices     []Price
	BadSymbols []string
}

type Price struct {
	Timestamp int64 // TS on when the data is being sampled in time's seconds since Jan 1 1970 (Unix time).
	Symbol    string
	Price     decimal.Decimal
}

type PriceBySymbol map[string]Price

type RoundData struct {
	RoundID uint64
	Tx      *types.Transaction
	Salt    *big.Int
	Hash    common.Hash
	Prices  PriceBySymbol
	Symbols []string
}

// Plugin list the information of the running plugins in oracle service.
type Plugin struct {
	Version string
	Name    string
	StartAt time.Time
}

type PluginByName map[string]Plugin

type OracleServiceConfig struct {
	Key           *keystore.Key
	AutonityWSUrl string
	Symbols       []string
	PluginDIR     string
}

type JSONRPCMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}
