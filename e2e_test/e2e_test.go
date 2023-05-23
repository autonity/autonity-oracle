package test

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	autonity "autonity-oracle/e2e_test/contracts"
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

var defaultVotePeriod = uint64(60)

func TestHappyCase(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      config.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	p, err := o.GetPrecision(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromInt(p.Int64())

	// first test happy case.
	endRound := uint64(10)
	testHappyCase(t, o, endRound, pricePrecision)
}

func TestAddNewSymbol(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      config.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	p, err := o.GetPrecision(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromInt(p.Int64())

	// test to add new symbols.
	endRound := uint64(5)
	testAddNewSymbols(t, network, client, o, endRound, pricePrecision)
}

func TestRMSymbol(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      config.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	p, err := o.GetPrecision(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromInt(p.Int64())

	// test to remove symbols.
	endRound := uint64(5)
	testRMSymbols(t, network, client, o, endRound, pricePrecision)
}

func TestRMCommitteeMember(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      config.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	// bind client with autonity contract address
	aut, err := autonity.NewAutonity(types.AutonityContractAddress, client)
	require.NoError(t, err)

	p, err := o.GetPrecision(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromInt(p.Int64())

	// test to remove validator from current committee.
	endRound := uint64(10)
	testRMValidatorFromCommittee(t, network, client, o, aut, endRound, pricePrecision)
}

func TestAddCommitteeMember(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      config.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	// bind client with autonity contract address
	aut, err := autonity.NewAutonity(types.AutonityContractAddress, client)
	require.NoError(t, err)

	p, err := o.GetPrecision(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromInt(p.Int64())

	// test to add validator into current committee.
	endRound := uint64(10)
	testNewValidatorJoinToCommittee(t, network, client, o, aut, endRound, pricePrecision)
}

func TestHappyCaseWithBinanceDataService(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      "BTCUSD,BTCUSDC,BTCUSDT,BTCUSD4",
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{binancePlugDir, binancePlugDir, binancePlugDir, binancePlugDir},
	}
	network, err := createNetwork(netConf)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	// first test happy case.
	endRound := uint64(10)
	testBinanceDataHappyCase(t, o, endRound)
}

func TestWithBinanceSimulatorOff(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      config.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{simulatorPlugDir, mixPluginDir, mixPluginDir, mixPluginDir},
	}
	network, err := createNetwork(netConf)
	require.NoError(t, err)
	defer network.Stop()

	for i := 0; i < 300; i++ {
		time.Sleep(time.Second)
		network.Simulator.Stop()
	}

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	symbols, err := o.GetSymbols(nil)
	require.NoError(t, err)

	pre, err := o.GetPrecision(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromInt(pre.Int64())

	for _, s := range symbols {
		d, err := o.LatestRoundData(nil, s)
		require.NoError(t, err)
		require.NotEqual(t, uint64(0), d.Price.Uint64())
		require.Equal(t, uint64(0), d.Status.Uint64())

		price, err := decimal.NewFromString(d.Price.String())
		require.NoError(t, err)
		require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
	}
}

func TestWithBinanceSimulatorTimeout(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs:    false,
		Symbols:         config.DefaultSymbols,
		VotePeriod:      defaultVotePeriod,
		PluginDIRs:      []string{simulatorPlugDir, mixPluginDir, mixPluginDir, mixPluginDir},
		SimulateTimeout: 10,
	}
	network, err := createNetwork(netConf)
	require.NoError(t, err)
	defer network.Stop()

	for i := 0; i < 300; i++ {
		time.Sleep(time.Second)
	}

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	symbols, err := o.GetSymbols(nil)
	require.NoError(t, err)

	for _, s := range symbols {
		d, err := o.LatestRoundData(nil, s)
		require.NoError(t, err)
		require.NotEqual(t, uint64(0), d.Price.Uint64())
		require.Equal(t, uint64(0), d.Status.Uint64())
	}
}

func testHappyCase(t *testing.T, o *contract.Oracle, beforeRound uint64, pricePrecision decimal.Decimal) {
	for {
		time.Sleep(1 * time.Minute)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		var rd []contract.IOracleRoundData
		symbols, err := o.GetSymbols(nil)
		require.NoError(t, err)

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, round, s)
			require.NoError(t, err)
			require.Equal(t, uint64(0), d.Status.Uint64())
			rd = append(rd, d)
		}

		// verify result.
		for i, s := range symbols {
			price, err := decimal.NewFromString(rd[i].Price.String())
			require.NoError(t, err)
			require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
		}
		break
	}
}

func testBinanceDataHappyCase(t *testing.T, o *contract.Oracle, beforeRound uint64) {
	for {
		time.Sleep(1 * time.Minute)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		symbols, err := o.GetSymbols(nil)
		require.NoError(t, err)

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, round, s)
			require.NoError(t, err)
			require.NotEqual(t, uint64(0), d.Price.Uint64())
			require.Equal(t, uint64(0), d.Status.Uint64())
		}

		break
	}
}

