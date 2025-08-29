package server

import (
	"autonity-oracle/helpers"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"slices"
	"time"

	"github.com/shopspring/decimal"
)

func (os *Server) aggregateProtocolSymbolPrices() (types.PriceBySymbol, error) {
	prices := make(types.PriceBySymbol)

	// if we need a bridger pair USDC-USD to convert ATN-USD or NTN-USD from ATN-USDC or NTN-USDC,
	// then we need to aggregate USDC-USD data point first.
	var usdcPrice *types.Price
	var err error
	if slices.Contains(os.protocolSymbols, ATNUSD) || slices.Contains(os.protocolSymbols, NTNUSD) {
		usdcPrice, err = os.aggregatePrice(USDCUSD, os.curSampleTS)
		if err != nil {
			os.logger.Error("aggregate USDC-USD price", "error", err.Error())
		}
	}

	for _, s := range os.protocolSymbols {
		// aggregate bridged symbols
		if s == ATNUSD || s == NTNUSD {
			if usdcPrice == nil {
				continue
			}

			p, e := os.aggBridgedPrice(s, os.curSampleTS, usdcPrice)
			if e != nil {
				os.logger.Error("aggregate bridged price", "error", e.Error(), "symbol", s)
				continue
			}
			prices[s] = *p
			continue
		}

		// aggregate none bridged symbols
		p, e := os.aggregatePrice(s, os.curSampleTS)
		if e != nil {
			os.logger.Debug("no data for aggregation", "reason", e.Error(), "symbol", s)
			continue
		}
		prices[s] = *p
	}

	// edge case: if NTN-ATN price was not computable from inside plugin,
	// try to compute it from NTNprice and ATNprice across from different plugins.
	if _, ok := prices[common.NTNATNSymbol]; !ok {
		ntnPrice, ntnExist := prices[NTNUSD]
		atnPrice, atnExist := prices[ATNUSD]
		if ntnExist && atnExist {
			ntnATNPrice, err := common.ComputeDerivedPrice(ntnPrice.Price.String(), atnPrice.Price.String()) //nolint
			if err == nil {
				p, err := decimal.NewFromString(ntnATNPrice.Price) // nolint
				if err == nil {
					prices[common.NTNATNSymbol] = types.Price{
						Timestamp:  time.Now().Unix(),
						Price:      p,
						Symbol:     common.NTNATNSymbol,
						Confidence: ntnPrice.Confidence,
					}
				} else {
					os.logger.Error("cannot parse NTN-ATN price in decimal", "error", err.Error())
				}
			} else {
				os.logger.Error("failed to compute NTN-ATN price", "error", err.Error())
			}
		}
	}

	return prices, nil
}

// aggBridgedPrice aggregates ATN-USD or NTN-USD from bridged ATN-USDC or NTN-USDC with USDC-USD price,
// it assumes the input usdcPrice is not nil.
func (os *Server) aggBridgedPrice(srcSymbol string, target int64, usdcPrice *types.Price) (*types.Price, error) {
	var bridgedSymbol string
	if srcSymbol == ATNUSD {
		bridgedSymbol = ATNUSDC
	}

	if srcSymbol == NTNUSD {
		bridgedSymbol = NTNUSDC
	}

	p, err := os.aggregatePrice(bridgedSymbol, target)
	if err != nil {
		os.logger.Error("aggregate bridged price", "error", err.Error(), "symbol", bridgedSymbol)
		return nil, err
	}

	// reset the symbol with source symbol,
	// and update price with: ATN-USD=ATN-USDC*USDC-USD / NTN-USD=NTN-USDC*USDC-USD
	// the confidence of ATN-USD and NTN-USD are inherit from ATN-USDC and NTN-USDC.
	p.Symbol = srcSymbol
	p.Price = p.Price.Mul(usdcPrice.Price)
	return p, nil
}

// aggregatePrice takes the symbol's aggregated data points from all the supported plugins, if there are multiple
// markets' datapoint, it will do a final VWAP aggregation to form the final reporting value.
func (os *Server) aggregatePrice(s string, target int64) (*types.Price, error) {
	prices, volumes := os.pluginManager.SelectSamples(s, target)
	if len(prices) == 0 {
		copyHistoricPrice, err := os.queryHistoricRoundPrice(s)
		if err != nil {
			return nil, err
		}

		return confidenceAdjustedPrice(&copyHistoricPrice, target)
	}

	// compute confidence of the symbol from the num of plugins' samples of it.
	confidence := computeConfidence(s, len(prices), os.conf.ConfidenceStrategy)
	price := &types.Price{
		Timestamp:  target,
		Price:      prices[0],
		Volume:     volumes[0],
		Symbol:     s,
		Confidence: confidence,
	}

	_, isForex := common.ForexCurrencies[s]

	// we have multiple markets' data for this forex symbol, update the price with median value.
	if len(prices) > 1 && isForex {
		p, err := helpers.Median(prices)
		if err != nil {
			return nil, err
		}
		price.Price = p
		price.Volume = types.DefaultVolume
		return price, nil
	}

	// we have multiple markets' data for this crypto symbol, update the price with VWAP.
	if len(prices) > 1 && !isForex {
		p, vol, err := helpers.VWAP(prices, volumes)
		if err != nil {
			return nil, err
		}
		price.Price = p
		price.Volume = vol
	}

	return price, nil
}

// queryHistoricRoundPrice queries the last available price for a given symbol from the historic rounds.
func (os *Server) queryHistoricRoundPrice(symbol string) (types.Price, error) {

	if len(os.voteRecords) == 0 {
		return types.Price{}, types.ErrNoDataRound
	}

	numOfRounds := len(os.voteRecords)
	// Iterate from the current round backward
	for i := 0; i < numOfRounds; i++ {
		roundID := os.curRound - uint64(i) - 1 //nolint
		// Get the round data for the current round ID
		voteRecord, exists := os.voteRecords[roundID]
		if !exists {
			continue
		}

		if voteRecord == nil {
			continue
		}

		// Check if the symbol exists in the Prices map
		if price, found := voteRecord.Prices[symbol]; found {
			return price, nil
		}
	}

	// If no price was found after checking all rounds, return an error
	return types.Price{}, types.ErrNoDataRound
}
