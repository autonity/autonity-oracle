package test

import (
	"autonity-oracle/config"
	"autonity-oracle/helpers"
	oracleserver "autonity-oracle/oracle_server"
	"autonity-oracle/types"
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
	"io/fs"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type AutonityContractGenesis struct {
	Bytecode         string         `json:"bytecode,omitempty" toml:",omitempty"`
	ABI              string         `json:"abi,omitempty" toml:",omitempty"`
	MinBaseFee       uint64         `json:"minBaseFee"`
	EpochPeriod      uint64         `json:"epochPeriod"`
	UnbondingPeriod  uint64         `json:"unbondingPeriod"`
	BlockPeriod      uint64         `json:"blockPeriod"`
	MaxCommitteeSize uint64         `json:"maxCommitteeSize"`
	Operator         common.Address `json:"operator"`
	Treasury         common.Address `json:"treasury"`
	TreasuryFee      uint64         `json:"treasuryFee"`
	DelegationRate   uint64         `json:"delegationRate"`
	Validators       []*Validator   `json:"validators"`
}

type ChainConfig struct {
	ChainID  *big.Int                 `json:"chainId"` // chainId identifies the current chain and is used for replay protection
	Autonity *AutonityContractGenesis `json:"autonity"`
}

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
	Balance *big.Int `json:"balance" gencodec:"required"`
}

type GenesisAlloc map[common.Address]GenesisAccount

