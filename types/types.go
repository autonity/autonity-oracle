package types

import (
	"encoding/json"
	"github.com/shopspring/decimal"
)

type Aggregator interface {
	Aggregate(prices []decimal.Decimal) decimal.Decimal
}

type PricePool interface {
	AddPrices(prices []Price)
}

type Adapter interface {
	Name() string
	Version() string
	FetchPrices(symbols []string) error
	Alive() bool
}

type Price struct {
	Timestamp int64
	Symbol    string
	Price     decimal.Decimal
}

type PriceBySymbol map[string]Price

type OracleServiceConfig struct {
	Providers []string
	Symbols   []string
	HttpPort  int
}

type JsonRpcMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}
