package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewOXClient(t *testing.T) {
	t.Skip("skip it due to the key of the test account reaches the rate limit.")
	client := NewOXClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
