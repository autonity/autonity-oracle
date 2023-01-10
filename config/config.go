package config

import (
	"autonity-oracle/types"
	"github.com/namsral/flag"
	"strings"
)

var (
	DefaultPluginDir = "./plugins/"
	DefaultSymbols   = "NTNUSDT,NTNUSDC,NTNBTC,NTNETH"
	DefaultPort      = 30311
)

func MakeConfig() *types.OracleServiceConfig {
	var port int
	var symbols string
	var pluginDir string

	flag.StringVar(&pluginDir, "oracle_plugin_dir", DefaultPluginDir, "The DIR where the adapter plugin binary is stored")
	flag.IntVar(&port, "oracle_http_port", DefaultPort, "The HTTP service port to be bind for oracle service")
	flag.StringVar(&symbols, "oracle_crypto_symbols", DefaultSymbols, "The symbols string separated by comma")
	flag.Parse()

	var symbolArray []string
	symbs := strings.Split(symbols, ",")
	for _, s := range symbs {
		symbol := strings.TrimSpace(s)
		if len(symbol) == 0 {
			continue
		}
		symbolArray = append(symbolArray, symbol)
	}

	return &types.OracleServiceConfig{
		Symbols:   symbolArray,
		HTTPPort:  port,
		PluginDIR: pluginDir,
	}
}
