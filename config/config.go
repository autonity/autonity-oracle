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
	"math"
	"os"
	"strconv"
)

var (
	SourceScalingFactor      = uint64(10)
	ConfidenceStrategyLinear = 0
	ConfidenceStrategyFixed  = 1

	DefaultConfidenceStrategy = ConfidenceStrategyLinear // 0: linear, 1: fixed.
	DefaultLogVerbosity       = 3                        // 0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error
	DefaultGasTipCap          = uint64(1)
	DefaultAutonityWSUrl      = "ws://127.0.0.1:8546"
	DefaultKeyFile            = "./UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"
	DefaultKeyPassword        = "123"
	DefaultPluginDir          = "./plugins"
	DefaultPluginConfFile     = "./plugins-conf.yml"
	DefaultOracleConfFile     = ""
	// DefaultSymbols are native symbols required by the oracle protocol:
	DefaultSymbols = []string{"AUD-USD", "CAD-USD", "EUR-USD", "GBP-USD", "JPY-USD", "SEK-USD", "ATN-USD", "NTN-USD", "NTN-ATN"}
	// BridgerSymbols are helper symbols to convert the ratio of ATN-USDC & NTN-USDC to the required symbols: ATN-USD, NTN-USD.
	// They are always attached into the sampled symbols set which is used for data sampling.
	BridgerSymbols = []string{"ATN-USDC", "NTN-USDC", "USDC-USD"}

	DefaultSampledSymbols = []string{"AUD-USD", "CAD-USD", "EUR-USD", "GBP-USD", "JPY-USD", "SEK-USD", "ATN-USD", "NTN-USD", "NTN-ATN", "ATN-USDC", "NTN-USDC", "USDC-USD"}

	// ForexCurrencies in this map can be applied with linear or fixed confidence, cryptos take fixed strategy right now.
	ForexCurrencies = map[string]struct{}{
		"AUD-USD": {},
		"CAD-USD": {},
		"EUR-USD": {},
		"GBP-USD": {},
		"JPY-USD": {},
		"SEK-USD": {},
	}
)

const Version uint8 = 19
const OracleDecimals uint8 = 18
const MaxConfidence = 100
const BaseConfidence = 40
const UsageOracleKey = "Set the oracle server key file path."
const UsagePluginConf = "Set the plugin's configuration file path."
const UsageOracleConf = "Set the oracle server configuration file path."
const UsagePluginDir = "Set the directory path of the data plugins."
const UsageOracleKeyPassword = "Set the password to decrypt oracle server key file."
const UsageGasTipCap = "Set the gas priority fee cap to issue the oracle data report transactions."
const UsageWSUrl = "Set the WS-RPC server listening interface and port of the connected Autonity Client node."
const UsageLogLevel = "Set the logging level, available levels are:  0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error"
const UsageConfidenceStrategy = "Set the confidence strategy, available values are:  0: linear, 1: fixed"

