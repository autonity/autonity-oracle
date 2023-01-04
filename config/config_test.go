package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestResolveSymbols(t *testing.T) {
	t.Run("no symbols set in system environment variable", func(t *testing.T) {
		symbols := resolveSymbols()
		expected := strings.Split(DefaultSymbols, ",")
		require.Equal(t, expected, symbols)
	})
	t.Run("symbols from system environment variable", func(t *testing.T) {
		err := os.Setenv(EnvCryptoSymbols, "NTNBTC,NTNUSDT,")
		require.NoError(t, err)
		symbols := resolveSymbols()
		require.Equal(t, []string{"NTNBTC", "NTNUSDT"}, symbols)
	})
	t.Run("symbols from system environment variable", func(t *testing.T) {
		err := os.Setenv(EnvCryptoSymbols, "NTNBTC,NTNUSDT,NTNUSDC,  ")
		require.NoError(t, err)
		symbols := resolveSymbols()
		require.Equal(t, []string{"NTNBTC", "NTNUSDT", "NTNUSDC"}, symbols)
	})
}

func TestResolvePort(t *testing.T) {
	t.Run("no port set in system environment variable", func(t *testing.T) {
		port := resolvePort()
		require.Equal(t, port, DefaultPort)
	})
	t.Run("port from system environment variable", func(t *testing.T) {
		err := os.Setenv(EnvHTTPPort, "20000")
		require.NoError(t, err)
		port := resolvePort()
		require.Equal(t, port, 20000)
	})
}

func TestMakeConfig(t *testing.T) {
	t.Run("no config is set from system variable", func(t *testing.T) {
		err := os.Unsetenv(EnvHTTPPort)
		require.NoError(t, err)
		err = os.Unsetenv(EnvCryptoSymbols)
		require.NoError(t, err)
		conf := MakeConfig()
		expectedSymbols := strings.Split(DefaultSymbols, ",")
		require.Equal(t, conf.HTTPPort, DefaultPort)
		require.Equal(t, conf.Symbols, expectedSymbols)
	})

	t.Run("config set by system environment variable", func(t *testing.T) {
		err := os.Setenv(EnvHTTPPort, "20000")
		require.NoError(t, err)
		err = os.Setenv(EnvCryptoSymbols, "NTNBTC,NTNUSDT,NTNUSDC,  ")
		require.NoError(t, err)
		conf := MakeConfig()
		require.Equal(t, 20000, conf.HTTPPort)
		require.Equal(t, []string{"NTNBTC", "NTNUSDT", "NTNUSDC"}, conf.Symbols)
	})
}
