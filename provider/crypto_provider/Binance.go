package crypto_provider

import (
	"autonity-oralce/types"
	"errors"
)

var ErrInvalidResponse = errors.New("invalid http response")

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

func (ba *BinanceAdapter) FetchPrices(symbols []string) error {
	// todo: fetch prices by symbols from provider and push those prices into price pool
	return nil
}
