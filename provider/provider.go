package provider

import (
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
)

const TVWAP_PERIOD = 3 * 60 * 1000 // 3 minutes

type Provider interface {
	Initialize()

	// Tick a tick to check if quoters have new trades to adjust the prices in oracle server.
	Tick(ts uint64) bool

	GetPriceBy(symbol string) (decimal.Decimal, error)

	GetPrices() (types.PriceBySymbol, error)

	CollectTrades(symbol string) (types.Trades, error)

	CollectPrice(symbol string) ([]decimal.Decimal, error)

	AdjustPrices()

	DumpPriceReport(ts uint64)
}
