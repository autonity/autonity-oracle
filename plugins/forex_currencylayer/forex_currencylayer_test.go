package main

import (
	"autonity-oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCLClient(t *testing.T) {
	var conf = types.PluginConfig{
		Name: "forex-CurrencyLayer",
	}

	resolveConf(&conf)

	client := NewCLClient(&conf)
	prices, err := client.FetchPrice([]string{"EUR/USD", "JPY/USD", "GBP/USD", "AUD/USD", "CAD/USD", "SEK/USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
