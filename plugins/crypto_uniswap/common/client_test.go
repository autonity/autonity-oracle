package common

import (
	config2 "autonity-oracle/config"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewUniswapClient(t *testing.T) {

	// using current piccadilly protocol configs.
	config := config2.PluginConfig{
		Name:               "crypto_uniswap",
		Scheme:             "wss",
		Endpoint:           "rpc-internal-1.piccadilly.autonity.org/ws",
		Timeout:            10,
		DataUpdateInterval: 30,
		ATNTokenAddress:    "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2",
		USDCTokenAddress:   "0xB855D5e83363A4494e09f0Bb3152A70d3f161940",
		SwapAddress:        "0x218F76e357594C82Cc29A88B90dd67b180827c88",
	}

	client, err := NewUniswapClient(&config)
	require.NoError(t, err)

	defer client.Close()

	prices, err := client.FetchPrice(supportedSymbols)
	require.NoError(t, err)
	require.Equal(t, 3, len(prices))
	for _, price := range prices {
		_, err := decimal.NewFromString(price.Price)
		require.NoError(t, err)
	}
	t.Log(prices)
}
