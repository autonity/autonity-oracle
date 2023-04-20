package test

import (
	"autonity-oracle/config"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/phayes/freeport"
	"io/fs"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"time"
)

var (
	nodeKeyDir         = "./autonity_l1_config/nodekeys"
	keyStoreDir        = "./autonity_l1_config/keystore"
	defaultHost        = "127.0.0.1"
	defaultPlugDir     = "./plugin_dir"
	defaultGenesis     = "./autonity_l1_config/genesis_template.json"
	defaultPassword    = "test"
	generatedGenesis   = "./autonity_l1_config/genesis_gen.json"
	defaultDataDirRoot = "./autonity_l1_config/nodes"

	defaultBondedStake = new(big.Int).SetUint64(1000)

	numberOfKeys              = 10
	numberOfValidators        = 4
	numberOfPortsForBindNodes = 2
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

// Autonity contract config. It'is used for deployment.
type OracleContractGenesis struct {
	// Bytecode of validators contract
	// would like this type to be []byte but the unmarshalling is not working
	Bytecode string `json:"bytecode,omitempty" toml:",omitempty"`
	// Json ABI of the contract
	ABI        string         `json:"abi,omitempty" toml:",omitempty"`
	Operator   common.Address `json:"operator"`
	Symbols    []string       `json:"symbols"`
	VotePeriod uint64         `json:"votePeriod"`
}

type ChainConfig struct {
	ChainID              *big.Int                 `json:"chainId"` // chainId identifies the current chain and is used for replay protection
	Autonity             *AutonityContractGenesis `json:"autonity"`
	OracleContractConfig *OracleContractGenesis   `json:"oracle,omitempty"`
}

type Validator struct {
	Treasury      common.Address `json:"treasury"`
	Enode         string         `json:"enode"`
	OracleAddress common.Address `json:"oracleAddress"`
	BondedStake   *big.Int       `json:"bondedStake"`
}

type Oracle struct {
	Key       *Key
	PluginDir string
	Host      string
	ProcessID int
	Command   *exec.Cmd
}

// Start starts the process and wait until it exists, the caller should use a go routine to invoke it.
func (o *Oracle) Start() {
	err := o.Command.Run()
	if err != nil {
		//panic(err)
	}
}

func (o *Oracle) Stop() {
	err := o.Command.Process.Kill()
	if err != nil {
		log.Info("stop oracle client failed", "error", err.Error())
	}
}

func (o *Oracle) GenCMD(wsEndpoint string) {
	c := exec.Command("./autoracle",
		fmt.Sprintf("-oracle_autonity_ws_url=%s", wsEndpoint),
		fmt.Sprintf("-oracle_crypto_symbols=%s", config.DefaultSymbols),
		fmt.Sprintf("-oracle_key_file=%s", o.Key.KeyFile),
		fmt.Sprintf("-oracle_key_password=%s", o.Key.Password),
		fmt.Sprintf("-oracle_plugin_dir=%s", o.PluginDir),
	)

	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	o.Command = c
}

type Key struct {
	KeyFile    string
	RawKeyFile string
	Password   string
	Key        *keystore.Key
}

type Node struct {
	EnableLog    bool
	DataDir      string
	NodeKey      *Key
	Host         string
	P2PPort      int
	WSPort       int
	ProcessID    int
	OracleClient *Oracle
	Command      *exec.Cmd
	Validator    *Validator
}

func (n *Node) GenCMD(genesisFile string) {
	c := exec.Command("./autonity",
		"--ipcdisable", "--datadir", n.DataDir, "--genesis", genesisFile, "--nodekey", n.NodeKey.RawKeyFile, "--ws",
		"--ws.addr", n.Host, "--ws.port", fmt.Sprintf("%d", n.WSPort), "--ws.api",
		"tendermint,eth,web3,admin,debug,miner,personal,txpool,net", "--syncmode", "full", "--miner.gaslimit",
		"100000000", "--miner.threads", fmt.Sprintf("%d", 1), "--port", fmt.Sprintf("%d", n.P2PPort))

	// enable logging in the standard outputs.
	if n.EnableLog {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}

	n.Command = c
	n.OracleClient.GenCMD(fmt.Sprintf("ws://%s:%d", n.Host, n.WSPort))
}

// Start starts the process and wait until it exists, the caller should use a go routine to invoke it.
func (n *Node) Start() {
	err := n.Command.Run()
	if err != nil {
		panic(err)
	}
}

func (n *Node) Stop() {
	if err := n.Command.Process.Kill(); err != nil {
		log.Warn("stop autonity client failed", "error", err.Error())
	}
	if err := os.RemoveAll(n.DataDir); err != nil {
		log.Warn("cleanup autonity client data DIR failed", "error", err.Error())
	}
}

type Network struct {
	EnableL1Logs bool
	GenesisFile  string
	OperatorKey  *Key
	TreasuryKey  *Key
	Nodes        []*Node
}

func (net *Network) genGenesisFile() error {
	srcGenesis := defaultGenesis
	dstGenesis := generatedGenesis

	var vals []*Validator
	for _, n := range net.Nodes {
		vals = append(vals, n.Validator)
	}

	err := makeGenesisConfig(srcGenesis, dstGenesis, vals, net.TreasuryKey.Key.Address, net.OperatorKey.Key.Address)
	if err != nil {
		return err
	}
	net.GenesisFile = dstGenesis
	return nil
}

// prepare configurations for autonity l1 node and oracle client node
func (net *Network) genConfigs() error {
	if err := net.genGenesisFile(); err != nil {
		return err
	}

	for _, n := range net.Nodes {
		n.GenCMD(net.GenesisFile)
	}
	return nil
}

func (net *Network) Start() {
	for _, n := range net.Nodes {
		go n.Start()
		time.Sleep(5 * time.Second)
		go n.OracleClient.Start()
	}
}

func (net *Network) Stop() {
	for _, n := range net.Nodes {
		n.OracleClient.Stop()
		n.Stop()
	}
}

// create with a four-nodes autonity l1 network for the integration of oracle service, with each of validator bind with
// an oracle node.
func createNetwork(enableL1Logs bool) (*Network, error) {
	keys, err := loadKeys(keyStoreDir, defaultPassword)
	if err != nil {
		return nil, err
	}

	if len(keys) != numberOfKeys {
		panic("keystore does not contains enough key for testbed")
	}

	var network = &Network{
		EnableL1Logs: enableL1Logs,
		OperatorKey:  keys[0],
		TreasuryKey:  keys[1],
	}

	freePorts, err := getFreePost(numberOfValidators * numberOfPortsForBindNodes)
	if err != nil {
		return nil, err
	}

	//plan the network with number of validators, allocate configs for L1 node and the corresponding oracle client.
	network, err = prepareResource(network, keys[2:], freePorts, numberOfValidators)
	if err != nil {
		return nil, err
	}

	err = network.genConfigs()
	if err != nil {
		return nil, err
	}

	network.Start()

	return network, nil
}

func prepareResource(network *Network, freeKeys []*Key, freePorts []int, nodes int) (*Network, error) {

	for i := 0; i < nodes; i++ {
		// allocate a key and a port for oracle client,
		var oracle = &Oracle{
			Key:       freeKeys[i*2],
			PluginDir: defaultPlugDir,
			Host:      defaultHost,
			ProcessID: -1,
		}

		// allocate a key and 2 ports for validator client,
		var aut = &Node{
			EnableLog:    network.EnableL1Logs,
			DataDir:      fmt.Sprintf("%s/node_%d/data", defaultDataDirRoot, i),
			NodeKey:      freeKeys[i*2+1],
			Host:         defaultHost,
			P2PPort:      freePorts[i*2],
			WSPort:       freePorts[i*2+1],
			OracleClient: oracle,
		}

		var validator = &Validator{
			Treasury:      aut.NodeKey.Key.Address,
			Enode:         genEnode(&aut.NodeKey.Key.PrivateKey.PublicKey, aut.Host, aut.P2PPort),
			OracleAddress: crypto.PubkeyToAddress(oracle.Key.Key.PrivateKey.PublicKey),
			BondedStake:   defaultBondedStake,
		}

		aut.OracleClient = oracle
		aut.Validator = validator

		// clean up legacy data in the data DIR for the test framework.
		if err := os.RemoveAll(aut.DataDir); err != nil {
			log.Warn("Cannot cleanup legacy data for the test framework", "err", err.Error())
		}

		network.Nodes = append(network.Nodes, aut)
	}
	return network, nil
}

func makeGenesisConfig(srcTemplate string, dstFile string, vals []*Validator, treasury common.Address, operator common.Address) error {
	file, err := os.Open(srcTemplate)
	if err != nil {
		return err
	}

	defer file.Close()

	genesis := new(Genesis)
	if err = json.NewDecoder(file).Decode(genesis); err != nil {
		return err
	}
	genesis.Config.Autonity.Operator = operator
	genesis.Config.Autonity.Treasury = treasury
	genesis.Config.Autonity.Validators = append(genesis.Config.Autonity.Validators, vals...)
	genesis.Config.OracleContractConfig.VotePeriod = 30
	genesis.Config.OracleContractConfig.Operator = operator

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
func loadKeys(kStore string, password string) ([]*Key, error) {
	files, err := listDir(kStore)
	if err != nil {
		return nil, err
	}

	var keys []*Key
	for _, f := range files {
		keyFile := fmt.Sprintf(fmt.Sprintf("%s/%s", kStore, f.Name()))
		keyJson, err := ioutil.ReadFile(keyFile)
		if err != nil {
			return nil, err
		}

		key, err := keystore.DecryptKey(keyJson, password)
		if err != nil {
			return nil, err
		}

		strKey := hex.EncodeToString(crypto.FromECDSA(key.PrivateKey))
		rawKeyFile := fmt.Sprintf("%s/%s", nodeKeyDir, key.Address)
		if err := os.WriteFile(rawKeyFile, []byte(strKey), 0666); err != nil {
			return nil, err
		}

		var k = &Key{Key: key, KeyFile: keyFile, Password: password, RawKeyFile: rawKeyFile}
		keys = append(keys, k)
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
