package main

import (
	"autonity-oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSIMClient(t *testing.T) {
	var conf = types.PluginConfig{
		Name: "simulator",
	}

	resolveConf(&conf)

	client := NewSIMClient(&conf)
	prices, err := client.FetchPrice([]string{"EUR/USD", "JPY/USD"})
	require.NoError(t, err)
	require.Equal(t, 2, len(prices))
}

func TestSIMClient_AvailableSymbols(t *testing.T) {
	var conf = types.PluginConfig{
		Name: "simulator",
	}

	resolveConf(&conf)

	client := NewSIMClient(&conf)
	symbols, err := client.AvailableSymbols()
	require.NoError(t, err)

	require.Contains(t, symbols, "ATN/USD")
}
