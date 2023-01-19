package aggregator

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAggregator_Mean(t *testing.T) {
	t.Run("normal cases with 1 sample", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0")}
		aggregator := NewAggregator()

		aggPrice, err := aggregator.Mean(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.0"))
	})
	t.Run("normal cases with 2 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"), decimal.RequireFromString("2.0")}
		aggregator := NewAggregator()

		aggPrice, err := aggregator.Mean(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.5"))
	})
	t.Run("normal cases with 3 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"),
			decimal.RequireFromString("2.0"), decimal.RequireFromString("3.0")}
		aggregator := NewAggregator()

		aggPrice, err := aggregator.Mean(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("2.0"))
	})
	t.Run("with an empty prices set", func(t *testing.T) {
		aggregator := NewAggregator()
		_, err := aggregator.Mean(nil)
		require.Error(t, err)
	})
}

func TestAggregator_Median(t *testing.T) {
	t.Run("normal cases with 1 sample", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0")}
		aggregator := NewAggregator()

		aggPrice, err := aggregator.Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.0"))
	})
	t.Run("normal cases with 2 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("2.0"), decimal.RequireFromString("1.0")}
		aggregator := NewAggregator()

		aggPrice, err := aggregator.Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.5"))
	})
	t.Run("normal cases with 3 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"),
			decimal.RequireFromString("2.0"), decimal.RequireFromString("3.0")}
		aggregator := NewAggregator()

		aggPrice, err := aggregator.Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("2.0"))
	})

	t.Run("normal cases with 4 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"),
			decimal.RequireFromString("2.0"), decimal.RequireFromString("3.0"),
			decimal.RequireFromString("4.0")}
		aggregator := NewAggregator()

		aggPrice, err := aggregator.Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("2.5"))
	})

	t.Run("with an empty prices set", func(t *testing.T) {
		aggregator := NewAggregator()
		_, err := aggregator.Median(nil)
		require.Error(t, err)
	})
}
