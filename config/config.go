package config

import (
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/namsral/flag"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

var (
	DefaultGasTipCap      = uint64(1)
	DefaultAutonityWSUrl  = "ws://127.0.0.1:8546"
	DefaultKeyFile        = "./test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"
	DefaultKeyPassword    = "123"
	DefaultPluginDir      = "./build/bin/plugins"
	DefaultPluginConfFile = "./build/bin/plugins/plugins-conf.yml"
	DefaultSymbols        = "NTN/USD,NTN/AUD,NTN/CAD,NTN/EUR,NTN/GBP,NTN/JPY,NTN/SEK"
)

const Version = "v0.0.2"

func MakeConfig() *types.OracleServiceConfig {
	var keyFile string
	var symbols string
	var pluginDir string
	var keyPassword string
	var autonityWSUrl string
	var pluginConfFile string
	var gasTipCap uint64

	flag.Uint64Var(&gasTipCap, "oracle_gas_tip_cap", DefaultGasTipCap, "The gas priority fee cap to issue the oracle data report transactions")
	flag.StringVar(&pluginDir, "oracle_plugin_dir", DefaultPluginDir, "The DIR where the adapter plugins are stored")
	flag.StringVar(&symbols, "oracle_symbols", DefaultSymbols, "The symbols string separated by comma")
	flag.StringVar(&keyFile, "oracle_key_file", DefaultKeyFile, "Oracle server key file")
	flag.StringVar(&autonityWSUrl, "oracle_autonity_ws_url", DefaultAutonityWSUrl, "WS-RPC server listening interface and port of the connected Autonity Go Client node")
	flag.StringVar(&keyPassword, "oracle_key_password", DefaultKeyPassword, "Password to the oracle server key file")
	flag.StringVar(&pluginConfFile, "oracle_plugin_conf", DefaultPluginConfFile, "The plugins' configuration file in YAML")

	flag.Parse()
	if len(flag.Args()) == 1 && flag.Args()[0] == "version" {
		log.SetFlags(0)
		log.Println(Version)
		os.Exit(0)
	}

	symbolArray := helpers.ParseSymbols(symbols)

	keyJson, err := os.ReadFile(keyFile)
	if err != nil {
		panic(fmt.Sprintf("invalid key file: %s", keyFile))
	}

	key, err := keystore.DecryptKey(keyJson, keyPassword)
	if err != nil {
		panic("cannot open keyfile with provided password")
	}

	return &types.OracleServiceConfig{
		GasTipCap:      gasTipCap,
		Key:            key,
		AutonityWSUrl:  autonityWSUrl,
		Symbols:        symbolArray,
		PluginDIR:      pluginDir,
		PluginConfFile: pluginConfFile,
	}
}

func LoadPluginsConfig(file string) (map[string]types.PluginConfig, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var configs []types.PluginConfig
	err = yaml.Unmarshal(content, &configs)
	if err != nil {
		return nil, err
	}

	confs := make(map[string]types.PluginConfig)
	for _, conf := range configs {
		c := conf
		confs[c.Name] = c
	}

	return confs, nil
}
