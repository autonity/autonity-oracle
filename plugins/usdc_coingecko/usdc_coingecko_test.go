package main

import (
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestNewCoinGeckoClient(t *testing.T) {
	client := NewCoinGeckoClient(&defaultConfig)
	defer client.Close()
	prices, err := client.FetchPrice([]string{"USDC-USD"})
	require.NoError(t, err)
	log.Println(prices)
}
