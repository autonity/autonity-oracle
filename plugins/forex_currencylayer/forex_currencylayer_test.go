package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCLClient(t *testing.T) {
	t.Skip("monthly usage limit has been reached. Please upgrade your Subscription Plan")
	client := NewCLClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
