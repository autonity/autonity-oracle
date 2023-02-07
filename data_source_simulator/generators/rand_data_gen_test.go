package generators

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestRandDataGenerator_NextDataPoint(t *testing.T) {
	referPoint := decimal.RequireFromString("7.0")
	gen := NewRandDataGenerator(referPoint, decimal.RequireFromString("0.01"))

	for i := 0; i < 100; i++ {
		println(gen.NextDataPoint().String())
	}

	gen.MoveTo(decimal.RequireFromString("6.5"))
	for i := 0; i < 100; i++ {
		println(gen.NextDataPoint().String())
	}

	gen.MoveBy(decimal.RequireFromString("0.5"))

	for i := 0; i < 100; i++ {
		println(gen.NextDataPoint().String())
	}

	gen.MoveBy(decimal.RequireFromString("-0.5"))

	for i := 0; i < 100; i++ {
		println(gen.NextDataPoint().String())
	}

	gen.SetDistributionRate(decimal.RequireFromString("0.03"))

	for i := 0; i < 100; i++ {
		println(gen.NextDataPoint().String())
	}
}
