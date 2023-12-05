package helpers

import (
	"autonity-oracle/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParsePlaybookHeader(t *testing.T) {
	playbook := "../test_data/samplePlaybook.csv"
	symbols, err := ParsePlaybookHeader(playbook)
	require.NoError(t, err)
	require.Equal(t, 8, len(symbols))
	require.Equal(t, "ATN-USD", symbols[0])
	require.Equal(t, "NTN-USD", symbols[1])
	require.Equal(t, "EUR-USD", symbols[2])
	require.Equal(t, "JPY-USD", symbols[3])
	require.Equal(t, "GBP-USD", symbols[4])
	require.Equal(t, "AUD-USD", symbols[5])
	require.Equal(t, "CAD-USD", symbols[6])
	require.Equal(t, "SEK-USD", symbols[7])
}

func TestResolveSimulatedPrice(t *testing.T) {
	symbol := "BTC-ETH"
	price := ResolveSimulatedPrice(symbol)
	require.Equal(t, true, types.SimulatedPrice.Equal(price))

	symbol = "NTN-USD"
	p := ResolveSimulatedPrice(symbol)
	require.Equal(t, true, pNTNUSD.Equal(p))

	symbol = "ATN-USD"
	p = ResolveSimulatedPrice(symbol)
	require.Equal(t, true, pATNUSD.Equal(p))
}

func TestMedian(t *testing.T) {
	t.Run("normal cases with 1 sample", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0")}
		aggPrice, err := Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.0"))
	})
	t.Run("normal cases with 2 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("2.0"), decimal.RequireFromString("1.0")}
		aggPrice, err := Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.5"))
	})
	t.Run("normal cases with 3 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"),
			decimal.RequireFromString("2.0"), decimal.RequireFromString("3.0")}
		aggPrice, err := Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("2.0"))
	})

	t.Run("normal cases with 4 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"),
			decimal.RequireFromString("2.0"), decimal.RequireFromString("3.0"),
			decimal.RequireFromString("4.0")}

		aggPrice, err := Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("2.5"))
	})

	t.Run("with an empty prices set", func(t *testing.T) {
		_, err := Median(nil)
		require.Error(t, err)
	})
}
