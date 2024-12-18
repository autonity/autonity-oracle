package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewNPClient(t *testing.T) {
	defaultConfig.Key = "sandbox"
	client := NewNPClient(&defaultConfig)
	defer client.Close()
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
