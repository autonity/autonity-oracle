package types

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"math/big"
)

var (
	EnvPluginDIR            = "PLUGIN_DIR"
	EnvKeyFile              = "KEY_FILE"
	EnvKeyFilePASS          = "KEY_PASSWORD"
	EnvWS                   = "AUTONITY_WS"
	EnvPluginCof            = "PLUGIN_CONF"
	EnvGasTipCap            = "GAS_TIP_CAP"
	EnvLogLevel             = "LOG_LEVEL"
	SimulatedPrice          = decimal.RequireFromString("11.11")
	InvalidPrice            = new(big.Int).Sub(math.BigPow(2, 255), big.NewInt(1))
	InvalidSalt             = big.NewInt(0)
	Deployer                = common.Address{}
	AutonityContractAddress = crypto.CreateAddress(Deployer, 0)
	OracleContractAddress   = crypto.CreateAddress(Deployer, 2)

	ErrPeerOnSync        = errors.New("l1 node is on peer sync")
	ErrNoAvailablePrice  = errors.New("no available prices collected yet")
	ErrNoDataRound       = errors.New("no data collected at current round")
	ErrNoSymbolsObserved = errors.New("no symbols observed from oracle contract")
	ErrMissingServiceKey = errors.New("the key to access the data source is missing, please check the plugin config")
)

// MaxBufferedRounds is the number of round data to be buffered.
const MaxBufferedRounds = 10

// Price is the structure contains the exchange rate of a symbol with a timestamp at which the sampling happens.
type Price struct {
	Timestamp int64 // TS on when the data is being sampled in time's seconds since Jan 1 1970 (Unix time).
	Symbol    string
	Price     decimal.Decimal
}

// PriceBySymbol group the price by symbols.
type PriceBySymbol map[string]Price

// RoundData contains the aggregated price by symbols for a round with those ordered symbols and a corresponding salt to
// compute the round commitment hash.
type RoundData struct {
	RoundID        uint64
	Tx             *types.Transaction
	Salt           *big.Int
	CommitmentHash common.Hash
	Prices         PriceBySymbol
	Symbols        []string
}

// OracleServiceConfig is the configuration of the oracle client.
type OracleServiceConfig struct {
	LoggingLevel   hclog.Level
	GasTipCap      uint64
	Key            *keystore.Key
	AutonityWSUrl  string
	PluginDIR      string
	PluginConfFile string
}

// JSONRPCMessage is the JSON spec to carry those data response from the binance data simulator.
type JSONRPCMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// SampleEvent carry the symbols and TS for the data sampling.
type SampleEvent struct {
	Symbols []string
	TS      int64
}

// PluginConfig carry the configuration of plugins.
type PluginConfig struct {
	Name               string `json:"name" yaml:"name"`                         // the name of the plugin binary.
	Key                string `json:"key" yaml:"key"`                           // the API key granted by your data provider to access their data API.
	Scheme             string `json:"scheme" yaml:"scheme"`                     // the data service scheme, http or https.
	Endpoint           string `json:"endpoint" yaml:"endpoint"`                 // the data service endpoint url of the data provider.
	Timeout            int    `json:"timeout" yaml:"timeout"`                   // the timeout period in seconds that an API request is lasting for.
	DataUpdateInterval int    `json:"refresh" yaml:"refresh"`                   // the interval in seconds to fetch data from data provider due to rate limit.
	ATNTokenAddress    string `json:"atnTokenAddress" yaml:"atnTokenAddress"`   // The Wrapped ATN erc20 token address.
	USDCTokenAddress   string `json:"usdcTokenAddress" yaml:"usdcTokenAddress"` // USDC erc20 token address.
	SwapAddress        string `json:"swapAddress" yaml:"swapAddress"`           // uniswap factory contract address or airswap SwapERC20 contract address.
}
