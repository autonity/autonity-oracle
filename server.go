package autonity_oralce

import (
	"autonity-oralce/aggregator"
	"autonity-oralce/provider/crypto_provider"
	"autonity-oralce/trade_pool"
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
	"sync"
	"time"
)

type OracleService struct {
	version     string
	aggTime     int64
	aggInterval int
	config      *types.OracleServiceConfig

	lock   sync.RWMutex
	prices types.PriceBySymbol

	doneCh chan struct{}
	ticker *time.Ticker

	symbols    []string
	aggregator types.Aggregator
	tradePool  types.TradePool
	adapters   []types.Adapter
}

func NewOracleService(config *types.OracleServiceConfig) *OracleService {
	os := &OracleService{
		version:     "v0.0.1",
		aggInterval: config.AggInterval,
		config:      config,
		symbols:     config.Symbols,
		prices:      make(types.PriceBySymbol),
		doneCh:      make(chan struct{}),
		ticker:      time.NewTicker(10 * time.Second),
		aggregator:  aggregator.NewVWAP(), // todo: resolve the aggregation algorithm by config
		tradePool:   trade_pool.NewTradesPool(),
	}
	return os
}

func (os *OracleService) Version() string {
	return os.version
}

func (os *OracleService) LastAggTime() int64 {
	return os.aggTime
}

func (os *OracleService) AggInterval() int {
	return os.aggInterval
}

func (os *OracleService) Config() *types.OracleServiceConfig {
	return os.config
}

func (os *OracleService) GetPrice(symbol string) decimal.Decimal {
	os.lock.RLock()
	defer os.lock.RUnlock()
	return os.prices[symbol]
}

func (os *OracleService) GetPrices() types.PriceBySymbol {
	os.lock.RLock()
	defer os.lock.RUnlock()
	// todo: double check we need return a copy.
	return os.prices
}

func (os *OracleService) UpdatePrice(symbol string, price decimal.Decimal) {
	os.lock.Lock()
	defer os.lock.Unlock()
	os.prices[symbol] = price
}

func (os *OracleService) CheckAndUpdatePrices() {
	for _, s := range os.symbols {
		if os.tradePool.TradeUpdated(s) {
			trs, err := os.tradePool.ConsumeTrades(s)
			if err != nil {
				// logger warning.
			}
			if len(trs) != 0 {
				price, err := os.aggregator.Aggregate(trs)
				if err != nil {
					// logger warning.
				}
				os.UpdatePrice(s, price)
			}
		}
	}
}

func (os *OracleService) Stop() {
	os.doneCh <- struct{}{}
	for _, a := range os.adapters {
		a.Stop()
	}
}

func (os *OracleService) Start() {

	// todo: create adapter instances by according to configuration.
	os.adapters = append(os.adapters, crypto_provider.NewBinanceAdapter())
	// start adapters one by one.
	for _, a := range os.adapters {
		a.Initialize(nil, os.tradePool)
		err := a.Start()
		if err != nil {
			// logging errors.
		}
	}

	// start ticker job.
	for {
		select {
		case <-os.doneCh:
			os.ticker.Stop()
			return
		case <-os.ticker.C:
			os.CheckAndUpdatePrices()
		}
	}
}
