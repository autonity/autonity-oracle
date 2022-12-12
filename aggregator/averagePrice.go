package aggregator

import (
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
)

// AveragePriceAggregator todo: implement the average price aggregation algorithm
type AveragePriceAggregator struct {
	config *types.AggregatorConfig
}

func (tv *AveragePriceAggregator) Initialize(config *types.AggregatorConfig) error {
	// todo check config.
	tv.config = config
	return nil
}

func (apa *AveragePriceAggregator) Aggregate(trs types.Trades) decimal.Decimal {
	return decimal.Decimal{}
}
