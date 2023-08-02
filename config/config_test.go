package config

import (
	"autonity-oracle/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMakeConfigWithConfiguration(t *testing.T) {
	t.Run("config set by system environment variable", func(t *testing.T) {
		err := os.Setenv(types.EnvKeyFile, "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe")
		require.NoError(t, err)
		err = os.Setenv(types.EnvKeyFilePASS, "123")
		require.NoError(t, err)
		err = os.Setenv(types.EnvSymbols, "NTN/BTC,NTN/USDT,NTN/USDC,  ")
		require.NoError(t, err)
		err = os.Setenv(types.EnvPluginDIR, "./")
		require.NoError(t, err)

		conf := MakeConfig()
		require.Equal(t, []string{"NTN/BTC", "NTN/USDT", "NTN/USDC"}, conf.Symbols)
		require.Equal(t, "./", conf.PluginDIR)
		require.Equal(t, common.HexToAddress("0xb749d3d83376276ab4ddef2d9300fb5ce70ebafe"), conf.Key.Address)
	})
}
