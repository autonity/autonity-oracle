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
func confidenceAdjustedPrice(historicRoundPrice *types.Price, target int64) (*types.Price, error) {
	// Calculate the time difference between the target timestamp and the historic price timestamp
	timeDifference := target - historicRoundPrice.Timestamp

	var reducedConfidence uint8
	if timeDifference < 60 { // Less than 1 minute
		reducedConfidence = historicRoundPrice.Confidence // Keep original confidence
	} else if timeDifference < 3600 { // Less than 1 hour
		reducedConfidence = historicRoundPrice.Confidence / 2 // Reduce confidence by half
	} else {
		reducedConfidence = 1 // Set the lowest confidence to 1 if more than 1 hour old
	}

	if reducedConfidence == 0 {
		return nil, types.ErrNoAvailablePrice
	}

	historicRoundPrice.Confidence = reducedConfidence
	return historicRoundPrice, nil
}
