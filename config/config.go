package config

import (
	"autonity-oracle/helpers"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/hashicorp/go-hclog"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strings"
)

var (
	defaultLogVerbosity           = 3 // 0: NoLevel, 1: Trace, 2:Debug, 3: Info, 4: Warn, 5: Error
	defaultGasTipCap              = uint64(1)
	defaultAutonityWSUrl          = "ws://127.0.0.1:8546"
	defaultKeyFile                = "./UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"
	defaultKeyPassword            = "123"
	defaultPluginDir              = "./plugins"
	DefaultProfileDir             = "."
	defaultVoteBufferAfterPenalty = uint64(3600 * 24) // The buffering time window in blocks to continue vote after the last penalty event.

	ConfidenceStrategyLinear  = 0
	ConfidenceStrategyFixed   = 1
	defaultConfidenceStrategy = ConfidenceStrategyLinear // 0: linear, 1: fixed.
)

// Version number of the oracle server in uint8. It is required
// for data reporting interface to collect oracle clients version.
const Version uint8 = 24
const MetricsNameSpace = "autoracle."

// DefaultConfig are values to be taken when the specific configs are omitted from config file.
var DefaultConfig = ServerConfig{
	LoggingLevel:       defaultLogVerbosity,
	GasTipCap:          defaultGasTipCap,
	VoteBuffer:         defaultVoteBufferAfterPenalty,
	KeyFile:            defaultKeyFile,
	KeyPassword:        defaultKeyPassword,
	AutonityWSUrl:      defaultAutonityWSUrl,
	PluginDIR:          defaultPluginDir,
	ProfileDir:         DefaultProfileDir,
	ConfidenceStrategy: defaultConfidenceStrategy,
	PluginConfigs:      nil,
	MetricConfigs:      DefaultMetricConfig,
}

// DefaultMetricConfig is the default config for metrics used in oracle-server.
var DefaultMetricConfig = MetricConfig{
	// common flags
	InfluxDBEndpoint: "http://localhost:8086",
	InfluxDBTags:     "host=localhost",

	// influxdbv1-specific flags.
	EnableInfluxDB:   false,
	InfluxDBDatabase: "autonity",
	InfluxDBUsername: "test",
	InfluxDBPassword: "test",

	// influxdbv2-specific flags
	EnableInfluxDBV2:     false,
	InfluxDBToken:        "test",
	InfluxDBBucket:       "autonity",
	InfluxDBOrganization: "autonity",
}

// MetricConfig contains the configuration for the metric collection of oracle-server.
type MetricConfig struct {
	// Common configs for influxDB V1 and V2.
	InfluxDBEndpoint string `json:"influxDBEndpoint" yaml:"influxDBEndpoint"`
	InfluxDBTags     string `json:"influxDBTags" yaml:"influxDBTags"`

	// InfluxDB V1 specific configs
	EnableInfluxDB   bool   `json:"enableInfluxDB" yaml:"enableInfluxDB"`
	InfluxDBDatabase string `json:"influxDBDatabase" yaml:"influxDBDatabase"`
	InfluxDBUsername string `json:"influxDBUsername" yaml:"influxDBUsername"`
	InfluxDBPassword string `json:"influxDBPassword" yaml:"influxDBPassword"`

	// InfluxDB V2 specific configs
	EnableInfluxDBV2     bool   `json:"enableInfluxDBV2" yaml:"enableInfluxDBV2"`
	InfluxDBToken        string `json:"influxDBToken" yaml:"influxDBToken"`
	InfluxDBBucket       string `json:"influxDBBucket" yaml:"influxDBBucket"`
	InfluxDBOrganization string `json:"influxDBOrganization" yaml:"influxDBOrganization"`
}

// ServerConfig is the schema of oracle-server's config.
type ServerConfig struct {
	LoggingLevel       int            `json:"logLevel" yaml:"logLevel"`
	GasTipCap          uint64         `json:"gasTipCap" yaml:"gasTipCap"`
	VoteBuffer         uint64         `json:"voteBuffer" yaml:"voteBuffer"`
	KeyFile            string         `json:"keyFile" yaml:"keyFile"`
	KeyPassword        string         `json:"keyPassword" yaml:"keyPassword"`
	AutonityWSUrl      string         `json:"autonityWSUrl" yaml:"autonityWSUrl"`
	PluginDIR          string         `json:"pluginDir" yaml:"pluginDir"`
	ProfileDir         string         `json:"profileDir" yaml:"profileDir"`
	ConfidenceStrategy int            `json:"confidenceStrategy" yaml:"confidenceStrategy"`
	PluginConfigs      []PluginConfig `json:"pluginConfigs" yaml:"pluginConfigs"`
	MetricConfigs      MetricConfig   `json:"metricConfigs" yaml:"metricConfigs"`
}

