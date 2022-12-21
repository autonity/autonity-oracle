package aggregator

import (
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
)

type VWAP struct {
	config *types.AggregatorConfig
}

func (vw *VWAP) Initialize(config *types.AggregatorConfig) error {
	// todo check config.
	vw.config = config
	return nil
}

// Aggregate returns the volume weighted aggregation price, the formula: aggPrice = Sum(volume_i * price_i) / (total volume)
func (vw *VWAP) Aggregate(trs types.Trades) (decimal.Decimal, error) {
	if len(trs) == 0 {
		return decimal.Decimal{}, types.ErrWrongParameters
	}
	if len(trs) == 1 {
		return trs[0].Price, nil
	}

	var priceVols []decimal.Decimal
	totalVols := decimal.Decimal{}
	for _, trd := range trs {
		totalVols = totalVols.Add(trd.Volume)
		pv := trd.Price.Mul(trd.Volume)
		priceVols = append(priceVols, pv)
	}

	sum := decimal.Sum(priceVols[0], priceVols[1:]...)
	return sum.Div(totalVols), nil
}
