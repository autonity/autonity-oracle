package types

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
)

var SimulatedPrice = decimal.RequireFromString("11.11")

type Aggregator interface {
	Aggregate(prices []decimal.Decimal) (decimal.Decimal, error)
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

type Price struct {
	Timestamp int64
	Symbol    string
	Price     decimal.Decimal
}

type PriceBySymbol map[string]Price

// Plugin list the information of the running plugins in oracle service.
type Plugin struct {
	Version string
	Name    string
	StartAt time.Time
}

type PluginByName map[string]Plugin

type OracleServiceConfig struct {
	Symbols   []string
	HTTPPort  int
	PluginDIR string
}

type JSONRPCMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}
