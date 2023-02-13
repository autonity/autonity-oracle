package config

import (
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"github.com/namsral/flag"
)

var (
	DefaultAutonityWSUrl = "ws://127.0.0.1:7000"
	DefaultKeyFile       = "./client.key"
	DefaultPluginDir     = "./plugins"
	DefaultSymbols       = "ETHUSDC,ETHUSDT,ETHBTC"
	DefaultPort          = 30311
)

func MakeConfig() *types.OracleServiceConfig {
	var port int
	var symbols string
	var keyFile string
	var pluginDir string
	var autonityWSUrl string

	flag.StringVar(&pluginDir, "oracle_plugin_dir", DefaultPluginDir, "The DIR where the adapter plugins are stored")
	flag.IntVar(&port, "oracle_http_port", DefaultPort, "The HTTP service port to be bind for oracle service")
	flag.StringVar(&symbols, "oracle_crypto_symbols", DefaultSymbols, "The symbols string separated by comma")
	flag.StringVar(&keyFile, "oracle_key_file", DefaultKeyFile, "The file that save the private key of the oracle client")
	flag.StringVar(&autonityWSUrl, "oracle_autonity_ws_url", DefaultAutonityWSUrl, "The websocket URL of autonity client")
	flag.Parse()

	symbolArray := helpers.ParseSymbols(symbols)

	// todo: resolve the private key from key file, a keystore should be expected to store the key with password.

	return &types.OracleServiceConfig{
		KeyFile:       keyFile,
		AutonityWSUrl: autonityWSUrl,
		Symbols:       symbolArray,
		HTTPPort:      port,
		PluginDIR:     pluginDir,
	}
}
