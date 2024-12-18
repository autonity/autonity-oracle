package common

import (
	"autonity-oracle/config"
	"github.com/stretchr/testify/require"
	"testing"
)

// set to bakerloo feed service by default.
var defaultEndpoint = "simfeed.bakerloo.autonity.org"

var defaultConfig = config.PluginConfig{
	Name:               "simulator_plugin",
	Key:                "",
	Scheme:             "https",
	Endpoint:           defaultEndpoint,
	Timeout:            10, //10s
	DataUpdateInterval: 10, //10s
}

func TestNewSIMClient(t *testing.T) {
	client := NewSIMClient(&defaultConfig)
	defer client.Close()
	prices, err := client.FetchPrice([]string{"ATN-USDC", "NTN-USDC", "NTN-ATN"})
	require.NoError(t, err)
	require.Equal(t, 3, len(prices))
}

func TestSIMClient_AvailableSymbols(t *testing.T) {
	client := NewSIMClient(&defaultConfig)
	defer client.Close()
	symbols, err := client.AvailableSymbols()
	require.NoError(t, err)

	require.Contains(t, symbols, "ATN-USDC")
}
