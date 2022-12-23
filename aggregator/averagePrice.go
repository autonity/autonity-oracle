package aggregator

import (
	"github.com/shopspring/decimal"
)

type AveragePriceAggregator struct {
}

func NewAveragePriceAggregator() *AveragePriceAggregator {
	return &AveragePriceAggregator{}
}

// Aggregate return the price aggregation in average price algorithm.
func (apa *AveragePriceAggregator) Aggregate(prices []decimal.Decimal) decimal.Decimal {
	return decimal.Avg(prices[0], prices[1:]...)
}
