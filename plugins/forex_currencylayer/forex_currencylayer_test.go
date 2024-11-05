package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCLClient(t *testing.T) {
	defaultConfig.Key = "c4817087691d124d6ddabbb93411633b"
	client := NewCLClient(&defaultConfig)
	defer client.Close()
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
