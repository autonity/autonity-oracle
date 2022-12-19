package trade_pool

import (
	"autonity-oralce/types"
	"sync"
)

type TradesByProvider struct {
	lock         sync.RWMutex
	tradePool    types.TradesBySymbol
	tradeUpdated map[string]bool
}

func NewTradesByProvider() *TradesByProvider {
	return &TradesByProvider{
		tradePool:    make(types.TradesBySymbol),
		tradeUpdated: make(map[string]bool),
	}
}

func (t *TradesByProvider) TradeUpdated(symbol string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if _, ok := t.tradeUpdated[symbol]; ok {
		return t.tradeUpdated[symbol]
	}
	return false
}

func (t *TradesByProvider) AddTrades(symbol string, trs types.Trades) error {
	if len(symbol) == 0 || len(trs) == 0 {
		return types.ErrWrongParameters
	}
	t.lock.Lock()
	defer t.lock.Unlock()
	// todo: merge trades with a 1 minute candle stick, and sort them.
	t.tradePool[symbol] = append(t.tradePool[symbol], trs...)
	t.tradeUpdated[symbol] = true
	return nil
}

func (t *TradesByProvider) ConsumeTrades(symbol string) (types.Trades, error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	// todo: remove those trades that are past 60 minutes.
	// todo: return those recent trades which are in 3 minutes for aggregation.
	// set the checkpoint for aggregation.
	t.tradeUpdated[symbol] = false
	return nil, nil
}

type TradesPool struct {
	conf *types.TradePoolConfig

	tradeEventChan chan *types.TradesEvent

	lock      sync.RWMutex
	symbols   map[string]struct{}
	tradePool map[string]*TradesByProvider
}

func NewTradesPool() *TradesPool {
	return &TradesPool{
		tradeEventChan: make(chan *types.TradesEvent, 1000),
		symbols:        make(map[string]struct{}),
		tradePool:      make(map[string]*TradesByProvider),
	}
}

func (tp *TradesPool) TradeEventReceiver() chan *types.TradesEvent {
	return tp.tradeEventChan
}

func (tp *TradesPool) Initialize(config *types.TradePoolConfig) error {
	tp.conf = config
	return nil
}

func (tp *TradesPool) PushTrades(provider string, symbol string, trs types.Trades) error {
	if len(symbol) == 0 || len(trs) == 0 {
		return types.ErrWrongParameters
	}

	p := tp.GetTradesPoolByProvider(provider)
	return p.AddTrades(symbol, trs)
}

// ConsumeTrades get recently trades by symbol for aggregation
func (tp *TradesPool) ConsumeTrades(symbol string) (types.Trades, error) {
	var allTrades types.Trades
	for provider, _ := range tp.tradePool {
		p := tp.GetTradesPoolByProvider(provider)
		trades, err := p.ConsumeTrades(symbol)
		if err == nil {
			allTrades = append(allTrades, trades...)
		}
	}
	return allTrades, nil
}

func (tp *TradesPool) TradeUpdated(symbol string) bool {
	for provider, _ := range tp.tradePool {
		p := tp.GetTradesPoolByProvider(provider)
		return p.TradeUpdated(symbol) == true
	}
	return false
}

func (tp *TradesPool) GetSymbols() []string {
	var ret []string
	for k := range tp.symbols {
		ret = append(ret, k)
	}
	return ret
}

func (tp *TradesPool) GetTradesPoolByProvider(provider string) *TradesByProvider {
	tp.lock.RLock()
	defer tp.lock.RUnlock()
	return tp.tradePool[provider]
}

func (tp *TradesPool) EventLoop() {
	// todo: think about this event handling is need or not.
}
