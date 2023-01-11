package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

var (
	EnvHTTPPort      = "ORACLE_HTTP_PORT"
	EnvCryptoSymbols = "ORACLE_CRYPTO_SYMBOLS"
	EnvPluginDIR     = "ORACLE_PLUGIN_DIR"
)

func TestMakeConfigWithConfiguration(t *testing.T) {
	t.Run("config set by system environment variable", func(t *testing.T) {
		err := os.Setenv(EnvHTTPPort, "20000")
		require.NoError(t, err)
		err = os.Setenv(EnvCryptoSymbols, "NTNBTC,NTNUSDT,NTNUSDC,  ")
		require.NoError(t, err)
		err = os.Setenv(EnvPluginDIR, "./")
		require.NoError(t, err)

		conf := MakeConfig()
		require.Equal(t, 20000, conf.HTTPPort)
		require.Equal(t, []string{"NTNBTC", "NTNUSDT", "NTNUSDC"}, conf.Symbols)
		require.Equal(t, "./", conf.PluginDIR)
	})
}
