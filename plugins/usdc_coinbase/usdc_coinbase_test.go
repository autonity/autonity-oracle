package main

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCoinBaseClient(t *testing.T) {
	client := NewCoinBaseClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"USDC-USD"})
	require.NoError(t, err)
	require.Equal(t, 1, len(prices))
	require.Equal(t, "USDC-USD", prices[0].Symbol)
	_, err = decimal.NewFromString(prices[0].Price)
	require.NoError(t, err)
}
