package trade_pool

import (
	"autonity-oralce/types"
	"sync"
)

type TradesPool struct {
	conf      *types.TradePoolConfig
	lock      sync.RWMutex
	tradePool types.TradesBySymbol
}

func (tp *TradesPool) Initialize(config *types.TradePoolConfig) error {
	return nil
}

func (tp *TradesPool) PushTrade(symbol string, tr *types.Trade) error {
	return nil
}

func (tp *TradesPool) PushTrades(symbol string, trs types.Trades) error {
	return nil
}

func (tp *TradesPool) GetTrades(symbol string) (types.Trades, error) {
	return nil, nil
}

func (tp *TradesPool) GetSymbols() ([]string, error) {
	return nil, nil
}
