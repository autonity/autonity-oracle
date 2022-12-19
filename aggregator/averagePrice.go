package aggregator

import (
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
)

type AveragePriceAggregator struct {
	config *types.AggregatorConfig
}

func (apa *AveragePriceAggregator) Initialize(config *types.AggregatorConfig) error {
	// todo check config.
	apa.config = config
	return nil
}

// Aggregate return the price aggregation in average price algorithm.
func (apa *AveragePriceAggregator) Aggregate(trs types.Trades) (decimal.Decimal, error) {
	if len(trs) == 0 {
		return decimal.Decimal{}, types.ErrWrongParameters
	}
	if len(trs) == 1 {
		return trs[0].Price, nil
	}

	var prices []decimal.Decimal
	for _, trd := range trs {
		prices = append(prices, trd.Price)
	}
	return decimal.Avg(prices[0], prices[1:]...), nil
}
