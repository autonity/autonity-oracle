package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewOXClient(t *testing.T) {
	// this key is only used by testing
	defaultConfig.Key = "a9482aed38a844e7b08bc29bcaca7985"
	client := NewOXClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	require.NoError(t, err)
	require.Equal(t, 6, len(prices))
}
