package provider

import (
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
)

type Quoter interface {
	Initialize()
	Tick(ts uint64) bool
	GetSymbols() string
	GetPrice(symbol string) (decimal.Decimal, error)
	GetTrades(symbol string) (types.Trades, error)
	Update() bool
	SetPrice(symbol string, price decimal.Decimal)
	SetTrades(symbol string, trades types.Trades)
	SetTrade(symbol string, ts uint64, price decimal.Decimal, volume uint64, isAccumulatedVolume bool) types.Trades
	Alive() bool
	CheckAlive()
}

type WebSocketQuoter interface {
	Quoter
	Connect(wsUrl string)
	Disconnect()
	Tick(ts uint64) bool
	OnConnect()
	OnDisconnect(code int, reason string)
	OnError(err error)
	OnData(data interface{})
	OnRawData(raw interface{})
}
