package oracleserver

import (
	"autonity-oracle/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOracleServer(t *testing.T) {
	var symbols = []string{"NTNUSD", "NTNAUD", "NTNCAD", "NTNEUR", "NTNGBP", "NTNJPY", "NTNSEK"}
	t.Run("oracle service getters", func(t *testing.T) {
		os := NewOracleServer(symbols, ".")
		version := os.Version()
		require.Equal(t, Version, version)
		actualSymbols := os.Symbols()
		require.Equal(t, symbols, actualSymbols)
		prices := os.GetPrices()
		require.Equal(t, 0, len(prices))
	})

	t.Run("oracle service setters", func(t *testing.T) {
		newSymbols := []string{"NTNUSD", "NTNAUD", "NTNCAD", "NTNEUR", "NTNGBP", "NTNJPY", "NTNSEK", "NTNRMB"}
		os := NewOracleServer(symbols, ".")
		os.UpdateSymbols(newSymbols)
		require.Equal(t, newSymbols, os.Symbols())

		NTNEURRate := types.Price{
			Timestamp: 0,
			Symbol:    "NTNEUR",
			Price:     decimal.RequireFromString("999.99"),
		}
		NTNUSDRate := types.Price{
			Timestamp: 0,
			Symbol:    "NTNUSD",
			Price:     decimal.RequireFromString("127.32"),
		}
		NTNGBPRate := types.Price{
			Timestamp: 0,
			Symbol:    "NTNGBP",
			Price:     decimal.RequireFromString("111.11"),
		}
		NTNRMBRate := types.Price{
			Timestamp: 0,
			Symbol:    "NTNRMB",
			Price:     decimal.RequireFromString("12.12"),
		}
		os.UpdatePrice(NTNEURRate)
		os.UpdatePrice(NTNUSDRate)
		os.UpdatePrice(NTNGBPRate)
		os.UpdatePrice(NTNRMBRate)

		require.Equal(t, 4, len(os.GetPrices()))
		actualPrices := os.GetPrices()
		require.Equal(t, true, NTNUSDRate.Price.Equals(actualPrices["NTNUSD"].Price))
		require.Equal(t, NTNUSDRate.Symbol, actualPrices["NTNUSD"].Symbol)

		require.Equal(t, true, NTNEURRate.Price.Equals(actualPrices["NTNEUR"].Price))
		require.Equal(t, NTNEURRate.Symbol, actualPrices["NTNEUR"].Symbol)

		require.Equal(t, true, NTNGBPRate.Price.Equals(actualPrices["NTNGBP"].Price))
		require.Equal(t, NTNGBPRate.Symbol, actualPrices["NTNGBP"].Symbol)

		require.Equal(t, true, NTNRMBRate.Price.Equals(actualPrices["NTNRMB"].Price))
		require.Equal(t, NTNRMBRate.Symbol, actualPrices["NTNRMB"].Symbol)
	})
}
