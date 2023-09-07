package main

import (
	"autonity-oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewOXClient(t *testing.T) {
	var conf = types.PluginConfig{
		Name: "forex-ExchangeRate",
	}

	resolveConf(&conf)

	client := NewOXClient(&conf)
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
