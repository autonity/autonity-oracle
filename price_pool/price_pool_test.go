package pricepool

import (
	"autonity-oracle/provider/crypto_provider"
	"autonity-oracle/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPriceProvider(t *testing.T) {
	provider := cryptoprovider.Binance
	NTNUSDC := "NTNUSDC"
	NTNETH := "NTNUETH"
	t.Run("Add price and get price", func(t *testing.T) {
		p := NewPriceProvider(provider)
		prices := []types.Price{{
			Timestamp: 0,
			Symbol:    NTNUSDC,
			Price:     decimal.RequireFromString("12.33"),
		}, {
			Timestamp: 0,
			Symbol:    NTNETH,
			Price:     decimal.RequireFromString("13.669"),
		}}

		p.AddPrices(prices)
		actualNTNUSDC, err := p.GetPrice(NTNUSDC)
		require.NoError(t, err)
		require.Equal(t, true, prices[0].Price.Equals(actualNTNUSDC.Price))

		actualNTNETH, err := p.GetPrice(NTNETH)
		require.NoError(t, err)
		require.Equal(t, true, prices[1].Price.Equals(actualNTNETH.Price))

		_, err = p.GetPrice("not existed symbol")
		require.Error(t, err)
	})

	t.Run("get prices", func(t *testing.T) {
		p := NewPriceProvider(provider)
		prices := []types.Price{{
			Timestamp: 0,
			Symbol:    NTNUSDC,
			Price:     decimal.RequireFromString("12.33"),
		}, {
			Timestamp: 0,
			Symbol:    NTNETH,
			Price:     decimal.RequireFromString("13.669"),
		}}

		p.AddPrices(prices)
		actual := p.GetPrices()
		require.Equal(t, 2, len(actual))
		require.Equal(t, prices[0].Symbol, actual[NTNUSDC].Symbol)
		require.Equal(t, true, prices[0].Price.Equals(actual[NTNUSDC].Price))
		require.Equal(t, prices[1].Symbol, actual[NTNETH].Symbol)
		require.Equal(t, true, prices[1].Price.Equals(actual[NTNETH].Price))
	})
}

func TestPriceProviderPool(t *testing.T) {
	provider := "Binance"
	t.Run("add price provider", func(t *testing.T) {
		pool := NewPriceProviderPool()
		pool.AddPriceProvider(provider)
		actualProvider := pool.GetPriceProvider(provider)
		require.Equal(t, provider, actualProvider.name)
	})
	t.Run("delete price provider", func(t *testing.T) {
		pool := NewPriceProviderPool()
		pool.AddPriceProvider(provider)
		actualProvider := pool.GetPriceProvider(provider)
		require.Equal(t, provider, actualProvider.name)
		pool.DeletePriceProvider(provider)
		require.Equal(t, true, pool.GetPriceProvider(provider) == nil)
	})
}
