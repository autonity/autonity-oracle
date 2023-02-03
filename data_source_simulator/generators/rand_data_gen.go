package generators

import (
	"github.com/shopspring/decimal"
	"math/rand"
	"time"
)

// RandDataGenerator start to generate data point from a reference with data point drifting under a percentage range,
// the reference point could be tuned during running by calling MoveTo or MoveBy interface.
type RandDataGenerator struct {
	referencePoint decimal.Decimal // the reference data point for the data generation within the percentage range.
	distRateRange  decimal.Decimal // the data distribution percentage range base on the reference data point.
}

func NewRandDataGenerator(ref decimal.Decimal, rateRange decimal.Decimal) *RandDataGenerator {
	return &RandDataGenerator{
		referencePoint: ref,
		distRateRange:  rateRange,
	}
}

func (rg *RandDataGenerator) SetDistributionRate(rate decimal.Decimal) {
	rg.distRateRange = rate
}

func (rg *RandDataGenerator) MoveTo(target decimal.Decimal) {
	rg.referencePoint = target
}

// MoveBy move to a new target by a percentage which could be negative as well.
func (rg *RandDataGenerator) MoveBy(percentage decimal.Decimal) {
	delta := rg.referencePoint.Mul(percentage)
	rg.referencePoint = rg.referencePoint.Add(delta)
}

func (rg *RandDataGenerator) NextDataPoint() decimal.Decimal {
	rand.Seed(time.Now().UnixNano())
	per := rg.distRateRange.Mul(decimal.NewFromFloat(rand.Float64())) // nolint
	delta := rg.referencePoint.Mul(per)
	if rand.Int()%2 == 0 { // nolint
		return delta.Add(rg.referencePoint)
	}
	return rg.referencePoint.Sub(delta)
}
