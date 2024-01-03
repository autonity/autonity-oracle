package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSIMClient(t *testing.T) {
	client := NewSIMClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"ATN-USD", "NTN-USD", "NTN-ATN"})
	require.NoError(t, err)
	require.Equal(t, 3, len(prices))
}

func TestSIMClient_AvailableSymbols(t *testing.T) {
	client := NewSIMClient(&defaultConfig)
	symbols, err := client.AvailableSymbols()
	require.NoError(t, err)

	require.Contains(t, symbols, "ATN-USD")
}
