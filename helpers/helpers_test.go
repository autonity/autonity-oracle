package helpers

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestParsePlaybookHeader(t *testing.T) {
	playbook := "../test_data/samplePlaybook.csv"
	symbols, err := ParsePlaybookHeader(playbook)
	require.NoError(t, err)
	require.Equal(t, 8, len(symbols))
	require.Equal(t, "ATN-USDC", symbols[0])
	require.Equal(t, "NTN-USDC", symbols[1])
	require.Equal(t, "EUR-USD", symbols[2])
	require.Equal(t, "JPY-USD", symbols[3])
	require.Equal(t, "GBP-USD", symbols[4])
	require.Equal(t, "AUD-USD", symbols[5])
	require.Equal(t, "CAD-USD", symbols[6])
	require.Equal(t, "SEK-USD", symbols[7])
}

func TestResolveSimulatedPrice(t *testing.T) {
	symbol := "BTC-ETH"
	price := ResolveSimulatedPrice(symbol)
	require.Equal(t, true, pBTCETH.Equal(price))

	symbol = "NTN-USDC"
	p := ResolveSimulatedPrice(symbol)
	require.Equal(t, true, pNTNUSD.Equal(p))

	symbol = "ATN-USDC"
	p = ResolveSimulatedPrice(symbol)
	require.Equal(t, true, pATNUSD.Equal(p))
}

func TestMedian(t *testing.T) {
	t.Run("normal cases with 1 sample", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0")}
		aggPrice, err := Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.0"))
	})
	t.Run("normal cases with 2 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("2.0"), decimal.RequireFromString("1.0")}
		aggPrice, err := Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("1.5"))
	})
	t.Run("normal cases with 3 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"),
			decimal.RequireFromString("2.0"), decimal.RequireFromString("3.0")}
		aggPrice, err := Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("2.0"))
	})

	t.Run("normal cases with 4 samples", func(t *testing.T) {
		var prices = []decimal.Decimal{decimal.RequireFromString("1.0"),
			decimal.RequireFromString("2.0"), decimal.RequireFromString("3.0"),
			decimal.RequireFromString("4.0")}

		aggPrice, err := Median(prices)
		require.NoError(t, err)
		aggPrice.Equals(decimal.RequireFromString("2.5"))
	})

	t.Run("with an empty prices set", func(t *testing.T) {
		_, err := Median(nil)
		require.Error(t, err)
	})
}

func TestVWAP(t *testing.T) {
	tests := []struct {
		prices             []decimal.Decimal
		volumes            []*big.Int
		expectedVWAP       decimal.Decimal
		expectedHighestVol *big.Int
		expectError        bool
	}{
		{
			prices: []decimal.Decimal{
				decimal.NewFromFloat(100.0),
				decimal.NewFromFloat(200.0),
				decimal.NewFromFloat(100.0),
			},
			volumes: []*big.Int{
				big.NewInt(10),
				big.NewInt(20),
				big.NewInt(10),
			},
			expectedVWAP:       decimal.NewFromFloatWithExponent(150, 0), // (100*10 + 200*20 + 150*30) / (10 + 20 + 30)
			expectedHighestVol: big.NewInt(20),
			expectError:        false,
		},
		{
			prices: []decimal.Decimal{
				decimal.NewFromFloat(50.0),
				decimal.NewFromFloat(75.0),
			},
			volumes: []*big.Int{
				big.NewInt(5),
				big.NewInt(0),
			},
			expectedVWAP:       decimal.NewFromFloatWithExponent(50, 0),
			expectedHighestVol: big.NewInt(5),
		},
		{
			prices:             []decimal.Decimal{},
			volumes:            []*big.Int{},
			expectedVWAP:       decimal.Zero,
			expectedHighestVol: nil,
			expectError:        true,
		},
		{
			prices: []decimal.Decimal{
				decimal.NewFromFloat(100.0),
			},
			volumes: []*big.Int{
				big.NewInt(10),
				big.NewInt(20),
			},
			expectedVWAP:       decimal.Zero,
			expectedHighestVol: nil,
			expectError:        true,
		},
	}

	for _, test := range tests {
		vwap, highestVol, err := VWAP(test.prices, test.volumes)

		if test.expectError {
			if err == nil {
				t.Errorf("Expected an error for prices: %v and volumes: %v, but got none", test.prices, test.volumes)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for prices: %v and volumes: %v: %v", test.prices, test.volumes, err)
			continue
		}

		if !vwap.Equal(test.expectedVWAP) {
			t.Errorf("For prices: %v and volumes: %v, expected VWAP: %s, but got: %s", test.prices, test.volumes, test.expectedVWAP.String(), vwap.String())
		}

		if highestVol.Cmp(test.expectedHighestVol) != 0 {
			t.Errorf("For prices: %v and volumes: %v, expected highest volume: %s, but got: %s", test.prices, test.volumes, test.expectedHighestVol.String(), highestVol.String())
		}
	}
}