func testAddNewSymbols(t *testing.T, network *Network, client *ethclient.Client, o *contract.Oracle, beforeRound uint64,
	pricePrecision decimal.Decimal) {
	chainID, err := client.ChainID(context.Background())
	require.NoError(t, err)

	auth, err := bind.NewKeyedTransactorWithChainID(network.OperatorKey.Key.PrivateKey, chainID)
	require.NoError(t, err)

	auth.Value = big.NewInt(0)
	auth.GasTipCap = big.NewInt(0)
	auth.GasLimit = uint64(3000000)

	legacySymbols, err := o.GetSymbols(nil)
	require.NoError(t, err)

	newSymbols := append(legacySymbols, "BTCETH")

	_, err = o.SetSymbols(auth, newSymbols)
	require.NoError(t, err)

	for {
		time.Sleep(1 * time.Minute)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		var rd []contract.IOracleRoundData
		symbols, err := o.GetSymbols(nil)
		require.NoError(t, err)

		require.Equal(t, len(newSymbols), len(symbols))

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, round, s)
			require.NoError(t, err)
			rd = append(rd, d)
		}

		// verify result.
		for i, s := range symbols {
			price, err := decimal.NewFromString(rd[i].Price.String())
			require.NoError(t, err)
			require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
			require.Equal(t, uint64(0), rd[i].Status.Uint64())
		}
		break
	}
}

func testRMSymbols(t *testing.T, network *Network, client *ethclient.Client, o *contract.Oracle, beforeRound uint64,
	pricePrecision decimal.Decimal) {
	from := network.OperatorKey.Key.Address

	nonce, err := client.PendingNonceAt(context.Background(), from)
	require.NoError(t, err)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	chainID, err := client.ChainID(context.Background())
	require.NoError(t, err)

	auth, err := bind.NewKeyedTransactorWithChainID(network.OperatorKey.Key.PrivateKey, chainID)
	require.NoError(t, err)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	legacySymbols, err := o.GetSymbols(nil)
	require.NoError(t, err)

	newSymbols := append(legacySymbols[0:1], legacySymbols[3:]...)

	_, err = o.SetSymbols(auth, newSymbols)
	require.NoError(t, err)

	for {
		time.Sleep(1 * time.Minute)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		var rd []contract.IOracleRoundData
		symbols, err := o.GetSymbols(nil)
		require.NoError(t, err)

		require.Equal(t, len(newSymbols), len(symbols))

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, round, s)
			require.NoError(t, err)
			rd = append(rd, d)
		}

		// verify result.
		for i, s := range symbols {
			price, err := decimal.NewFromString(rd[i].Price.String())
			require.NoError(t, err)
			require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
			require.Equal(t, uint64(0), rd[i].Status.Uint64())
		}
		break
	}
}

func testRMValidatorFromCommittee(t *testing.T, network *Network, client *ethclient.Client, o *contract.Oracle,
	aut *autonity.Autonity, beforeRound uint64, pricePrecision decimal.Decimal) {
	from := network.OperatorKey.Key.Address

	nonce, err := client.PendingNonceAt(context.Background(), from)
	require.NoError(t, err)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	chainID, err := client.ChainID(context.Background())
	require.NoError(t, err)

	auth, err := bind.NewKeyedTransactorWithChainID(network.OperatorKey.Key.PrivateKey, chainID)
	require.NoError(t, err)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	newCommitteeSize := int64(len(network.L1Nodes) / 2)
	_, err = aut.SetCommitteeSize(auth, big.NewInt(newCommitteeSize))
	require.NoError(t, err)

	for {
		time.Sleep(1 * time.Minute)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		var rd []contract.IOracleRoundData
		symbols, err := o.GetSymbols(nil)
		require.NoError(t, err)

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, round, s)
			require.NoError(t, err)
			rd = append(rd, d)
		}

		// verify result.
		for i, s := range symbols {
			price, err := decimal.NewFromString(rd[i].Price.String())
			require.NoError(t, err)
			require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
			require.Equal(t, uint64(0), rd[i].Status.Uint64())
		}

		// todo: check the leaving validator shouldn't be slashed if it does not omit any report.
		break
	}
}

func testNewValidatorJoinToCommittee(t *testing.T, network *Network, client *ethclient.Client, o *contract.Oracle,
	aut *autonity.Autonity, beforeRound uint64, pricePrecision decimal.Decimal) {
	from := network.OperatorKey.Key.Address

	nonce, err := client.PendingNonceAt(context.Background(), from)
	require.NoError(t, err)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	require.NoError(t, err)

	chainID, err := client.ChainID(context.Background())
	require.NoError(t, err)

	auth, err := bind.NewKeyedTransactorWithChainID(network.OperatorKey.Key.PrivateKey, chainID)
	require.NoError(t, err)

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	newCommitteeSize := int64(len(network.L1Nodes))
	_, err = aut.SetCommitteeSize(auth, big.NewInt(newCommitteeSize))
	require.NoError(t, err)

	for {
		time.Sleep(1 * time.Minute)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		var rd []contract.IOracleRoundData
		symbols, err := o.GetSymbols(nil)
		require.NoError(t, err)

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, round, s)
			require.NoError(t, err)
			rd = append(rd, d)
		}

		// verify result.
		for i, s := range symbols {
			price, err := decimal.NewFromString(rd[i].Price.String())
			require.NoError(t, err)
			require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
			require.Equal(t, uint64(0), rd[i].Status.Uint64())
		}

		// todo: newly added validator shouldn't be slashed if they not omit any report.
		break
	}
}

// todo: missing report by omission voter, the faulty node should be slashed.
func TestOmissionVoter(t *testing.T) {

}