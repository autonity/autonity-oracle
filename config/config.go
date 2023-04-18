package config

import (
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/namsral/flag"
	"io/ioutil" //nolint
)

var (
	DefaultAutonityWSUrl = "ws://127.0.0.1:8645"
	DefaultKeyFile       = "./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"
	DefaultKeyPassword   = "123"
	DefaultPluginDir     = "./build/bin/plugins"
	DefaultSymbols       = "NTNUSD,NTNAUD,NTNCAD,NTNEUR,NTNGBP,NTNJPY,NTNSEK"
)

func MakeConfig() *types.OracleServiceConfig {
	var keyFile string
	var symbols string
	var pluginDir string
	var keyPassword string
	var autonityWSUrl string

	flag.StringVar(&pluginDir, "oracle_plugin_dir", DefaultPluginDir, "The DIR where the adapter plugins are stored")
	flag.StringVar(&symbols, "oracle_crypto_symbols", DefaultSymbols, "The symbols string separated by comma")
	flag.StringVar(&keyFile, "oracle_key_file", DefaultKeyFile, "The file that save the private key of the oracle client")
	flag.StringVar(&autonityWSUrl, "oracle_autonity_ws_url", DefaultAutonityWSUrl, "The websocket URL of autonity client")
	flag.StringVar(&keyPassword, "oracle_key_password", DefaultKeyPassword, "The password to decode your oracle account's key file")
	flag.Parse()

	symbolArray := helpers.ParseSymbols(symbols)

	keyJson, err := ioutil.ReadFile(keyFile)
	if err != nil {
		panic(fmt.Sprintf("invalid key file: %s", keyFile))
	}

	key, err := keystore.DecryptKey(keyJson, keyPassword)
	if err != nil {
		panic("cannot open keyfile with provided password")
	}

	return &types.OracleServiceConfig{
		Key:           key,
		AutonityWSUrl: autonityWSUrl,
		Symbols:       symbolArray,
		PluginDIR:     pluginDir,
	}
}
