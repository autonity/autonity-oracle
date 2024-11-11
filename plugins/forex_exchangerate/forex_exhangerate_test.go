package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewEXClient(t *testing.T) {
	// this key is only used by testing
	defaultConfig.Key = "fc2e53282835eb092f8cafd4"
	client := NewEXClient(&defaultConfig)
	defer client.Close()
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
