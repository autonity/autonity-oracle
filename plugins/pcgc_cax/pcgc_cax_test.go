package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewCAXClient(t *testing.T) {
	t.Skipf("skipped duo to the remote service endpoint is not available")
	// set the CAX of dev-net for testing by default since piccadilly CAX is not ready.
	defaultConfig.Endpoint = "cax.devnet.clearmatics.network"
	routers = "orderbooks"
	client := NewCAXClient(&defaultConfig)
	prices, err := client.FetchPrice([]string{"ATN-USD", "NTN-USD", "NTN-ATN"})
	require.NoError(t, err)
	require.Equal(t, 3, len(prices))
}
