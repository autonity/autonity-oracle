package types

import (
	"errors"
	"github.com/shopspring/decimal"
)

var (
	ErrWrongParameters = errors.New("wrong parameters")
)

type Aggregator interface {
	Initialize(config *AggregatorConfig) error
	Aggregate(trs Trades) (decimal.Decimal, error)
}

type TradePool interface {
	Initialize(config *TradePoolConfig) error
	PushTrades(provider string, symbol string, trs Trades, isAccumulatedVolume bool)
	TradeEventReceiver() chan *TradesEvent
	ConsumeTrades(symbol string) (Trades, error)
	GetSymbols() []string
	TradeUpdated(symbol string) bool
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
	Timestamp int64
	Price     decimal.Decimal
	Volume    decimal.Decimal
}

type Trades []*Trade

type TradesEvent struct {
	Provider string
	Symbol   string
	Trs      Trades
}

type TradesBySymbol map[string]Trades

type PriceBySymbol map[string]decimal.Decimal

type AdapterConfig struct {
	symbols  []string // support symbols
	interval int64    // update interval
	timeout  int      // api call timeout
	apiKey   string   // optional, api key.
}

type TradePoolConfig struct {
}

type AggregatorConfig struct {
}
