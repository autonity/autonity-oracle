package pluginwrapper

import (
	"autonity-oracle/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPriceProvider(t *testing.T) {
	provider := "binance_simulator"
	NTNUSDC := "NTNUSDC"
	NTNETH := "NTNUETH"
	t.Run("Add price and get price", func(t *testing.T) {
		p := NewDataCache(provider)
		prices := []types.Price{{
			Timestamp: 0,
			Symbol:    NTNUSDC,
			Price:     decimal.RequireFromString("12.33"),
		}, {
			Timestamp: 0,
			Symbol:    NTNETH,
			Price:     decimal.RequireFromString("13.669"),
		}}

		p.AddSample(prices)
		actualNTNUSDC, err := p.GetSample(NTNUSDC)
		require.NoError(t, err)
		require.Equal(t, true, prices[0].Price.Equals(actualNTNUSDC.Price))

		actualNTNETH, err := p.GetSample(NTNETH)
		require.NoError(t, err)
		require.Equal(t, true, prices[1].Price.Equals(actualNTNETH.Price))

		_, err = p.GetSample("not existed symbol")
		require.Error(t, err)
	})
}

func TestPriceProviderPool(t *testing.T) {
	provider := "Binance"
	t.Run("add price plugin_wrapper", func(t *testing.T) {
		pool := NewDataCacheSet()
		pool.AddDataCache(provider)
		actualProvider := pool.GetDataCache(provider)
		require.Equal(t, provider, actualProvider.name)
	})
	t.Run("delete price plugin_wrapper", func(t *testing.T) {
		pool := NewDataCacheSet()
		pool.AddDataCache(provider)
		actualProvider := pool.GetDataCache(provider)
		require.Equal(t, provider, actualProvider.name)
		pool.DeleteDataCache(provider)
		require.Equal(t, true, pool.GetDataCache(provider) == nil)
	})
}
