package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCAXClient(t *testing.T) {
	// set the CAX of dev-net for testing by default since piccadilly CAX is not ready.
	defaultConfig.Endpoint = "cax.devnet.clearmatics.network"
	client := NewCAXClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"ATN-USDC", "NTN-USDC", "NTN-ATN"})
	require.NoError(t, err)
	require.Equal(t, 3, len(prices))
}
