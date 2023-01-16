package aggregator

import (
	"fmt"
	"github.com/shopspring/decimal"
)

type Aggregator struct {
}

func NewAveragePriceAggregator() *Aggregator {
	return &Aggregator{}
}

// Mean return the price aggregation in average price algorithm.
func (apa *Aggregator) Mean(prices []decimal.Decimal) (decimal.Decimal, error) {
	if len(prices) == 0 {
		return decimal.Decimal{}, fmt.Errorf("nothing to aggregate")
	}

	if len(prices) == 1 {
		return prices[0], nil
	}

	return decimal.Avg(prices[0], prices[1:]...), nil
}

// Median return the median value in the provided data set
func (apa *Aggregator) Median(prices []decimal.Decimal) (decimal.Decimal, error) {
	// todo: compute median value from the provided data set.
	return decimal.Decimal{}, nil
}
