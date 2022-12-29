package aggregator

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAveragePriceAggregator_Aggregate(t *testing.T) {
	t.Run("normal cases with 1 sample", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0")}
		aggregator := NewAveragePriceAggregator()

		aggPrice, err := aggregator.Aggregate(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.0"))
	})
	t.Run("normal cases with 2 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"), decimal.RequireFromString("2.0")}
		aggregator := NewAveragePriceAggregator()

		aggPrice, err := aggregator.Aggregate(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.5"))
	})
	t.Run("normal cases with 3 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"),
			decimal.RequireFromString("2.0"), decimal.RequireFromString("3.0")}
		aggregator := NewAveragePriceAggregator()

		aggPrice, err := aggregator.Aggregate(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("2.0"))
	})
	t.Run("with an empty prices set", func(t *testing.T) {
		aggregator := NewAveragePriceAggregator()
		_, err := aggregator.Aggregate(nil)
		require.Error(t, err)
	})
}
