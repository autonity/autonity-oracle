package aggregator

import (
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
)

// VWAP todo: implement the volume weight average price aggregation algorithm.
type VWAP struct {
	config *types.AggregatorConfig
}

func (tv *VWAP) Initialize(config *types.AggregatorConfig) error {
	// todo check config.
	tv.config = config
	return nil
}

func (vw *VWAP) Aggregate(trs types.Trades) decimal.Decimal {
	return decimal.Decimal{}
}
