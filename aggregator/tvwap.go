package aggregator

import (
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
)

const TVWAP_PERIOD = 3 * 60 * 1000 // 3 minutes

// TVWAP todo: implement the time and volume weighted average price aggregation algorithm.
type TVWAP struct {
	config *types.AggregatorConfig
}

func (tv *TVWAP) Initialize(config *types.AggregatorConfig) error {
	// todo check config.
	tv.config = config
	return nil
}

func (tv *TVWAP) Aggregate(trs types.Trades) decimal.Decimal {
	return decimal.Decimal{}
}
