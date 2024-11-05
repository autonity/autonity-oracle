package main

import (
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestNewCoinBaseClient(t *testing.T) {
	client := NewCoinBaseClient(&defaultConfig)
	defer client.Close()
	prices, err := client.FetchPrice([]string{"USDC-USD"})
	require.NoError(t, err)
	log.Println(prices)
}
