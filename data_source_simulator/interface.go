package data_source_simulator

import "github.com/shopspring/decimal"

type DataGenerator interface {
	MoveTo(target decimal.Decimal)
	MoveBy(percentage decimal.Decimal)
	SetDistributionRate(rate decimal.Decimal)
	NextDataPoint() decimal.Decimal
}
