package crypto_provider

import (
	"autonity-oralce/types"
)

type BinanceAdapter struct {
	aliveAt   uint64
	tradePool types.TradePool
	config    *types.AdapterConfig
}

func (ba *BinanceAdapter) Name() string {
	return "Binance"
}

func (ba *BinanceAdapter) Version() string {
	return "0.0.1"
}

func (ba *BinanceAdapter) Initialize(config *types.AdapterConfig, tradePool types.TradePool) error {
	// todo: check config.
	ba.config = config
	ba.tradePool = tradePool
	return nil
}

func (ba *BinanceAdapter) Start() error {
	// todo: start the ticker routine
	return nil
}

func (ba *BinanceAdapter) Stop() error {
	// todo: stop the ticker routine
	return nil
}

func (ba *BinanceAdapter) Symbols() []string {
	// todo: return all the symbols supported or configured in this adapter.
	return nil
}

func (ba *BinanceAdapter) Alive() bool {
	return true
}

func (ba *BinanceAdapter) Config() *types.AdapterConfig {
	return ba.config
}

// TODO: implement the adapter which fetch trades via pull mode and a push mode from binance endpoint.
