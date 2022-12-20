package common

import (
	"autonity-oralce/types"
)

type BasicAdapter struct {
	config  types.AdapterConfig
	symbols []string
	isAlive bool
}

func (ba *BasicAdapter) Init(config types.AdapterConfig) {
	ba.config = config
	//todo: init symbols from config
}

func (ba *BasicAdapter) GetSymbols() []string {
	return ba.symbols
}

func (ba *BasicAdapter) AddSymbol(symbol string) {
	ba.symbols = append(ba.symbols, symbol)
}

func (ba *BasicAdapter) IsAlive() bool {
	return ba.isAlive
}