// PluginConfig is the schema of plugins' config.
type PluginConfig struct {
	Name               string `json:"name" yaml:"name"`                         // The name of the plugin binary.
	Key                string `json:"key" yaml:"key"`                           // The API key granted by your data provider to access their data API.
	Scheme             string `json:"scheme" yaml:"scheme"`                     // The data service scheme, http or https.
	Endpoint           string `json:"endpoint" yaml:"endpoint"`                 // The data service endpoint url of the data provider.
	Timeout            int    `json:"timeout" yaml:"timeout"`                   // The timeout period in seconds that an API request is lasting for.
	DataUpdateInterval int    `json:"refresh" yaml:"refresh"`                   // The interval in seconds to fetch data from data provider due to rate limit.
	NTNTokenAddress    string `json:"ntnTokenAddress" yaml:"ntnTokenAddress"`   // The NTN erc20 token address on the target blockchain.
	ATNTokenAddress    string `json:"atnTokenAddress" yaml:"atnTokenAddress"`   // The Wrapped ATN erc20 token address on the target blockchain.
	USDCTokenAddress   string `json:"usdcTokenAddress" yaml:"usdcTokenAddress"` // The USDC erc20 token address on the target blockchain.
	SwapAddress        string `json:"swapAddress" yaml:"swapAddress"`           // The UniSwap factory contract address or AirSwap SwapERC20 contract address on the target blockchain.
	Disabled           bool   `json:"disabled" yaml:"disabled"`                 // The flag to disable a plugin.
}

// Config is the resolved configuration of the oracle-server.
type Config struct {
	ConfigFile         string
	LoggingLevel       hclog.Level
	GasTipCap          uint64
	VoteBuffer         uint64
	Key                *keystore.Key
	AutonityWSUrl      string
	PluginDIR          string
	ProfileDir         string
	ConfidenceStrategy int
	PluginConfigs      map[string]PluginConfig
	MetricConfigs      MetricConfig
}

func MakeConfig() *Config {
	if len(os.Args) != 2 {
		log.SetFlags(0)
		helpers.PrintUsage()
		os.Exit(1)
	}

	if os.Args[1] == "version" {
		log.SetFlags(0)
		log.Println(VersionString(Version))
		os.Exit(0)
	}

	oracleConfFile := os.Args[1]
	config, err := LoadServerConfig(oracleConfFile)
	if err != nil {
		log.SetFlags(0)
		log.Printf("could not load oracle_server config: %s, err: %s", oracleConfFile, err.Error())
		helpers.PrintUsage()
		os.Exit(1)
	}

	key, err := LoadKey(config.KeyFile, config.KeyPassword)
	if err != nil {
		log.SetFlags(0)
		log.Printf("could not load key from key store: %s with password, err: %s", config.KeyFile, err.Error())
		os.Exit(1)
	}

	if config.MetricConfigs.EnableInfluxDB && config.MetricConfigs.EnableInfluxDBV2 {
		log.SetFlags(0)
		log.Println("There are two metrics engine enabled, please select one: influxDB or influxDBV2")
		os.Exit(1)
	}

	pluginConfigs := make(map[string]PluginConfig)
	for _, conf := range config.PluginConfigs {
		c := conf
		pluginConfigs[c.Name] = c
	}

	return &Config{
		VoteBuffer:         config.VoteBuffer,
		GasTipCap:          config.GasTipCap,
		Key:                key,
		AutonityWSUrl:      config.AutonityWSUrl,
		PluginDIR:          config.PluginDIR,
		ProfileDir:         config.ProfileDir,
		LoggingLevel:       hclog.Level(config.LoggingLevel), //nolint
		ConfidenceStrategy: config.ConfidenceStrategy,
		ConfigFile:         oracleConfFile,
		PluginConfigs:      pluginConfigs,
		MetricConfigs:      config.MetricConfigs,
	}
}

func LoadKey(keyFile, password string) (*keystore.Key, error) {
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

func LoadServerConfig(file string) (*ServerConfig, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	config := DefaultConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %v", err)
	}

	return &config, nil
}

func LoadPluginsConfig(file string) (map[string]PluginConfig, error) {
	serverConf, err := LoadServerConfig(file)
	if err != nil {
		return nil, err
	}

	pluginConfigs := make(map[string]PluginConfig)
	for _, conf := range serverConf.PluginConfigs {
		c := conf
		pluginConfigs[c.Name] = c
	}

	return pluginConfigs, nil
}

func VersionString(version uint8) string {
	major := version / 100
	minor := (version / 10) % 10
	patch := version % 10
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

func SplitTagsFlag(tagsFlag string) map[string]string {
	tags := strings.Split(tagsFlag, ",")
	tagsMap := map[string]string{}

	for _, t := range tags {
		if t != "" {
			kv := strings.Split(t, "=")

			if len(kv) == 2 {
				tagsMap[kv[0]] = kv[1]
			}
		}
	}

	return tagsMap
}
