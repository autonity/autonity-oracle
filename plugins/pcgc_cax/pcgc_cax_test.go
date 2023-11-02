package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCAXClient(t *testing.T) {
	var conf = types.PluginConfig{
		Name: "pcgc-cax",
	}
	common.ResolveConf(&conf, &defaultConfig)

	client := NewCAXClient(&conf)
	prices, err := client.FetchPrice([]string{"ATN-USD", "NTN-USD", "NTN-ATN"})
	require.NoError(t, err)
	require.Equal(t, 3, len(prices))
}
