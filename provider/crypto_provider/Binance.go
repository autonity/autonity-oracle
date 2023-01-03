package crypto_provider

import (
	"autonity-oralce/types"
	"log"
	"time"
)

// Since the data provider of Autonity cryptos is not being clarified, for the time being we assume that Binance might
// be the provider, and in the FetchPrices interface, we simulate data points for the symbols we want to have.

type BinanceAdapter struct {
	pricePool types.PricePool

	aliveAt int64
	alive   bool
	url     string
}

func NewBinanceAdapter() *BinanceAdapter {
	return &BinanceAdapter{}
}

func (ba *BinanceAdapter) Name() string {
	return "Binance"
}

func (ba *BinanceAdapter) Version() string {
	return "0.0.1"
}

func (ba *BinanceAdapter) Alive() bool {
	return ba.alive
}

func (ba *BinanceAdapter) Initialize(pricePool types.PricePool) {
	ba.pricePool = pricePool
}

// todo: fetch prices by symbols from provider and push those prices into price pool once provider is clarified.
func (ba *BinanceAdapter) FetchPrices(symbols []string) error {
	// some fake data is simulated here since none data provider is clarified.
	log.Printf("fetching data prices from provider: %s\n", ba.Name())
	var prices []types.Price
	for _, s := range symbols {
		p := types.Price{
			Timestamp: time.Now().UnixMilli(),
			Symbol:    s,
			Price:     types.SimulatedPrice,
		}
		prices = append(prices, p)
	}

	// push data to price pool
	ba.pricePool.AddPrices(prices)
	return nil
}
