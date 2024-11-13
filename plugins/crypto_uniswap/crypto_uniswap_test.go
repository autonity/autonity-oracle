package main

import (
	"autonity-oracle/types"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewUniswapClient(t *testing.T) {

	// using current piccadilly protocol configs.
	config := types.PluginConfig{
		Name:               "crypto_uniswap",
		Scheme:             "wss",
		Endpoint:           "rpc1.piccadilly.autonity.org/ws",
		Timeout:            10,
		DataUpdateInterval: 30,
		ATNTokenAddress:    "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2",
		USDCTokenAddress:   "0x3a60C03a86eEAe30501ce1af04a6C04Cf0188700",
		SwapAddress:        "0x218F76e357594C82Cc29A88B90dd67b180827c88",
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   config.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})
	client, err := NewUniswapClient(&config, logger)
	require.NoError(t, err)

	defer client.Close()

	prices, err := client.FetchPrice(supportedSymbols)
	require.NoError(t, err)
	require.Equal(t, 3, len(prices))
	for _, price := range prices {
		_, err := decimal.NewFromString(price.Price)
		require.NoError(t, err)
	}
}
