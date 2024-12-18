package config

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMakeConfigWithEnvironmentVariables(t *testing.T) {
	t.Run("config set by system environment variable", func(t *testing.T) {
		err := os.Setenv(envKeyFile, "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe")
		require.NoError(t, err)
		defer os.Unsetenv(envKeyFile)

		err = os.Setenv(envKeyFilePASS, "123")
		require.NoError(t, err)
		defer os.Unsetenv(envKeyFilePASS)

		err = os.Setenv(envPluginDIR, "./")
		require.NoError(t, err)
		defer os.Unsetenv(envPluginDIR)

		err = os.Setenv(envConfidenceStrategy, "1")
		require.NoError(t, err)
		defer os.Unsetenv(envConfidenceStrategy)

		err = os.Setenv(envLogLevel, "3")
		require.NoError(t, err)
		defer os.Unsetenv(envLogLevel)

		err = os.Setenv(envGasTipCap, "30")
		require.NoError(t, err)
		defer os.Unsetenv(envGasTipCap)

		err = os.Setenv(envWS, "ws://127.0.0.1:30303")
		require.NoError(t, err)
		defer os.Unsetenv(envWS)

		err = os.Setenv(envPluginCof, "./plugin-conf.yml")
		require.NoError(t, err)
		defer os.Unsetenv(envPluginCof)

		err = os.Setenv(envProfDIR, "./profile_dir")
		require.NoError(t, err)
		defer os.Unsetenv(envProfDIR)

		conf := MakeConfig()
		require.Equal(t, "./", conf.PluginDIR)
		require.Equal(t, common.HexToAddress("0xb749d3d83376276ab4ddef2d9300fb5ce70ebafe"), conf.Key.Address)
		require.Equal(t, hclog.Info, conf.LoggingLevel)
		require.Equal(t, uint64(30), conf.GasTipCap)
		require.Equal(t, "ws://127.0.0.1:30303", conf.AutonityWSUrl)
		require.Equal(t, "./plugin-conf.yml", conf.PluginConfFile)
		require.Equal(t, ConfidenceStrategyFixed, conf.ConfidenceStrategy)
		require.Equal(t, "./profile_dir", conf.ProfileDir)
	})
}

func TestFormatVersion(t *testing.T) {
	require.Equal(t, "v0.0.0", VersionString(0))
	require.Equal(t, "v0.0.1", VersionString(1))
	require.Equal(t, "v0.0.9", VersionString(9))
	require.Equal(t, "v0.1.0", VersionString(10))
	require.Equal(t, "v0.1.9", VersionString(19))
	require.Equal(t, "v0.2.0", VersionString(20))
	require.Equal(t, "v1.2.5", VersionString(125))
	require.Equal(t, "v2.5.5", VersionString(255))
}