type Genesis struct {
	Config     *ChainConfig   `json:"config"`
	Nonce      uint64         `json:"nonce"`
	Timestamp  uint64         `json:"timestamp"`
	GasLimit   uint64         `json:"gasLimit"   gencodec:"required"`
	Difficulty *big.Int       `json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash    `json:"mixHash"`
	Coinbase   common.Address `json:"coinbase"`
	Alloc      GenesisAlloc   `json:"alloc"      gencodec:"required"`

	Number     uint64      `json:"number"`
	GasUsed    uint64      `json:"gasUsed"`
	ParentHash common.Hash `json:"parentHash"`
	BaseFee    *big.Int    `json:"baseFee"`
}

type Validator struct {
	Treasury    common.Address `json:"treasury"`
	Enode       string         `json:"enode"`
	Voter       common.Address `json:"voter"`
	BondedStake *big.Int       `json:"bondedStake"`
}

func makeGenesisConfig(srcTemplate string, dstFile string, vals []*Validator) error {
	file, err := os.Open(srcTemplate)
	if err != nil {
		return err
	}

	defer file.Close()

	genesis := new(Genesis)
	if err = json.NewDecoder(file).Decode(genesis); err != nil {
		return err
	}

	genesis.Config.Autonity.Validators = append(genesis.Config.Autonity.Validators, vals...)

	jsonData, err := json.MarshalIndent(genesis, "", " ")
	if err != nil {
		return err
	}

	if err = os.WriteFile(dstFile, jsonData, 0644); err != nil {
		return err
	}

	return nil
}

// load all keys from keystore with the corresponding password.
func loadKeys(kStore string, password string) ([]*keystore.Key, error) {
	files, err := listDir(kStore)
	if err != nil {
		return nil, err
	}

	var keys []*keystore.Key
	for _, f := range files {
		keyJson, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", kStore, f))
		if err != nil {
			return nil, err
		}

		key, err := keystore.DecryptKey(keyJson, password)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// generate enode url
func genEnode(key *ecdsa.PublicKey, host string, port int) string {
	pub := fmt.Sprintf("%x", crypto.FromECDSAPub(key)[1:])
	return fmt.Sprintf("enode://%s@%s:%d", pub, host, port)
}

// get free ports from current system
func getFreePost(numOfPort int) ([]int, error) {
	return freeport.GetFreePorts(numOfPort)
}

func testReplacePlugin(t *testing.T, port int, pluginDir string) {
	// get the plugins before replacement.
	reqMsg := &types.JSONRPCMessage{Method: "list_plugins"}
	respMsg, err := httpPost(t, reqMsg, port)
	require.NoError(t, err)
	var pluginsAtStart types.PluginByName
	err = json.Unmarshal(respMsg.Result, &pluginsAtStart)
	require.NoError(t, err)

	// do the replacement of plugins.
	err = replacePlugins(pluginDir)
	require.NoError(t, err)
	// wait for replaced plugins to be loaded.
	time.Sleep(10 * time.Second)

	respMsg, err = httpPost(t, reqMsg, port)
	require.NoError(t, err)
	var pluginsAfterReplace types.PluginByName
	err = json.Unmarshal(respMsg.Result, &pluginsAfterReplace)
	require.NoError(t, err)

	for k, p := range pluginsAfterReplace {
		require.Equal(t, p.Name, pluginsAtStart[k].Name)
		require.Equal(t, true, p.StartAt.After(pluginsAtStart[k].StartAt))
	}
}

func testAddPlugin(t *testing.T, port int, pluginDir string) {
	clonerPrefix := "clone"
	clonedPlugins, err := clonePlugins(pluginDir, clonerPrefix, pluginDir)
	defer func() {
		for _, f := range clonedPlugins {
			os.Remove(f) // nolint
		}
	}()

	require.NoError(t, err)
	require.Equal(t, true, len(clonedPlugins) > 0)
	// wait for cloned plugins to be loaded.
	time.Sleep(10 * time.Second)
	testListPlugins(t, port, pluginDir)
}

func testGetVersion(t *testing.T, port int) {
	var reqMsg = &types.JSONRPCMessage{
		Method: "get_version",
	}

	respMsg, err := httpPost(t, reqMsg, port)
	require.NoError(t, err)
	type Version struct {
		Version string
	}
	var V Version
	err = json.Unmarshal(respMsg.Result, &V)
	require.NoError(t, err)
	require.Equal(t, oracleserver.Version, V.Version)
}

func testGetPrices(t *testing.T, port int) {
	reqMsg := &types.JSONRPCMessage{
		Method: "get_prices",
	}

	respMsg, err := httpPost(t, reqMsg, port)
	require.NoError(t, err)
	type PriceAndSymbol struct {
		Prices  types.PriceBySymbol
		Symbols []string
	}
	var ps PriceAndSymbol
	err = json.Unmarshal(respMsg.Result, &ps)
	require.NoError(t, err)
	require.Equal(t, strings.Split(config.DefaultSymbols, ","), ps.Symbols)
	for _, s := range ps.Symbols {
		require.Equal(t, s, ps.Prices[s].Symbol)
		require.Equal(t, true, ps.Prices[s].Price.Equal(helpers.ResolveSimulatedPrice(s)))
	}
}

func testListPlugins(t *testing.T, port int, pluginDir string) {
	reqMsg := &types.JSONRPCMessage{Method: "list_plugins"}

	respMsg, err := httpPost(t, reqMsg, port)
	require.NoError(t, err)
	var plugins types.PluginByName
	err = json.Unmarshal(respMsg.Result, &plugins)
	require.NoError(t, err)
	files, err := listDir(pluginDir)
	require.NoError(t, err)
	require.Equal(t, len(files), len(plugins))
}

func httpPost(t *testing.T, reqMsg *types.JSONRPCMessage, port int) (*types.JSONRPCMessage, error) {
	jsonData, err := json.Marshal(reqMsg)
	require.NoError(t, err)

	resp, err := http.Post(fmt.Sprintf("http://127.0.0.1:%d", port), "application/json", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	var respMsg types.JSONRPCMessage
	err = json.NewDecoder(resp.Body).Decode(&respMsg)
	require.NoError(t, err)
	return &respMsg, nil
}

func replacePlugins(pluginDir string) error {
	rawPlugins, err := listDir(pluginDir)
	if err != nil {
		return err
	}

	clonePrefix := "clone"
	clonedPlugins, err := clonePlugins(pluginDir, clonePrefix, fmt.Sprintf("%s/..", pluginDir))
	if err != nil {
		return err
	}

	for _, file := range clonedPlugins {
		for _, info := range rawPlugins {
			if strings.Contains(file, info.Name()) {
				err := os.Rename(file, fmt.Sprintf("%s/%s", pluginDir, info.Name()))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func clonePlugins(pluginDIR string, clonePrefix string, destDir string) ([]string, error) {

	var clonedPlugins []string
	files, err := listDir(pluginDIR)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// read srcFile
		srcContent, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", pluginDIR, file.Name()))
		if err != nil {
			return clonedPlugins, err
		}

		// create dstFile and copy the content
		newPlugin := fmt.Sprintf("%s/%s%s", destDir, clonePrefix, file.Name())
		err = ioutil.WriteFile(newPlugin, srcContent, file.Mode())
		if err != nil {
			return clonedPlugins, err
		}
		clonedPlugins = append(clonedPlugins, newPlugin)
	}
	return clonedPlugins, nil
}

func listDir(pluginDIR string) ([]fs.FileInfo, error) {
	var plugins []fs.FileInfo

	files, err := ioutil.ReadDir(pluginDIR)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		plugins = append(plugins, file)
	}
	return plugins, nil
}
