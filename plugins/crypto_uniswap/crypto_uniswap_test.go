package main

import (
	"autonity-oracle/types"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"testing"
)

func TestNewUniswapClient(t *testing.T) {

	// using current piccadilly protocol configs.
	config := types.PluginConfig{
		Name:               "crypto_uniswap",
		Scheme:             "wss",
		Endpoint:           "rpc1.piccadilly.autonity.org/ws",
		Timeout:            10,
		DataUpdateInterval: 30,
		ATNTokenAddress:    "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2",
		USDCTokenAddress:   "0x3a60C03a86eEAe30501ce1af04a6C04Cf0188700",
		SwapAddress:        "0x218F76e357594C82Cc29A88B90dd67b180827c88",
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   config.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})
	client, err := NewUniswapClient(&config, logger)
	require.NoError(t, err)

	defer client.Close()

	prices, err := client.FetchPrice(supportedSymbols)
	require.NoError(t, err)
	require.Equal(t, 2, len(prices))
	for _, price := range prices {
		_, err := decimal.NewFromString(price.Price)
		require.NoError(t, err)
	}
}

func TestComputeExchangeRatio(t *testing.T) {
	tests := []struct {
		reserve0  *big.Int
		reserve1  *big.Int
		expected  float64
		expectErr bool
	}{
		{
			reserve0:  big.NewInt(100),
			reserve1:  big.NewInt(200),
			expected:  0.5,
			expectErr: false,
		},
		{
			reserve0:  big.NewInt(0),
			reserve1:  big.NewInt(200),
			expected:  0.0,
			expectErr: false,
		},
		{
			reserve0:  big.NewInt(200),
			reserve1:  big.NewInt(0),
			expected:  0.0,
			expectErr: true,
		},
		{
			reserve0:  big.NewInt(150),
			reserve1:  big.NewInt(300),
			expected:  0.5,
			expectErr: false,
		},
		{
			reserve0:  big.NewInt(1),
			reserve1:  big.NewInt(3),
			expected:  0.3333333333333333,
			expectErr: false,
		},
		{
			reserve0:  big.NewInt(1),
			reserve1:  big.NewInt(1),
			expected:  1.0,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.reserve0.String()+"_"+tt.reserve1.String(), func(t *testing.T) {
			ratio, err := ComputeExchangeRatio(tt.reserve0, tt.reserve1)

			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			if !tt.expectErr {
				if ratio == nil {
					t.Fatal("expected ratio to be non-nil")
				}

				result, _ := ratio.Float64()
				if result != tt.expected {
					t.Errorf("expected: %v, got: %v", tt.expected, result)
				}
			}
		})
	}
}
