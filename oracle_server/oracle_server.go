package oracle_server

import (
	"autonity-oracle/aggregator"
	"autonity-oracle/price_pool"
	"autonity-oracle/provider/crypto_provider"
	"autonity-oracle/types"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"log"
	"sync"
	"time"
)

var (
	ValidDataAge   = 3 * 60 * 1000 // 3 minutes, data fetched within 3 minutes are valid to update the price.
	Version        = "0.0.1"
	UpdateInterval = 10 * time.Second // 10s, the data fetching interval for the oracle server's ticker job.
)

type OracleServer struct {
	version string

	lock sync.RWMutex

	prices types.PriceBySymbol // aggregated prices which is referred by http data service to provide the data service.

	doneCh chan struct{}
	ticker *time.Ticker // the clock source to trigger the 10s interval job.

	symbols           []string                      // the symbols for data fetching in oracle service.
	aggregator        types.Aggregator              // the price aggregator once we have multiple data providers.
	priceProviderPool *price_pool.PriceProviderPool // the price pool organized by provider and by symbols
	adapters          []types.Adapter               // the adaptors which adapts with different data providers.
}

func NewOracleServer(symbols []string) *OracleServer {
	os := &OracleServer{
		version:           Version,
		symbols:           symbols,
		prices:            make(types.PriceBySymbol),
		doneCh:            make(chan struct{}),
		ticker:            time.NewTicker(UpdateInterval),
		aggregator:        aggregator.NewAveragePriceAggregator(),
		priceProviderPool: price_pool.NewPriceProviderPool(),
	}

	// todo: create adapters for all the providers once we have the providers clarified.
	adapter := crypto_provider.NewBinanceAdapter()
	pool := os.priceProviderPool.AddPriceProvider(adapter.Name())
	os.adapters = append(os.adapters, adapter)
	adapter.Initialize(pool)

	return os
}

func (os *OracleServer) Version() string {
	return os.version
}

func (os *OracleServer) UpdateSymbols(symbols []string) {
	os.symbols = symbols
}

func (os *OracleServer) Symbols() []string {
	return os.symbols
}

func (os *OracleServer) GetPrices() types.PriceBySymbol {
	os.lock.RLock()
	defer os.lock.RUnlock()
	return os.prices
}

func (os *OracleServer) UpdatePrice(price types.Price) {
	os.lock.Lock()
	defer os.lock.Unlock()
	os.prices[price.Symbol] = price
}

func (os *OracleServer) UpdatePrices() {
	wg := &errgroup.Group{}
	for _, ad := range os.adapters {
		provider := ad
		wg.Go(func() error {
			return provider.FetchPrices(os.symbols)
		})
	}
	err := wg.Wait()
	if err != nil {
		log.Printf("error %s occurs when fetching prices from provider", err.Error())
	}

	now := time.Now().UnixMilli()

	for _, s := range os.symbols {
		var prices []decimal.Decimal
		for _, ad := range os.adapters {
			p, err := os.priceProviderPool.GetPriceProvider(ad.Name()).GetPrice(s)
			if err != nil {
				continue
			}
			// only those price collected within 3 minutes are valid.
			if now-p.Timestamp < int64(ValidDataAge) && now >= p.Timestamp {
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

		os.UpdatePrice(price)
	}
}

func (os *OracleServer) Stop() {
	os.doneCh <- struct{}{}
}

func (os *OracleServer) Start() {
	// start the ticker job to fetch prices for all the symbols from all adapters on every 10s.
	for {
		select {
		case <-os.doneCh:
			os.ticker.Stop()
			log.Println("the ticker job for data update is stopped")
			return
		case <-os.ticker.C:
			os.UpdatePrices()
		}
	}
}
