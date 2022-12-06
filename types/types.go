package types

import "github.com/shopspring/decimal"

type Trade struct {
	timestamp uint64
	price     decimal.Decimal
	volume    uint64
}

type Trades []Trade

type TradesBySymbol map[string]Trades

type PriceBySymbol map[string]decimal.Decimal
