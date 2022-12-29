package aggregator

import (
	"fmt"
	"github.com/shopspring/decimal"
)

type AveragePriceAggregator struct {
}

func NewAveragePriceAggregator() *AveragePriceAggregator {
	return &AveragePriceAggregator{}
}

// Aggregate return the price aggregation in average price algorithm.
func (apa *AveragePriceAggregator) Aggregate(prices []decimal.Decimal) (decimal.Decimal, error) {
	if len(prices) == 0 {
		return decimal.Decimal{}, fmt.Errorf("nothing to aggregate")
	}

	if len(prices) == 1 {
		return prices[0], nil
	}

	return decimal.Avg(prices[0], prices[1:]...), nil
}
