package config

import (
	"autonity-oracle/types"
	"os"
	"strconv"
	"strings"
)

var (
	EnvHTTPPort      = "ORACLE_HTTP_PORT"
	EnvCryptoSymbols = "ORACLE_CRYPTO_SYMBOLS"

	DefaultSymbols = "NTNUSDT,NTNUSDC,NTNBTC,NTNETH"
	DefaultPort    = 30311
)

func MakeConfig() *types.OracleServiceConfig {
	port := resolvePort()
	symbols := resolveSymbols()
	return &types.OracleServiceConfig{
		Symbols:  symbols,
		HTTPPort: port,
	}
}

func resolvePort() int {
	p, ok := os.LookupEnv(EnvHTTPPort)
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
		if len(symbol) == 0 {
			continue
		}
		result = append(result, symbol)
	}
	return result
}
