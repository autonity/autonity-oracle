package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSIMClient(t *testing.T) {
	var conf = types.PluginConfig{
		Name: "simulator",
	}

	common.ResolveConf(&conf, &defaultConfig)

	client := NewSIMClient(&conf)
	prices, err := client.FetchPrice([]string{"ATN-USD", "NTN-USD", "NTN-ATN"})
	require.NoError(t, err)
	require.Equal(t, 3, len(prices))
}

func TestSIMClient_AvailableSymbols(t *testing.T) {
	var conf = types.PluginConfig{
		Name: "simulator",
	}

	common.ResolveConf(&conf, &defaultConfig)

	client := NewSIMClient(&conf)
	symbols, err := client.AvailableSymbols()
	require.NoError(t, err)

	require.Contains(t, symbols, "ATN-USD")
}
