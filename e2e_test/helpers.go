package test

import (
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
	"strings"
	"time"
)

var (
	nodeKeyDir         = "./autonity_l1_config/nodekeys"
	keyStoreDir        = "./autonity_l1_config/keystore"
	defaultHost        = "127.0.0.1"
	defaultPlugDir     = "./plugins/fake_plugins"
	binancePlugDir     = "./plugins/production_plugins"
	simulatorPlugDir   = "./plugins/simulator_plugins"
	mixPluginDir       = "./plugins/mix_plugins"
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

// OracleContractGenesis contract config. It'is used for deployment.
type OracleContractGenesis struct {
	// Bytecode of validators contract
	// would like this type to be []byte but the unmarshalling is not working
	Bytecode string `json:"bytecode,omitempty" toml:",omitempty"`
	// Json ABI of the contract
	ABI        string   `json:"abi,omitempty" toml:",omitempty"`
	Symbols    []string `json:"symbols"`
	VotePeriod uint64   `json:"defaultVotePeriod"`
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

type DataSimulator struct {
	Command    *exec.Cmd
	SimulateTM int
}

func (s *DataSimulator) Start() {
	err := s.Command.Run()
	if err != nil {
		log.Error("start data simulator failed", "error", err.Error())
	}
}

func (s *DataSimulator) Stop() {
	err := s.Command.Process.Kill()
	if err != nil {
		log.Error("stop data simulator failed", "error", err.Error())
	}
}

func (s *DataSimulator) GenCMD() {
	c := exec.Command("./simulator", fmt.Sprintf("-sim_timeout=%d", s.SimulateTM))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	s.Command = c
}

type Oracle struct {
	Key       *Key
	PluginDir string
	Host      string
	Command   *exec.Cmd
	Symbols   string
}

// Start starts the process and wait until it exists, the caller should use a go routine to invoke it.
func (o *Oracle) Start() {
	err := o.Command.Run()
	if err != nil {
		panic(err)
	}
}

func (o *Oracle) Stop() {
	err := o.Command.Process.Kill()
	if err != nil {
		log.Error("stop oracle client failed", "error", err.Error())
	}
}

func (o *Oracle) GenCMD(wsEndpoint string) {
	c := exec.Command("./autoracle",
		fmt.Sprintf("-oracle_autonity_ws_url=%s", wsEndpoint),
		fmt.Sprintf("-oracle_crypto_symbols=%s", o.Symbols),
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

type L1Node struct {
	EnableLog bool
	DataDir   string
	NodeKey   *Key
	Host      string
	P2PPort   int
	WSPort    int
	Command   *exec.Cmd
	Validator *Validator
}

func (n *L1Node) GenCMD(genesisFile string) {
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
}

// Start starts the process and wait until it exists, the caller should use a go routine to invoke it.
func (n *L1Node) Start() {
	err := n.Command.Run()
	if err != nil {
		panic(err)
	}
}

func (n *L1Node) Stop() {
	if err := n.Command.Process.Kill(); err != nil {
		log.Warn("stop autonity client failed", "error", err.Error())
	}
	if err := os.RemoveAll(n.DataDir); err != nil {
		log.Warn("cleanup autonity client data DIR failed", "error", err.Error())
	}
}

type NetworkConfig struct {
	EnableL1Logs    bool
	Symbols         string
	VotePeriod      uint64
	PluginDIRs      []string // different oracle can have different plugins configured.
	SimulateTimeout int      // to simulate timeout in seconds at data source simulator when processing http request.
}

type Network struct {
	EnableL1Logs bool
	GenesisFile  string
	OperatorKey  *Key
	TreasuryKey  *Key
	L1Nodes      []*L1Node
	L2Nodes      []*Oracle
	Simulator    *DataSimulator
	Symbols      string
	VotePeriod   uint64
	PluginDirs   []string // different oracle can have different plugins configured.
}

func (net *Network) genGenesisFile() error {
	srcGenesis := defaultGenesis
	dstGenesis := generatedGenesis

	var vals []*Validator
	for _, n := range net.L1Nodes {
		vals = append(vals, n.Validator)
	}

	err := makeGenesisConfig(srcGenesis, dstGenesis, vals, net)
	if err != nil {
		return err
	}
	net.GenesisFile = dstGenesis
	return nil
}

func (net *Network) Start() {
	if net.Simulator != nil {
		net.Simulator.GenCMD()
		go net.Simulator.Start()
	}

	for _, n := range net.L1Nodes {
		n.GenCMD(net.GenesisFile)
		go n.Start()
	}

	time.Sleep(5 * time.Second)

	for i, n := range net.L2Nodes {
		n.GenCMD(fmt.Sprintf("ws://%s:%d", net.L1Nodes[i].Host, net.L1Nodes[i].WSPort))
		go n.Start()
	}
}

func (net *Network) Stop() {
	for _, n := range net.L1Nodes {
		n.Stop()
	}

	for _, n := range net.L2Nodes {
		n.Stop()
	}

	if net.Simulator != nil {
		net.Simulator.Stop()
	}
}

// create with a four-nodes autonity l1 network for the integration of oracle service, with each of validator bind with
// an oracle node.
func createNetwork(netConf *NetworkConfig) (*Network, error) {
	keys, err := loadKeys(keyStoreDir, defaultPassword)
	if err != nil {
		return nil, err
	}

	if len(keys) != numberOfKeys {
		panic("keystore does not contains enough key for testbed")
	}

	var pluginDIRs = []string{defaultPlugDir, defaultPlugDir, defaultPlugDir, defaultPlugDir}
	var simulator *DataSimulator
	for i, d := range netConf.PluginDIRs {
		if i >= numberOfValidators {
			break
		}
		if len(d) != 0 {
			pluginDIRs[i] = d
			if (d == simulatorPlugDir || d == mixPluginDir) && simulator == nil {
				simulator = &DataSimulator{SimulateTM: netConf.SimulateTimeout}
			}
		}
	}

	var network = &Network{
		EnableL1Logs: netConf.EnableL1Logs,
		OperatorKey:  keys[0],
		TreasuryKey:  keys[1],
		Symbols:      netConf.Symbols,
		VotePeriod:   netConf.VotePeriod,
		PluginDirs:   pluginDIRs,
		Simulator:    simulator,
	}

	freePorts, err := getFreePost(numberOfValidators * numberOfPortsForBindNodes)
	if err != nil {
		return nil, err
	}

	network, err = configNetwork(network, keys[2:], freePorts, numberOfValidators)
	if err != nil {
		return nil, err
	}

	network.Start()

	return network, nil
}

func configNetwork(network *Network, freeKeys []*Key, freePorts []int, nodes int) (*Network, error) {

	for i := 0; i < nodes; i++ {
		// allocate a key and a port for l2Node client,
		var l2Node = &Oracle{
			Key:       freeKeys[i*2],
			PluginDir: network.PluginDirs[i],
			Host:      defaultHost,
			Symbols:   network.Symbols,
		}

		// allocate a key and 2 ports for l1 validator client,
		var l1Node = &L1Node{
			EnableLog: network.EnableL1Logs,
			DataDir:   fmt.Sprintf("%s/node_%d/data", defaultDataDirRoot, i),
			NodeKey:   freeKeys[i*2+1],
			Host:      defaultHost,
			P2PPort:   freePorts[i*2],
			WSPort:    freePorts[i*2+1],
		}

		var validator = &Validator{
			Treasury:      l1Node.NodeKey.Key.Address,
			Enode:         genEnode(&l1Node.NodeKey.Key.PrivateKey.PublicKey, l1Node.Host, l1Node.P2PPort),
			OracleAddress: crypto.PubkeyToAddress(l2Node.Key.Key.PrivateKey.PublicKey),
			BondedStake:   defaultBondedStake,
		}

		l1Node.Validator = validator

		// clean up legacy data in the data DIR for the test framework.
		if err := os.RemoveAll(l1Node.DataDir); err != nil {
			log.Warn("Cannot cleanup legacy data for the test framework", "err", err.Error())
		}

		network.L1Nodes = append(network.L1Nodes, l1Node)
		network.L2Nodes = append(network.L2Nodes, l2Node)
	}

	if err := network.genGenesisFile(); err != nil {
		return nil, err
	}

	return network, nil
}

func makeGenesisConfig(srcTemplate string, dstFile string, vals []*Validator, net *Network) error {
	file, err := os.Open(srcTemplate)
	if err != nil {
		return err
	}

	defer file.Close()

	genesis := new(Genesis)
	if err = json.NewDecoder(file).Decode(genesis); err != nil {
		return err
	}
	genesis.Config.Autonity.Operator = net.OperatorKey.Key.Address
	genesis.Config.Autonity.Treasury = net.TreasuryKey.Key.Address
	genesis.Config.Autonity.Validators = append(genesis.Config.Autonity.Validators, vals...)

	genesis.Config.OracleContractConfig.Symbols = strings.Split(net.Symbols, ",")
	genesis.Config.OracleContractConfig.VotePeriod = net.VotePeriod

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
