package trade_pool

import (
	"autonity-oralce/types"
	"math"
	"sort"
	"sync"
	"time"
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

func (t *TradesByProvider) AddTrades(symbol string, trs types.Trades, isAccumulatedVolume bool) {
	if len(symbol) == 0 || len(trs) == 0 {
		return
	}
	t.lock.Lock()
	defer t.lock.Unlock()

	sort.SliceStable(trs, func(i, j int) bool {
		return trs[i].Timestamp < trs[j].Timestamp
	})

	for _, tr := range trs {
		// merge trades within a minute candle stick.
		candleTimestamp := int64(math.Floor(float64(tr.Timestamp/60000)) * 60000)

		found := false
		for _, trade := range t.tradePool[symbol] {
			if trade.Timestamp == candleTimestamp {
				found = true
				trade.Price = tr.Price
				if isAccumulatedVolume {
					trade.Volume = tr.Volume
				} else {
					trade.Volume = trade.Volume.Add(tr.Volume)
				}
				break
			}
		}

		if !found {
			tr.Timestamp = candleTimestamp
			t.tradePool[symbol] = append(t.tradePool[symbol], tr)
			t.tradeUpdated[symbol] = true
		}
	}
	return
}

func (t *TradesByProvider) ConsumeTrades(symbol string) (types.Trades, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	now := time.Now().UnixMilli()
	var recent60minutes types.Trades
	var recent3minutes types.Trades
	for _, td := range t.tradePool[symbol] {
		if now-td.Timestamp < 3*60*1000 && now >= td.Timestamp {
			recent3minutes = append(recent3minutes, td)
		}
		if now-td.Timestamp < 60*60*1000 && now >= td.Timestamp {
			recent60minutes = append(recent60minutes, td)
		}
	}
	// just keep recent 60minutes trades.
	t.tradePool[symbol] = nil
	t.tradePool[symbol] = recent60minutes

	// set the checkpoint for aggregation.
	t.tradeUpdated[symbol] = false
	return recent3minutes, nil
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

func (tp *TradesPool) PushTrades(provider string, symbol string, trs types.Trades, isAccumulatedVolume bool) {
	if len(symbol) == 0 || len(trs) == 0 {
		return
	}

	p := tp.GetTradesPoolByProvider(provider)
	p.AddTrades(symbol, trs, isAccumulatedVolume)
}

// ConsumeTrades get recently trades by symbol for aggregation
func (tp *TradesPool) ConsumeTrades(symbol string) (types.Trades, error) {
	var allTrades types.Trades
	for provider := range tp.tradePool {
		p := tp.GetTradesPoolByProvider(provider)
		trades, err := p.ConsumeTrades(symbol)
		if err == nil {
			allTrades = append(allTrades, trades...)
		}
	}
	return allTrades, nil
}

func (tp *TradesPool) TradeUpdated(symbol string) bool {
	for provider := range tp.tradePool {
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
