package config

import (
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/namsral/flag"
	"io/ioutil" //nolint
)

var (
	DefaultValidatorAccount = "0x"
	DefaultAutonityWSUrl    = "ws://127.0.0.1:7000"
	DefaultKeyFile          = "a path to your key file"
	DefaultKeyPassword      = "key-password"
	DefaultPluginDir        = "./plugins"
	DefaultSymbols          = "ETHUSDC,ETHUSDT,ETHBTC"
	DefaultPort             = 30311
)

func MakeConfig() *types.OracleServiceConfig {
	var port int
	var keyFile string
	var symbols string
	var pluginDir string
	var keyPassword string
	var autonityWSUrl string
	var validatorAccount string

	flag.IntVar(&port, "oracle_http_port", DefaultPort, "The HTTP service port to be bind for oracle service")
	flag.StringVar(&pluginDir, "oracle_plugin_dir", DefaultPluginDir, "The DIR where the adapter plugins are stored")
	flag.StringVar(&symbols, "oracle_crypto_symbols", DefaultSymbols, "The symbols string separated by comma")
	flag.StringVar(&keyFile, "oracle_key_file", DefaultKeyFile, "The file that save the private key of the oracle client")
	flag.StringVar(&autonityWSUrl, "oracle_autonity_ws_url", DefaultAutonityWSUrl, "The websocket URL of autonity client")
	flag.StringVar(&validatorAccount, "oracle_validator_account", DefaultValidatorAccount, "The account address in HEX string of the validator that this oracle client served for")
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
		ValidatorAccount: common.HexToAddress(validatorAccount),
		Key:              key,
		AutonityWSUrl:    autonityWSUrl,
		Symbols:          symbolArray,
		HTTPPort:         port,
		PluginDIR:        pluginDir,
	}
}
