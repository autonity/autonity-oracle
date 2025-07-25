package main

import (
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestNewYahooFinanceClient(t *testing.T) {
	// this key is a free data plan only used by testing, do not use it in production as it will be rate limited.
	defaultConfig.Key = "Snp9kNMKrs8TiKvz4aMC96KqoHo6edIj9Y2xbPzR"
	client := NewYahooClient(&defaultConfig)
	defer client.Close()
	prices, err := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD", "USDC-USD"})
	require.NoError(t, err)
	require.Len(t, prices, 7)
	log.Println(prices)
}
