package main

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewKrakenClient(t *testing.T) {
	client := NewKrakenClient(&defaultConfig)
	defer client.Close()
	prices, err := client.FetchPrice([]string{"USDC-USD"})
	require.NoError(t, err)
	require.Equal(t, 1, len(prices))
	require.Equal(t, "USDC-USD", prices[0].Symbol)
	_, err = decimal.NewFromString(prices[0].Price)
	require.NoError(t, err)
}
