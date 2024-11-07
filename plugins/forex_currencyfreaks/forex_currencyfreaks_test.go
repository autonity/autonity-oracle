package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCFClient(t *testing.T) {
	// this key is only used by testing
	defaultConfig.Key = "4a1a9ae24658499fb4e8d790f10a0bcd"
	client := NewCFClient(&defaultConfig)
	defer client.Close()
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
