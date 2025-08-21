package oracleserver

import (
	"autonity-oracle/config"
	"math"
	"testing"
)

// TestComputeConfidence tests the computeConfidence function.
func TestComputeConfidence(t *testing.T) {
	tests := []struct {
		symbol       string
		numOfSamples int
		strategy     int
		expected     uint8
	}{
		// Forex symbols with ConfidenceStrategyFixed, max confidence are expected.
		{"AUD-USD", 1, config.ConfidenceStrategyFixed, MaxConfidence},
		{"CAD-USD", 2, config.ConfidenceStrategyFixed, MaxConfidence},
		{"EUR-USD", 3, config.ConfidenceStrategyFixed, MaxConfidence},
		{"GBP-USD", 4, config.ConfidenceStrategyFixed, MaxConfidence},
		{"JPY-USD", 5, config.ConfidenceStrategyFixed, MaxConfidence},
		{"SEK-USD", 10, config.ConfidenceStrategyFixed, MaxConfidence},

		// Forex symbols with ConfidenceStrategyLinear
		{"AUD-USD", 1, config.ConfidenceStrategyLinear, uint8(BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, 1)))}, //nolint
		{"CAD-USD", 2, config.ConfidenceStrategyLinear, uint8(BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, 2)))}, //nolint
		{"EUR-USD", 3, config.ConfidenceStrategyLinear, uint8(BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, 3)))}, //nolint
		{"GBP-USD", 4, config.ConfidenceStrategyLinear, MaxConfidence},
		{"JPY-USD", 5, config.ConfidenceStrategyLinear, MaxConfidence},
		{"SEK-USD", 10, config.ConfidenceStrategyLinear, MaxConfidence},

		// Non-forex symbols with ConfidenceStrategyLinear, max confidence are expected.
		{"ATN-USD", 1, config.ConfidenceStrategyLinear, MaxConfidence},
		{"NTN-USD", 1, config.ConfidenceStrategyLinear, MaxConfidence},
		{"NTN-ATN", 1, config.ConfidenceStrategyLinear, MaxConfidence},

		// Non-forex symbols with ConfidenceStrategyFixed, max confidence are expected.
		{"ATN-USD", 1, config.ConfidenceStrategyFixed, MaxConfidence},
		{"NTN-USD", 1, config.ConfidenceStrategyFixed, MaxConfidence},
		{"NTN-ATN", 1, config.ConfidenceStrategyFixed, MaxConfidence},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			got := computeConfidence(tt.symbol, tt.numOfSamples, tt.strategy)
			if got != tt.expected {
				t.Errorf("computeConfidence(%q, %d, %d) = %d; want %d", tt.symbol, tt.numOfSamples, tt.strategy, got, tt.expected)
			}
		})
	}
}
