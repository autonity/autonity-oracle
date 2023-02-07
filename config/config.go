package config

import (
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"github.com/namsral/flag"
)

var (
	DefaultPluginDir = "./plugins"
	DefaultSymbols   = "ETHUSDC,ETHUSDT,ETHBTC"
	DefaultPort      = 30311
)

func MakeConfig() *types.OracleServiceConfig {
	var port int
	var symbols string
	var pluginDir string

	flag.StringVar(&pluginDir, "oracle_plugin_dir", DefaultPluginDir, "The DIR where the adapter plugins are stored")
	flag.IntVar(&port, "oracle_http_port", DefaultPort, "The HTTP service port to be bind for oracle service")
	flag.StringVar(&symbols, "oracle_crypto_symbols", DefaultSymbols, "The symbols string separated by comma")
	flag.Parse()

	symbolArray := helpers.ParseSymbols(symbols)

	return &types.OracleServiceConfig{
		Symbols:   symbolArray,
		HTTPPort:  port,
		PluginDIR: pluginDir,
	}
}
