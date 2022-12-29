package config

import (
	"autonity-oralce/types"
	"os"
	"strconv"
	"strings"
)

var (
	EnvHttpPort      = "ORACLE_HTTP_PORT"
	EnvCryptoSymbols = "ORACLE_CRYPTO_SYMBOLS"

	DefaultSymbols = "BNBBTC,BTCUSDT"
	DefaultPort    = 30309
)

func MakeConfig() *types.OracleServiceConfig {
	port := resolvePort()
	symbols := resolveSymbols()
	return &types.OracleServiceConfig{
		Symbols:  symbols,
		HttpPort: port,
	}
}

func resolvePort() int {
	p, ok := os.LookupEnv(EnvHttpPort)
	if !ok {
		return DefaultPort
	}
	port, err := strconv.Atoi(p)
	if err != nil {
		return DefaultPort
	}
	return port
}

func resolveSymbols() []string {
	symbols, ok := os.LookupEnv(EnvCryptoSymbols)
	if !ok {
		return strings.Split(DefaultSymbols, ",")
	}
	var result []string
	symbs := strings.Split(symbols, ",")
	for _, s := range symbs {
		symbol := strings.TrimSpace(s)
		result = append(result, symbol)
	}
	return result
}
