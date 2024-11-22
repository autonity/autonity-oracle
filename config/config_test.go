package config

import (
	"autonity-oracle/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	"math"
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

		err = os.Setenv(types.EnvConfidenceStrategy, "1")
		require.NoError(t, err)
		defer os.Unsetenv(types.EnvConfidenceStrategy)

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
		require.Equal(t, ConfidenceStrategyFixed, conf.ConfidenceStrategy)
	})
}

func TestFormatVersion(t *testing.T) {
	require.Equal(t, "v0.0.0", FormatVersion(0))
	require.Equal(t, "v0.0.1", FormatVersion(1))
	require.Equal(t, "v0.0.9", FormatVersion(9))
	require.Equal(t, "v0.1.0", FormatVersion(10))
	require.Equal(t, "v0.1.9", FormatVersion(19))
	require.Equal(t, "v0.2.0", FormatVersion(20))
	require.Equal(t, "v1.2.5", FormatVersion(125))
	require.Equal(t, "v2.5.5", FormatVersion(255))
}

// TestComputeConfidence tests the ComputeConfidence function.
func TestComputeConfidence(t *testing.T) {
	tests := []struct {
		symbol       string
		numOfSamples int
		strategy     int
		expected     uint8
	}{
		// Forex symbols with ConfidenceStrategyFixed, max confidence are expected.
		{"AUD-USD", 1, ConfidenceStrategyFixed, MaxConfidence},
		{"CAD-USD", 2, ConfidenceStrategyFixed, MaxConfidence},
		{"EUR-USD", 3, ConfidenceStrategyFixed, MaxConfidence},
		{"GBP-USD", 4, ConfidenceStrategyFixed, MaxConfidence},
		{"JPY-USD", 5, ConfidenceStrategyFixed, MaxConfidence},
		{"SEK-USD", 10, ConfidenceStrategyFixed, MaxConfidence},

		// Forex symbols with ConfidenceStrategyLinear
		{"AUD-USD", 1, ConfidenceStrategyLinear, uint8(BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, 1)))}, //nolint
		{"CAD-USD", 2, ConfidenceStrategyLinear, uint8(BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, 2)))}, //nolint
		{"EUR-USD", 3, ConfidenceStrategyLinear, uint8(BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, 3)))}, //nolint
		{"GBP-USD", 4, ConfidenceStrategyLinear, MaxConfidence},
		{"JPY-USD", 5, ConfidenceStrategyLinear, MaxConfidence},
		{"SEK-USD", 10, ConfidenceStrategyLinear, MaxConfidence},

		// Non-forex symbols with ConfidenceStrategyLinear, max confidence are expected.
		{"ATN-USD", 1, ConfidenceStrategyLinear, MaxConfidence},
		{"NTN-USD", 1, ConfidenceStrategyLinear, MaxConfidence},
		{"NTN-ATN", 1, ConfidenceStrategyLinear, MaxConfidence},

		// Non-forex symbols with ConfidenceStrategyFixed, max confidence are expected.
		{"ATN-USD", 1, ConfidenceStrategyFixed, MaxConfidence},
		{"NTN-USD", 1, ConfidenceStrategyFixed, MaxConfidence},
		{"NTN-ATN", 1, ConfidenceStrategyFixed, MaxConfidence},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			got := ComputeConfidence(tt.symbol, tt.numOfSamples, tt.strategy)
			if got != tt.expected {
				t.Errorf("ComputeConfidence(%q, %d, %d) = %d; want %d", tt.symbol, tt.numOfSamples, tt.strategy, got, tt.expected)
			}
		})
	}
}
