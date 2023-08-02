package helpers

import (
	"autonity-oracle/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseSymbols(t *testing.T) {
	defaultSymbols := "ETH/USDC,ETH/USDT,ETH/BTC"
	symbolsWithSpace := " ETH/USDC , ETH/USDT, ETH/BTC "
	outputSymbols := ParseSymbols(defaultSymbols)
	require.Equal(t, 3, len(outputSymbols))
	require.Equal(t, "ETH/USDC", outputSymbols[0])
	require.Equal(t, "ETH/USDT", outputSymbols[1])
	require.Equal(t, "ETH/BTC", outputSymbols[2])

	results := ParseSymbols(symbolsWithSpace)
	require.Equal(t, outputSymbols, results)
}

func TestParsePlaybookHeader(t *testing.T) {
	playbook := "../test_data/samplePlaybook.csv"
	symbols, err := ParsePlaybookHeader(playbook)
	require.NoError(t, err)
	require.Equal(t, 4, len(symbols))
	require.Equal(t, "NTN/USD", symbols[0])
	require.Equal(t, "NTN/EUR", symbols[1])
	require.Equal(t, "NTN/AUD", symbols[2])
	require.Equal(t, "NTN/JPY", symbols[3])
}

func TestResolveSimulatedPrice(t *testing.T) {
	symbol := "BTC/ETH"
	price := ResolveSimulatedPrice(symbol)
	require.Equal(t, true, types.SimulatedPrice.Equal(price))

	symbol = "NTN/USD"
	p := ResolveSimulatedPrice(symbol)
	require.Equal(t, true, pNTNUSD.Equal(p))

	symbol = "NTN/SEK"
	p = ResolveSimulatedPrice(symbol)
	require.Equal(t, true, pNTNSEK.Equal(p))

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
