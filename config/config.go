package config

import (
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/hashicorp/go-hclog"
	"github.com/namsral/flag"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strconv"
)

var (
	defaultLogVerbosity           = 3 // 0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error
	defaultGasTipCap              = uint64(1)
	defaultAutonityWSUrl          = "ws://127.0.0.1:8546"
	defaultKeyFile                = "./UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"
	defaultKeyPassword            = "123"
	defaultPluginDir              = "./plugins"
	DefaultProfileDir             = "."
	defaultPluginConfFile         = "./plugins-conf.yml"
	defaultOracleConfFile         = ""
	defaultVoteBufferAfterPenalty = uint64(3600 * 24) // The buffering time window in blocks to continue vote after the last penalty event.

	ConfidenceStrategyLinear  = 0
	ConfidenceStrategyFixed   = 1
	defaultConfidenceStrategy = ConfidenceStrategyLinear // 0: linear, 1: fixed.

	envKeyFile            = "KEY_FILE"
	envKeyFilePASS        = "KEY_PASSWORD"
	envPluginCof          = "PLUGIN_CONF"
	envPluginDIR          = "PLUGIN_DIR"
	envWS                 = "AUTONITY_WS"
	envGasTipCap          = "GAS_TIP_CAP"
	envLogLevel           = "LOG_LEVEL"
	envProfDIR            = "PROFILE_DIR"
	envConfidenceStrategy = "CONFIDENCE_STRATEGY"
	envVoteBuffer         = "VOTE_BUFFER"
)

// Version number of the oracle server in uint8. It is required
// for data reporting interface to collect oracle clients version.
const Version uint8 = 23

const usageOracleKey = "Set the oracle server key file path."
const usagePluginConf = "Set the plugin's configuration file path."
const usageOracleConf = "Set the oracle server configuration file path."
const usagePluginDir = "Set the directory path of the data plugins."
const usageProfileDir = "Set the directory path to dump profile data."
const usageOracleKeyPassword = "Set the password to decrypt oracle server key file."
const usageGasTipCap = "Set the gas priority fee cap to issue the oracle data report transactions."
const usageWSUrl = "Set the WS-RPC server listening interface and port of the connected Autonity Client node."
const usageLogLevel = "Set the logging level, available levels are:  0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error"
const usageConfidenceStrategy = "Set the confidence strategy, available values are:  0: linear, 1: fixed"
const usageVoteBuffer = "Set the buffering time window in blocks to continue vote after the last penalty event. Default value is 86400 (1 day)."

// PluginConfig carry the configuration of plugins.
type PluginConfig struct {
	Name               string `json:"name" yaml:"name"`                         // the name of the plugin binary.
	Key                string `json:"key" yaml:"key"`                           // the API key granted by your data provider to access their data API.
	Scheme             string `json:"scheme" yaml:"scheme"`                     // the data service scheme, http or https.
	Endpoint           string `json:"endpoint" yaml:"endpoint"`                 // the data service endpoint url of the data provider.
	Timeout            int    `json:"timeout" yaml:"timeout"`                   // the timeout period in seconds that an API request is lasting for.
	DataUpdateInterval int    `json:"refresh" yaml:"refresh"`                   // the interval in seconds to fetch data from data provider due to rate limit.
	NTNTokenAddress    string `json:"ntnTokenAddress" yaml:"ntnTokenAddress"`   // The NTN erc20 token address on the target blockchain.
	ATNTokenAddress    string `json:"atnTokenAddress" yaml:"atnTokenAddress"`   // The Wrapped ATN erc20 token address on the target blockchain.
	USDCTokenAddress   string `json:"usdcTokenAddress" yaml:"usdcTokenAddress"` // USDC erc20 token address on the target blockchain.
	SwapAddress        string `json:"swapAddress" yaml:"swapAddress"`           // UniSwap factory contract address or AirSwap SwapERC20 contract address on the target blockchain.
}

