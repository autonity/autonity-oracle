package types

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"math/big"
	"time"
)

var SimulatedPrice = decimal.RequireFromString("11.11")

type Aggregator interface {
	Mean(prices []decimal.Decimal) (decimal.Decimal, error)
	Median(prices []decimal.Decimal) (decimal.Decimal, error)
}

type PricePool interface {
	AddPrices(prices []Price)
}

type PluginClient interface {
	Name() string
	Version() string
	FetchPrices(symbols []string) error
	Close()
	StartTime() time.Time
}

type OracleService interface {
	UpdateSymbols([]string)
	GetPrices() PriceBySymbol
	GetPricesBySymbols(symbols []string) PriceBySymbol
}

type Price struct {
	Timestamp int64
	Symbol    string
	Price     decimal.Decimal
}

type PriceBySymbol map[string]Price

type RoundData struct {
	Salt   *big.Int
	Hash   common.Hash
	Prices PriceBySymbol
}

// Plugin list the information of the running plugins in oracle service.
type Plugin struct {
	Version string
	Name    string
	StartAt time.Time
}

type PluginByName map[string]Plugin

type OracleServiceConfig struct {
	ValidatorAccount common.Address
	Key              *keystore.Key
	AutonityWSUrl    string
	Symbols          []string
	HTTPPort         int
	PluginDIR        string
}

type JSONRPCMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}
