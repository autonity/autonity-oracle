package helpers

import (
	"autonity-oracle/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseSymbols(t *testing.T) {
	defaultSymbols := "ETHUSDC,ETHUSDT,ETHBTC"
	symbolsWithSpace := " ETHUSDC , ETHUSDT, ETHBTC "
	outputSymbols := ParseSymbols(defaultSymbols)
	require.Equal(t, 3, len(outputSymbols))
	require.Equal(t, "ETHUSDC", outputSymbols[0])
	require.Equal(t, "ETHUSDT", outputSymbols[1])
	require.Equal(t, "ETHBTC", outputSymbols[2])

	results := ParseSymbols(symbolsWithSpace)
	require.Equal(t, outputSymbols, results)
}

func TestParsePlaybookHeader(t *testing.T) {
	playbook := "../test_data/samplePlaybook.csv"
	symbols, err := ParsePlaybookHeader(playbook)
	require.NoError(t, err)
	require.Equal(t, 4, len(symbols))
	require.Equal(t, "NTNUSD", symbols[0])
	require.Equal(t, "NTNEUR", symbols[1])
	require.Equal(t, "NTNAUD", symbols[2])
	require.Equal(t, "NTNJPY", symbols[3])
}

func TestResolveSimulatedPrice(t *testing.T) {
	symbol := "BTCETH"
	price := ResolveSimulatedPrice(symbol)
	require.Equal(t, true, types.SimulatedPrice.Equal(price))

	symbol = "NTNUSD"
	p := ResolveSimulatedPrice(symbol)
	require.Equal(t, true, pNTNUSD.Equal(p))

	symbol = "NTNSEK"
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