func MakeConfig() *types.OracleServerConfig {
	var logLevel int
	var keyFile string
	var gasTipCap uint64
	var pluginDir string
	var profileDir string
	var keyPassword string
	var autonityWSUrl string
	var pluginConfFile string
	var oracleConfFile string
	var confidenceStrategy int
	var voteBuffer uint64

	flag.Uint64Var(&gasTipCap, "tip", defaultGasTipCap, usageGasTipCap)
	flag.StringVar(&keyFile, "key.file", defaultKeyFile, usageOracleKey)
	flag.IntVar(&logLevel, "log.level", defaultLogVerbosity, usageLogLevel)
	flag.StringVar(&autonityWSUrl, "ws", defaultAutonityWSUrl, usageWSUrl)
	flag.StringVar(&pluginDir, "plugin.dir", defaultPluginDir, usagePluginDir)
	flag.StringVar(&profileDir, "profile.dir", DefaultProfileDir, usageProfileDir)
	flag.StringVar(&pluginConfFile, "plugin.conf", defaultPluginConfFile, usagePluginConf)
	flag.StringVar(&keyPassword, "key.password", defaultKeyPassword, usageOracleKeyPassword)
	flag.Uint64Var(&voteBuffer, "vote.buffer", defaultVoteBufferAfterPenalty, usageVoteBuffer)
	flag.StringVar(&oracleConfFile, flag.DefaultConfigFlagname, defaultOracleConfFile, usageOracleConf)
	flag.IntVar(&confidenceStrategy, "confidence.strategy", defaultConfidenceStrategy, usageConfidenceStrategy)

	flag.Parse()
	if len(flag.Args()) == 1 && flag.Args()[0] == "version" {
		log.SetFlags(0)
		log.Println(VersionString(Version))
		os.Exit(0)
	}

	// configs which are not set in the CLI flags or in a config file, should be set by system environment variables.
	// if they are not set by below environment variables, the default values are applied.
	if lvl, presented := os.LookupEnv(envLogLevel); presented && logLevel == defaultLogVerbosity {
		l, err := strconv.Atoi(lvl)
		if err != nil {
			log.Printf("wrong log level configed in $LOG_LEVEL")
			helpers.PrintUsage()
			os.Exit(1)
		}
		logLevel = l
		if hclog.Level(logLevel) < hclog.NoLevel || hclog.Level(logLevel) > hclog.Error { //nolint
			log.Printf("wrong logging level configed %d, %s", logLevel, usageLogLevel)
			helpers.PrintUsage()
			os.Exit(1)
		}
	}

	// Try to resolve confidence strategy from environment variable.
	if cs, presented := os.LookupEnv(envConfidenceStrategy); presented && confidenceStrategy == defaultConfidenceStrategy {
		strategy, err := strconv.Atoi(cs)
		if err != nil {
			log.Printf("wrong confidence strategy configed in $CONFIDENCE_STRATEGY")
			helpers.PrintUsage()
			os.Exit(1)
		}
		confidenceStrategy = strategy
		if confidenceStrategy < defaultConfidenceStrategy || confidenceStrategy > ConfidenceStrategyFixed {
			log.Printf("wrong confidence strategy configed %d, %s", confidenceStrategy, usageConfidenceStrategy)
			helpers.PrintUsage()
			os.Exit(1)
		}
	}

	if vb, presented := os.LookupEnv(envVoteBuffer); presented && voteBuffer == defaultVoteBufferAfterPenalty {
		voteBuff, err := strconv.ParseUint(vb, 0, 64)
		if err != nil {
			log.Printf("wrong value configed in $VOTE_BUFFER")
			helpers.PrintUsage()
			os.Exit(1)
		}
		voteBuffer = voteBuff
	}

	if pDir, presented := os.LookupEnv(envProfDIR); presented && profileDir == DefaultProfileDir {
		profileDir = pDir
	}

	if pluginBase, presented := os.LookupEnv(envPluginDIR); presented && pluginDir == defaultPluginDir {
		pluginDir = pluginBase
	}

	if k, presented := os.LookupEnv(envKeyFile); presented && keyFile == defaultKeyFile {
		keyFile = k
	}

	if password, presented := os.LookupEnv(envKeyFilePASS); presented && keyPassword == defaultKeyPassword {
		keyPassword = password
	}

	if ws, presented := os.LookupEnv(envWS); presented && autonityWSUrl == defaultAutonityWSUrl {
		autonityWSUrl = ws
	}

	if pluginConf, presented := os.LookupEnv(envPluginCof); presented && pluginConfFile == defaultPluginConfFile {
		pluginConfFile = pluginConf
	}

	if capGasTip, presented := os.LookupEnv(envGasTipCap); presented && gasTipCap == defaultGasTipCap {
		gasTip, err := strconv.ParseUint(capGasTip, 0, 64)
		if err != nil {
			log.Printf("wrong value configed in $GAS_TIP_CAP")
			helpers.PrintUsage()
			os.Exit(1)
		}
		gasTipCap = gasTip
	}

	key, err := loadKey(keyFile, keyPassword)
	if err != nil {
		helpers.PrintUsage()
		os.Exit(1)
	}

	return &types.OracleServerConfig{
		VoteBuffer:         voteBuffer,
		GasTipCap:          gasTipCap,
		Key:                key,
		AutonityWSUrl:      autonityWSUrl,
		PluginDIR:          pluginDir,
		ProfileDir:         profileDir,
		PluginConfFile:     pluginConfFile,
		LoggingLevel:       hclog.Level(logLevel), //nolint
		ConfidenceStrategy: confidenceStrategy,
	}
}

func loadKey(keyFile, password string) (*keystore.Key, error) {
	keyJson, err := os.ReadFile(keyFile)
	if err != nil {
		log.Printf("cannot read key from oracle key file: %s, %s", keyFile, err.Error())
		return nil, err
	}

	key, err := keystore.DecryptKey(keyJson, password)
	if err != nil {
		log.Printf("cannot decrypt oracle key file: %s, with the provided password!", keyFile)
		return nil, err
	}
	return key, nil
}

func LoadPluginsConfig(file string) (map[string]PluginConfig, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var configs []PluginConfig
	err = yaml.Unmarshal(content, &configs)
	if err != nil {
		return nil, err
	}

	confs := make(map[string]PluginConfig)
	for _, conf := range configs {
		c := conf
		confs[c.Name] = c
	}

	return confs, nil
}

func VersionString(version uint8) string {
	major := version / 100
	minor := (version / 10) % 10
	patch := version % 10
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}
