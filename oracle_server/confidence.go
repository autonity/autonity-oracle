package oracleserver

import (
	"autonity-oracle/config"
	common2 "autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"math"
)

// computeConfidence calculates the confidence weight based on the number of data samples. Note! Cryptos take
// fixed strategy as we have very limited number of data sources at the genesis phase. Thus, the confidence
// computing is just for forex currencies for the time being.
func computeConfidence(symbol string, numOfSamples, strategy int) uint8 {

	// Todo: once the community have more extensive AMM and DEX markets, we will remove this to enable linear
	//  strategy as well for cryptos.
	if _, is := common2.ForexCurrencies[symbol]; !is {
		return MaxConfidence
	}

	// Forex currencies with fixed strategy.
	if strategy == config.ConfidenceStrategyFixed {
		return MaxConfidence
	}

	// Forex currencies with "linear" strategy. Labeled "linear" but uses exponential scaling (1.75^n) since we
	// are at the network bootstrapping phase with very limited number of data sources.
	weight := BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, float64(numOfSamples)))

	if weight > MaxConfidence {
		weight = MaxConfidence
	}

	return uint8(weight) //nolint
}

// by according to the spreading of price timestamp from the target timestamp,
// we reduce the confidence of the price, set the lowest confidence as 1.
func confidenceAdjustedPrice(copyHistoricPrice *types.Price, target int64) (*types.Price, error) {
	// Calculate the time difference between the target timestamp and the historic price timestamp
	timeDifference := target - copyHistoricPrice.Timestamp

	var reducedConfidence uint8
	if timeDifference < 60 {
		// Less than 1 minute, keep original confidence
		reducedConfidence = copyHistoricPrice.Confidence
	} else if timeDifference < 3600 {
		// Less than 1 hour, reduce by 50%, it could be zero.
		reducedConfidence = copyHistoricPrice.Confidence / 2
	} else {
		// More than 1 hour, set to zero. For some vote records loaded from the persistence, they could be out of updated.
		reducedConfidence = 0
	}

	// If the confidence decades to zero, not to report to avoid potential outlier slashing.
	if reducedConfidence == 0 {
		return nil, types.ErrNoAvailablePrice
	}

	copyHistoricPrice.Confidence = reducedConfidence
	return copyHistoricPrice, nil
}
