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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
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

func TestLostL1Connectivity(t *testing.T) {
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

	resetRound := uint64(5)
	endRound := uint64(10)
	nodeIndex := 2
	testRestartL1Node(t, network, nodeIndex, o, resetRound, endRound)
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
		Symbols:      []string{"BTC-USD", "BTC-USDC", "BTC-USDT", "BTC-USD4"},
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

func TestFeeRefund(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      config.DefaultSymbols,
		VotePeriod:   20,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	oc, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	// bind client with autonity contract address
	_, err = autonity.NewAutonity(types.AutonityContractAddress, client)
	require.NoError(t, err)
	// first test happy case.
	tc := time.NewTicker(10 * time.Second)
	lastBalance, err := client.BalanceAt(context.Background(), network.L2Nodes[0].Key.Key.Address, nil)
	require.NoError(t, err)
	for {
		select {
		// check round update after every 10 seconds
		case <-tc.C:
			currRound, err := oc.GetRound(nil)
			curBalance, err := client.BalanceAt(context.Background(), network.L2Nodes[0].Key.Key.Address, nil)
			require.NoError(t, err)
			require.Equal(t, lastBalance, curBalance)
			// check balance update for 5 rounds
			if currRound.Cmp(big.NewInt(5)) == 0 {
				return
			}
			lastBalance = curBalance
		}
	}
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

func TestForexPluginsHappyCase(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      []string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"},
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{forexPlugDir, forexPlugDir, forexPlugDir, forexPlugDir},
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

func TestCAXPluginsHappyCase(t *testing.T) {
	// run the test after the data source cax.devnet.clearmatics.network provides available data.
	//t.Skip("this test depends on the remote service endpoint of cax.devnet.clearmatics.network")
	var conf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      []string{"NTN-USD", "ATN-USD", "NTN-ATN"},
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{caxPlugDir, caxPlugDir, caxPlugDir, caxPlugDir},
	}

	net, err := createNetwork(conf)
	require.NoError(t, err)
	defer net.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", net.L1Nodes[0].Host, net.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	// subscribe on-chain round rotation event
	chRoundEvent := make(chan *contract.OracleNewRound)
	subRoundEvent, err := o.WatchNewRound(new(bind.WatchOpts), chRoundEvent)
	require.NoError(t, err)
	defer subRoundEvent.Unsubscribe()

	symbols := []string{"NTN-USD", "ATN-USD", "NTN-ATN"}
	endRound := uint64(5)

	var prices []contract.IOracleRoundData

	for {
		select {
		case rEvent := <-chRoundEvent:
			if rEvent.Round.Uint64() <= 3 {
				continue
			}

			if rEvent.Round.Uint64() > endRound {
				return
			}

			lastRound := new(big.Int).SetUint64(rEvent.Round.Uint64() - 1)

			for _, s := range symbols {
				p, err := o.GetRoundData(nil, lastRound, s)
				require.NoError(t, err)
				prices = append(prices, p)
			}

			hasBadValue := false
			for _, p := range prices {
				if p.Status.Uint64() != 0 || p.Price.Uint64() == 0 {
					hasBadValue = true
				}
			}
			if !hasBadValue {
				break
			}
		}
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

func testRestartL1Node(t *testing.T, net *Network, index int, o *contract.Oracle, resetRound, beforeRound uint64) {
	for {
		time.Sleep(1 * time.Minute)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		if round.Uint64() == resetRound {
			net.ResetL1Node(index)
		}

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		// verify result.
		_, err = os.FindProcess(net.L2Nodes[index].Command.Process.Pid)
		require.NoError(t, err)
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

		// get last round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, round.Sub(round, common.Big1), s)
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

	newSymbols := append(legacySymbols, "BTC-ETH")

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
