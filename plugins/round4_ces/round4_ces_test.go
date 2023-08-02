package main

import (
	"autonity-oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewR4CESClient(t *testing.T) {
	var conf = types.PluginConfig{
		Name: "r4-ces",
	}

	resolveConf(&conf)

	client := NewR4CESClient(&conf)
	prices, err := client.FetchPrice([]string{"EUR/USD", "JPY/USD"})
	require.NoError(t, err)
	require.Equal(t, 2, len(prices))
}

func TestR4CESClient_AvailableSymbols(t *testing.T) {
	var conf = types.PluginConfig{
		Name: "r4-ces",
	}

	resolveConf(&conf)

	client := NewR4CESClient(&conf)
	symbols, err := client.AvailableSymbols()
	require.NoError(t, err)

	require.Contains(t, symbols, "ATN/USD")
}
