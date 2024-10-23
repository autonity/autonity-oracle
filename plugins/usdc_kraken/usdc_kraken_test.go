package main

import (
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestNewKrakenClient(t *testing.T) {
	client := NewKrakenClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"USDC-USD"})
	require.NoError(t, err)
	log.Println(prices)
}
