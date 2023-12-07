package config

import (
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/hashicorp/go-hclog"
	"github.com/namsral/flag"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strconv"
)

var (
	DefaultLogVerbosity   = 2 // 0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error
	DefaultGasTipCap      = uint64(1)
	DefaultAutonityWSUrl  = "ws://127.0.0.1:8546"
	DefaultKeyFile        = "./UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"
	DefaultKeyPassword    = "123"
	DefaultPluginDir      = "./plugins"
	DefaultPluginConfFile = "./plugins-conf.yml"
	DefaultSymbols        = []string{"AUD-USD", "CAD-USD", "EUR-USD", "GBP-USD", "JPY-USD", "SEK-USD", "ATN-USD", "NTN-USD", "NTN-ATN"}
)

const Version = "v0.1.5"

func MakeConfig() *types.OracleServiceConfig {
	var logLevel int
	var keyFile string
	var pluginDir string
	var keyPassword string
	var autonityWSUrl string
	var pluginConfFile string
	var gasTipCap uint64

	flag.IntVar(&logLevel, "log.level", DefaultLogVerbosity, "Set the logging level, available levels are:  0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error")
	flag.Uint64Var(&gasTipCap, "tip", DefaultGasTipCap, "Set the gas priority fee cap to issue the oracle data report transactions.")
	flag.StringVar(&pluginDir, "plugin.dir", DefaultPluginDir, "Set the directory of the data plugins.")
	flag.StringVar(&keyFile, "key.file", DefaultKeyFile, "Set oracle server key file")
	flag.StringVar(&keyPassword, "key.password", DefaultKeyPassword, "Set the password to decrypt oracle server key file")
	flag.StringVar(&autonityWSUrl, "ws", DefaultAutonityWSUrl, "Set the WS-RPC server listening interface and port of the connected Autonity Client node")
	flag.StringVar(&pluginConfFile, "plugin.conf", DefaultPluginConfFile, "Set the plugins' configuration file")

	flag.Parse()
	if len(flag.Args()) == 1 && flag.Args()[0] == "version" {
		log.SetFlags(0)
		log.Println(Version)
		os.Exit(0)
	}

	if lvl, presented := os.LookupEnv(types.EnvLogLevel); presented {
		l, err := strconv.Atoi(lvl)
		if err != nil {
			log.Printf("Wrong LOG_LEVEL configed by system environment variable")
			helpers.PrintUsage()
			os.Exit(1)
		}
		logLevel = l
	}

	if pluginBase, presented := os.LookupEnv(types.EnvPluginDIR); presented {
		pluginDir = pluginBase
	}

	if k, presented := os.LookupEnv(types.EnvKeyFile); presented {
		keyFile = k
	}

	if password, presented := os.LookupEnv(types.EnvKeyFilePASS); presented {
		keyPassword = password
	}

	if ws, presented := os.LookupEnv(types.EnvWS); presented {
		autonityWSUrl = ws
	}

	if pluginConf, presented := os.LookupEnv(types.EnvPluginCof); presented {
		pluginConfFile = pluginConf
	}

	if capGasTip, presented := os.LookupEnv(types.EnvGasTipCap); presented {
		gasTip, err := strconv.ParseUint(capGasTip, 0, 64)
		if err != nil {
			log.Printf("Wrong GAS_TIP_CAP configed by system environment variable")
			helpers.PrintUsage()
			os.Exit(1)
		}
		gasTipCap = gasTip
	}

	// verify configurations.
	keyJson, err := os.ReadFile(keyFile)
	if err != nil {
		log.Printf("Cannot read key file: %s, %s", keyFile, err.Error())
		helpers.PrintUsage()
		os.Exit(1)
	}

	key, err := keystore.DecryptKey(keyJson, keyPassword)
	if err != nil {
		log.Printf("Cannot decrypt keyfile: %s, with the provided password!", keyFile)
		helpers.PrintUsage()
		os.Exit(1)
	}

	if hclog.Level(logLevel) < hclog.NoLevel || hclog.Level(logLevel) > hclog.Error {
		log.Printf("Wrong logging level configed %d", logLevel)
		helpers.PrintUsage()
		os.Exit(1)
	}

	return &types.OracleServiceConfig{
		GasTipCap:      gasTipCap,
		Key:            key,
		AutonityWSUrl:  autonityWSUrl,
		PluginDIR:      pluginDir,
		PluginConfFile: pluginConfFile,
		LoggingLevel:   hclog.Level(logLevel),
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
