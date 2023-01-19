package aggregator

import (
	"fmt"
	"github.com/shopspring/decimal"
	"sort"
)

type Aggregator struct {
}

func NewAggregator() *Aggregator {
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
	l := len(prices)
	if l == 0 {
		return decimal.Decimal{}, fmt.Errorf("empty data set")
	}

	if l == 1 {
		return prices[0], nil
	}

	sort.SliceStable(prices, func(i, j int) bool {
		return prices[i].Cmp(prices[j]) == -1
	})

	if len(prices)%2 == 0 {
		return prices[l/2-1].Add(prices[l/2]).Div(decimal.RequireFromString("2.0")), nil
	}

	return prices[l/2], nil
}