func MakeConfig() *types.OracleServiceConfig {
	var logLevel int
	var keyFile string
	var gasTipCap uint64
	var pluginDir string
	var keyPassword string
	var autonityWSUrl string
	var pluginConfFile string
	var oracleConfFile string
	var confidenceStrategy int

	flag.Uint64Var(&gasTipCap, "tip", DefaultGasTipCap, UsageGasTipCap)
	flag.StringVar(&keyFile, "key.file", DefaultKeyFile, UsageOracleKey)
	flag.IntVar(&logLevel, "log.level", DefaultLogVerbosity, UsageLogLevel)
	flag.StringVar(&autonityWSUrl, "ws", DefaultAutonityWSUrl, UsageWSUrl)
	flag.StringVar(&pluginDir, "plugin.dir", DefaultPluginDir, UsagePluginDir)
	flag.StringVar(&pluginConfFile, "plugin.conf", DefaultPluginConfFile, UsagePluginConf)
	flag.StringVar(&keyPassword, "key.password", DefaultKeyPassword, UsageOracleKeyPassword)
	flag.StringVar(&oracleConfFile, flag.DefaultConfigFlagname, DefaultOracleConfFile, UsageOracleConf)
	flag.IntVar(&confidenceStrategy, "confidence.strategy", DefaultConfidenceStrategy, UsageConfidenceStrategy)

	flag.Parse()
	if len(flag.Args()) == 1 && flag.Args()[0] == "version" {
		log.SetFlags(0)
		log.Println(VersionString(Version))
		os.Exit(0)
	}

	// configs which are not set in the CLI flags or in a config file, should be set by system environment variables.
	// if they are not set by below environment variables, the default values are applied.
	if lvl, presented := os.LookupEnv(types.EnvLogLevel); presented && logLevel == DefaultLogVerbosity {
		l, err := strconv.Atoi(lvl)
		if err != nil {
			log.Printf("wrong log level configed in $LOG_LEVEL")
			helpers.PrintUsage()
			os.Exit(1)
		}
		logLevel = l
		if hclog.Level(logLevel) < hclog.NoLevel || hclog.Level(logLevel) > hclog.Error { //nolint
			log.Printf("wrong logging level configed %d, %s", logLevel, UsageLogLevel)
			helpers.PrintUsage()
			os.Exit(1)
		}
	}

	if cs, presented := os.LookupEnv(types.EnvConfidenceStrategy); presented && confidenceStrategy == DefaultConfidenceStrategy {
		strategy, err := strconv.Atoi(cs)
		if err != nil {
			log.Printf("wrong confidence strategy configed in $CONFIDENCE_STRATEGY")
			helpers.PrintUsage()
			os.Exit(1)
		}
		confidenceStrategy = strategy
		if confidenceStrategy < DefaultConfidenceStrategy || confidenceStrategy > ConfidenceStrategyFixed {
			log.Printf("wrong confidence strategy configed %d, %s", confidenceStrategy, UsageConfidenceStrategy)
			helpers.PrintUsage()
			os.Exit(1)
		}
	}

	if pluginBase, presented := os.LookupEnv(types.EnvPluginDIR); presented && pluginDir == DefaultPluginDir {
		pluginDir = pluginBase
	}

	if k, presented := os.LookupEnv(types.EnvKeyFile); presented && keyFile == DefaultKeyFile {
		keyFile = k
	}

	if password, presented := os.LookupEnv(types.EnvKeyFilePASS); presented && keyPassword == DefaultKeyPassword {
		keyPassword = password
	}

	if ws, presented := os.LookupEnv(types.EnvWS); presented && autonityWSUrl == DefaultAutonityWSUrl {
		autonityWSUrl = ws
	}

	if pluginConf, presented := os.LookupEnv(types.EnvPluginCof); presented && pluginConfFile == DefaultPluginConfFile {
		pluginConfFile = pluginConf
	}

	if capGasTip, presented := os.LookupEnv(types.EnvGasTipCap); presented && gasTipCap == DefaultGasTipCap {
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

	return &types.OracleServiceConfig{
		GasTipCap:          gasTipCap,
		Key:                key,
		AutonityWSUrl:      autonityWSUrl,
		PluginDIR:          pluginDir,
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

func VersionString(version uint8) string {
	major := version / 100
	minor := (version / 10) % 10
	patch := version % 10
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

// ComputeConfidence calculates the confidence weight based on the number of data samples. Note! Cryptos take
// fixed strategy as we have very limited number of data sources at the genesis phase. Thus, the confidence
// computing is just for forex currencies for the time being.
func ComputeConfidence(symbol string, numOfSamples, strategy int) uint8 {

	// Todo: once the community have more extensive AMM and DEX markets, we will remove this to enable linear
	//  strategy as well for cryptos.
	if _, is := ForexCurrencies[symbol]; !is {
		return MaxConfidence
	}

	// Forex currencies with fixed strategy.
	if strategy == ConfidenceStrategyFixed {
		return MaxConfidence
	}

	// Forex currencies with linear strategy.
	weight := BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, float64(numOfSamples)))

	if weight > MaxConfidence {
		weight = MaxConfidence
	}

	return uint8(weight) //nolint
}
