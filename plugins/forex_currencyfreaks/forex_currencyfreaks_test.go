package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCFClient(t *testing.T) {
	t.Skip("skip it due to monthly usage limit has been reached. Please upgrade your Subscription Plan.")
	client := NewCFClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
