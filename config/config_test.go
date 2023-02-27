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
	EnvKeyFile       = "ORACLE_KEY_FILE"
	EnvKeyFilePASS   = "ORACLE_KEY_PASSWORD"
)

func TestMakeConfigWithConfiguration(t *testing.T) {
	t.Run("config set by system environment variable", func(t *testing.T) {
		err := os.Setenv(EnvHTTPPort, "20000")
		require.NoError(t, err)
		err = os.Setenv(EnvKeyFile, "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe")
		require.NoError(t, err)
		err = os.Setenv(EnvKeyFilePASS, "123")
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
