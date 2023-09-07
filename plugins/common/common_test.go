package common

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConvertSymbol(t *testing.T) {
	srcSym := "BTC-USD"
	require.Equal(t, "BTC/USD", ConvertSymbol(srcSym, "/"))
	require.Equal(t, "BTC|USD", ConvertSymbol(srcSym, "|"))
	srcSym = "NTNUSD"
	require.Equal(t, "NTNUSD", ConvertSymbol(srcSym, "|"))
}

func TestResolveSeparator(t *testing.T) {
	symbol := "NTN/USD"
	require.Equal(t, "/", ResolveSeparator(symbol))

	symbol = "NTN|USD"
	require.Equal(t, "|", ResolveSeparator(symbol))

	symbol = "NTN-USD"
	require.Equal(t, "-", ResolveSeparator(symbol))

	symbol = "NTN,USD"
	require.Equal(t, ",", ResolveSeparator(symbol))

	symbol = "NTN.USD"
	require.Equal(t, ".", ResolveSeparator(symbol))

	symbol = "BTCUSD"
	require.Equal(t, "", ResolveSeparator(symbol))
}
