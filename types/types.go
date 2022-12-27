package types

import (
	"errors"
	"github.com/shopspring/decimal"
)

var (
	ErrWrongParameters = errors.New("wrong parameters")
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
