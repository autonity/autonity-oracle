package usdc_coinbase

import (
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestNewCoinBaseClient(t *testing.T) {
	client := NewCoinBaseClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"USDC-USD"})
	require.NoError(t, err)
	log.Println(prices)
}
