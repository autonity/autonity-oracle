package aggregator

import (
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
)

const TVWAP_PERIOD = 3 * 60 * 1000 // 3 minutes

type TVWAP struct {
	config *types.AggregatorConfig
}

func (tv *TVWAP) Initialize(config *types.AggregatorConfig) error {
	// todo check config.
	tv.config = config
	return nil
}

func (tv *TVWAP) Aggregate(trs types.Trades) (decimal.Decimal, error) {
	if len(trs) == 0 {
		return decimal.Decimal{}, types.ErrWrongParameters
	}
	if len(trs) == 1 {
		return trs[0].Price, nil
	}
	// todo: weight the time period that a trade holds.
	return decimal.Decimal{}, nil
}
