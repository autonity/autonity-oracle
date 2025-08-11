package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewForexRateAPIClient(t *testing.T) {
	// This key is only used for testing.
	defaultConfig.Key = "6ec1e9207007c984847915da9440e6d7"

	client := NewForexRateAPIClient(&defaultConfig)
	defer client.Close()

	symbolsToTest := []string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"}
	prices, err := client.FetchPrice(symbolsToTest)

	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
