package config

import (
	"autonity-oracle/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMakeConfigWithEnvironmentVariables(t *testing.T) {
	t.Run("config set by system environment variable", func(t *testing.T) {
		err := os.Setenv(types.EnvKeyFile, "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe")
		require.NoError(t, err)
		defer os.Unsetenv(types.EnvKeyFile)

		err = os.Setenv(types.EnvKeyFilePASS, "123")
		require.NoError(t, err)
		defer os.Unsetenv(types.EnvKeyFilePASS)

		err = os.Setenv(types.EnvPluginDIR, "./")
		require.NoError(t, err)
		defer os.Unsetenv(types.EnvPluginDIR)

		err = os.Setenv(types.EnvLogLevel, "3")
		require.NoError(t, err)
		defer os.Unsetenv(types.EnvLogLevel)

		err = os.Setenv(types.EnvGasTipCap, "30")
		require.NoError(t, err)
		defer os.Unsetenv(types.EnvGasTipCap)

		err = os.Setenv(types.EnvWS, "ws://127.0.0.1:30303")
		require.NoError(t, err)
		defer os.Unsetenv(types.EnvWS)

		err = os.Setenv(types.EnvPluginCof, "./plugin-conf.yml")
		require.NoError(t, err)
		defer os.Unsetenv(types.EnvPluginCof)

		conf := MakeConfig()
		require.Equal(t, "./", conf.PluginDIR)
		require.Equal(t, common.HexToAddress("0xb749d3d83376276ab4ddef2d9300fb5ce70ebafe"), conf.Key.Address)
		require.Equal(t, hclog.Info, conf.LoggingLevel)
		require.Equal(t, uint64(30), conf.GasTipCap)
		require.Equal(t, "ws://127.0.0.1:30303", conf.AutonityWSUrl)
		require.Equal(t, "./plugin-conf.yml", conf.PluginConfFile)
	})
}
