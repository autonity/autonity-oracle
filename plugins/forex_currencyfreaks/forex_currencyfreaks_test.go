package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCFClient(t *testing.T) {
	t.Skip("skip it due to the key of the test account reaches the rate limit.")
	var conf = types.PluginConfig{
		Name: "forex-CurrencyFreaks",
	}

	common.ResolveConf(&conf, &defaultConfig)

	client := NewCFClient(&conf)
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
