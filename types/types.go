package types

import (
	"encoding/json"
	"errors"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
	"math/big"
)

var (
	EnvCryptoSymbols        = "ORACLE_CRYPTO_SYMBOLS"
	EnvPluginDIR            = "ORACLE_PLUGIN_DIR"
	EnvKeyFile              = "ORACLE_KEY_FILE"
	EnvKeyFilePASS          = "ORACLE_KEY_PASSWORD"
	SimulatedPrice          = decimal.RequireFromString("11.11")
	InvalidPrice            = new(big.Int).Sub(math.BigPow(2, 255), big.NewInt(1))
	InvalidSalt             = big.NewInt(0)
	Deployer                = common.Address{}
	AutonityContractAddress = crypto.CreateAddress(Deployer, 0)
	OracleContractAddress   = crypto.CreateAddress(Deployer, 1)

	ErrPeerOnSync        = errors.New("l1 node is on peer sync")
	ErrNoAvailablePrice  = errors.New("no available prices collected yet")
	ErrNoSymbolsObserved = errors.New("no symbols observed from oracle contract")
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
	Key           *keystore.Key
	AutonityWSUrl string
	Symbols       []string
	PluginDIR     string
}

// JSONRPCMessage is the JSON spec to carry those data response from the binance data simulator.
type JSONRPCMessage struct {
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}
