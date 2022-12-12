package types

import "github.com/shopspring/decimal"

type Aggregator interface {
	Initialize(config *AggregatorConfig) error
	Aggregate(trs Trades) decimal.Decimal
}

type TradePool interface {
	Initialize(config *TradePoolConfig) error
	PushTrade(symbol string, tr *Trade) error
	PushTrades(symbol string, trs Trades) error
	GetTrades(symbol string) (Trades, error)
	GetSymbols() ([]string, error)
}

type Adapter interface {
	Name() string
	Version() string
	Initialize(config *AdapterConfig, tradePool TradePool) error
	Start() error
	Stop() error
	Symbols() []string
	Alive() bool
	Config() *AdapterConfig
}

type Trade struct {
	timestamp uint64
	price     decimal.Decimal
	volume    uint64
}

type Trades []*Trade

type TradesBySymbol map[string]Trades

type PriceBySymbol map[string]decimal.Decimal

type AdapterConfig struct {
}

type TradePoolConfig struct {
}

type AggregatorConfig struct {
}
