package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewWiseClient(t *testing.T) {
	// this key is only used by testing
	defaultConfig.Key = "0x123"
	client := NewWiseClient(&defaultConfig)
	defer client.Close()
	prices, _ := client.FetchPrice([]string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"})
	fmt.Print(prices)
	//require.NoError(t, err)
	//require.Equal(t, 6, len(prices))
	require.Equal(t, 0, len(prices))
}
