package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigs(t *testing.T) {
	configFile := "./config_for_test.yml"
	config, err := LoadServerConfig(configFile)
	require.NoError(t, err)
	require.NotEmpty(t, config)
	require.Equal(t, defaultLogVerbosity, config.LoggingLevel)
	require.Equal(t, "ws://localhost:8546", config.AutonityWSUrl)
	require.Equal(t, "oracle", config.MetricConfigs.InfluxDBOrganization)
	require.Equal(t, 5, len(config.PluginConfigs))

	pluginConfigs, err := LoadPluginsConfig(configFile)
	require.NoError(t, err)
	require.NotEmpty(t, pluginConfigs)
	require.Equal(t, 5, len(pluginConfigs))

	cf, ok := pluginConfigs["not exists"]
	require.False(t, ok)
	require.NotNil(t, cf)
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
