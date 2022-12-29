package autonity_oralce

import (
	"autonity-oralce/aggregator"
	"autonity-oralce/price_pool"
	"autonity-oralce/provider/crypto_provider"
	"autonity-oralce/types"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

var PERIOD = 3 * 60 * 1000

type OracleService struct {
	version string

	lock sync.RWMutex
	// aggregated prices
	prices types.PriceBySymbol

	doneCh chan struct{}
	ticker *time.Ticker

	symbols           []string
	aggregator        types.Aggregator
	priceProviderPool *price_pool.PriceProviderPool
	adapters          []types.Adapter
}

func NewOracleService(symbols []string) *OracleService {
	os := &OracleService{
		version:           "v0.0.1",
		symbols:           symbols,
		prices:            make(types.PriceBySymbol),
		doneCh:            make(chan struct{}),
		ticker:            time.NewTicker(10 * time.Second),
		aggregator:        aggregator.NewAveragePriceAggregator(),
		priceProviderPool: price_pool.NewPriceProviderPool(),
	}

	// todo: create adapters for autonity cryptos
	pool := os.priceProviderPool.AddPriceProvider("Binance")
	adapter := crypto_provider.NewBinanceAdapter()
	os.adapters = append(os.adapters, adapter)
	adapter.Initialize(pool)

	return os
}

func (os *OracleService) Version() string {
	return os.version
}

func (os *OracleService) UpdateSymbols(symbols []string) {
	os.symbols = symbols
}

func (os *OracleService) Symbols() []string {
	return os.symbols
}

func (os *OracleService) GetPrice(symbol string) types.Price {
	os.lock.RLock()
	defer os.lock.RUnlock()
	return os.prices[symbol]
}

func (os *OracleService) GetPrices() types.PriceBySymbol {
	os.lock.RLock()
	defer os.lock.RUnlock()
	return os.prices
}

func (os *OracleService) UpdatePrice(symbol string, price types.Price) {
	os.lock.Lock()
	defer os.lock.Unlock()
	os.prices[symbol] = price
}

func (os *OracleService) UpdatePrices() {
	wg := &errgroup.Group{}
	for _, ad := range os.adapters {
		wg.Go(func() error {
			return ad.FetchPrices(os.symbols)
		})
	}
	err := wg.Wait()
	if err != nil {
		// todo: logging..
	}

	now := time.Now().UnixMilli()

	for _, s := range os.symbols {
		var prices []decimal.Decimal
		for _, ad := range os.adapters {
			p := os.priceProviderPool.GetPriceProvider(ad.Name()).GetPrice(s)
			// only those price collected within 3 minutes are valid.
			if now-p.Timestamp < int64(PERIOD) && now >= p.Timestamp {
				prices = append(prices, p.Price)
			}
		}

		if len(prices) == 0 {
			continue
		}

		price := types.Price{
			Timestamp: now,
			Price:     prices[0],
			Symbol:    s,
		}

		if len(prices) > 1 {
			p, err := os.aggregator.Aggregate(prices)
			if err != nil {
				continue
			}
			price.Price = p
		}

		os.UpdatePrice(s, price)
	}
}

func (os *OracleService) Stop() {
	os.doneCh <- struct{}{}
}

func (os *OracleService) Start() {
	// start ticker job.
	for {
		select {
		case <-os.doneCh:
			os.ticker.Stop()
			return
		case <-os.ticker.C:
			os.UpdatePrices()
		}
	}
}
