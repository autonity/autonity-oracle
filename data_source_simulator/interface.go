package data_source_simulator

import (
	"autonity-oracle/data_source_simulator/binance_simulator/types"
	"github.com/shopspring/decimal"
)

type DataGenerator interface {
	MoveTo(target decimal.Decimal)
	MoveBy(percentage decimal.Decimal)
	SetDistributionRate(rate decimal.Decimal)
	NextDataPoint() decimal.Decimal
}

type GeneratorManager interface {
	Start()
	Stop()
	GetSymbolPrice([]string) (types.Prices, error)
	AdjustParams(params types.GeneratorParams, method string) error
}
