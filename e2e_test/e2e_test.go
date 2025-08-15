package test

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	autonity "autonity-oracle/e2e_test/contracts"
	"autonity-oracle/e2e_test/contracts/erc20"
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var defaultVotePeriod = uint64(30)
var defaultTransferAmount = uint64(100000000000000000)

func TestHappyCase(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, 2)

	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	p, err := o.GetDecimals(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromBigInt(common.Big1, int32(p))

	// first test happy case.
	endRound := uint64(10)
	testHappyCase(t, o, endRound, pricePrecision)
}

func TestLostL1Connectivity(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, numberOfValidators)
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
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, 2)

	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	p, err := o.GetDecimals(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromBigInt(common.Big1, int32(p))

	// test to add new symbols.
	endRound := uint64(5)
	testAddNewSymbols(t, network, client, o, endRound, pricePrecision)
}

func TestRMSymbol(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, numberOfValidators)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	p, err := o.GetDecimals(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromBigInt(common.Big1, int32(p))

	// test to remove symbols.
	endRound := uint64(5)
	testRMSymbols(t, network, client, o, endRound, pricePrecision)
}

func TestRMCommitteeMember(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, numberOfValidators)
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

	p, err := o.GetDecimals(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromBigInt(common.Big1, int32(p))

	// test to remove validator from current committee.
	endRound := uint64(10)
	testRMValidatorFromCommittee(t, network, client, o, aut, endRound, pricePrecision)
}

func TestAddCommitteeMember(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, numberOfValidators)
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

	p, err := o.GetDecimals(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromBigInt(common.Big1, int32(p))

	// test to add validator into current committee.
	endRound := uint64(10)
	testNewValidatorJoinToCommittee(t, network, client, o, aut, endRound, pricePrecision)
}

func TestFeeRefund(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   20,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, numberOfValidators)
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
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{simulatorPlugDir, mixPluginDir, mixPluginDir, mixPluginDir},
	}
	network, err := createNetwork(netConf, numberOfValidators)
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

	p, err := o.GetDecimals(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromBigInt(common.Big1, int32(p))

	for _, s := range symbols {
		d, err := o.LatestRoundData(nil, s)
		require.NoError(t, err)
		require.NotEqual(t, uint64(0), d.Price.Uint64())
		require.Equal(t, true, d.Success)

		price, err := decimal.NewFromString(d.Price.String())
		require.NoError(t, err)
		require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
	}
}

func TestWithBinanceSimulatorTimeout(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs:    false,
		Symbols:         helpers.DefaultSymbols,
		VotePeriod:      defaultVotePeriod,
		PluginDIRs:      []string{simulatorPlugDir, mixPluginDir, mixPluginDir, mixPluginDir},
		SimulateTimeout: 10,
	}
	network, err := createNetwork(netConf, numberOfValidators)
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
		require.Equal(t, true, d.Success)
	}
}

func TestForexPluginsHappyCase(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      []string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"},
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{forexPlugDir, forexPlugDir, forexPlugDir, forexPlugDir},
	}
	network, err := createNetwork(netConf, numberOfValidators)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	p, err := o.GetDecimals(nil)
	require.NoError(t, err)
	pricePrecision := decimal.NewFromBigInt(common.Big1, int32(p))

	// first test happy case.
	endRound := uint64(10)
	testHappyCase(t, o, endRound, pricePrecision)
}

func TestCryptoPluginsHappyCase(t *testing.T) {
	var conf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      []string{"NTN-USD", "ATN-USD", "NTN-ATN"},
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{cryptoPlugDir, cryptoPlugDir},
	}

	net, err := createNetwork(conf, 2)
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
	endRound := uint64(10)

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
				if !p.Success || p.Price.Uint64() == 0 {
					hasBadValue = true
				}
			}
			if !hasBadValue {
				break
			}
		}
	}
}

func TestSingleNodeCryptoPluginsHappyCase(t *testing.T) {
	var conf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      []string{"NTN-USD", "ATN-USD", "NTN-ATN"},
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{cryptoPlugDir, cryptoPlugDir},
	}

	net, err := createNetwork(conf, 2)
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
	endRound := uint64(10)

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
				if !p.Success || p.Price.Uint64() == 0 {
					hasBadValue = true
				}
			}
			if !hasBadValue {
				break
			}
		}
	}
}

func TestOmissionFaultyVoter(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   20,      // 20s to shorten this test.
		EpochPeriod:  20 * 12, // 4 minutes to shorten this test.
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir, defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, numberOfValidators)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	aut, err := autonity.NewAutonity(types.AutonityContractAddress, client)
	require.NoError(t, err)

	// bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)

	endRound := uint64(90)
	stopRound := uint64(2)

	doneCh := make(chan struct{})

	// Start watching the round event and stop oracle client on the stop round.
	go L2NodeResetEventLoop(t, 0, network, doneCh, o, endRound, stopRound, aut, client)

	// Start a timeout to wait for the ending for the test.
	timeout := time.After(35 * time.Minute) // Adjust the timeout as needed
	select {
	case <-timeout:
		close(doneCh)
	}

	for _, n := range network.L2Nodes {
		account := n.Key.Key.Address
		atnBalance, err := client.BalanceAt(context.Background(), account, nil)
		require.NoError(t, err)
		t.Log("get oracle ATN reward", "address", account.Hex(), "atn balance", atnBalance.String())
		//require.Equal(t, true, atnBalance.Cmp(big.NewInt(0)) > 0)
	}

	// bind client with autonity contract address
	ntnContract, err := erc20.NewErc20(types.AutonityContractAddress, client)
	require.NoError(t, err)
	for _, n := range network.L2Nodes {
		account := n.Key.Key.Address
		ntnBalance, err := ntnContract.BalanceOf(nil, account)
		require.NoError(t, err)
		t.Log("get oracle NTN reward", "address", account.Hex(), "ntn balance", ntnBalance.String())
		//require.Equal(t, true, ntnBalance.Cmp(big.NewInt(0)) > 0)
	}
}

func TestOutlierVoter(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   20,  // 20s to shorten this test.
		EpochPeriod:  120, // 2 minutes to shorten this test.
		PluginDIRs:   []string{outlierPlugDir, defaultPlugDir, defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, numberOfValidators)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.L1Nodes[0].Host, network.L1Nodes[0].WSPort))
	require.NoError(t, err)
	defer client.Close()

	// Bind client with oracle contract address
	o, err := contract.NewOracle(types.OracleContractAddress, client)
	require.NoError(t, err)
	maliciousNode := network.L2Nodes[0].Key.Key.Address
	doneCh := make(chan struct{})
	resultCh := make(chan uint64) // Channel to receive the result
	endRound := uint64(15)

	// Start watching the event and count the number of events received.
	go func() {
		resultCh <- penalizeEventListener(t, maliciousNode, doneCh, o, endRound)
	}()

	// Start a timeout to wait for the ending for the test, and verify the number of penalized events received.
	timeout := time.After(1 * time.Hour) // Adjust the timeout as needed
	//timeout := time.After(250 * time.Second) // Adjust the timeout as needed
	select {
	case penalizedCounter := <-resultCh:
		t.Log("Number of penalized events received:", penalizedCounter)
		require.LessOrEqual(t, penalizedCounter, uint64(2))
		defer os.Remove("./server_state_dump.json") //nolint
	case <-timeout:
		t.Fatal("Test timed out waiting for penalized events")
	}

	close(doneCh)
}

func TestDisableAndEnablePlugin(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, 2)

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
	testDisableAndEnablePlugin(t, network, o, endRound)
}

func TestAddAndRemovePlugin(t *testing.T) {
	var netConf = &NetworkConfig{
		EnableL1Logs: false,
		Symbols:      helpers.DefaultSymbols,
		VotePeriod:   defaultVotePeriod,
		PluginDIRs:   []string{defaultPlugDir, defaultPlugDir},
	}
	network, err := createNetwork(netConf, 2)

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
	testAddAndRemovePlugin(t, o, endRound)
}

func TestResetOracleServer(t *testing.T) {
	// todo: test oracle server resetting with vote persistence recovery.
}

// todo: not a high priority, refine the tests in this e2e test framework.
func testAddAndRemovePlugin(t *testing.T, o *contract.Oracle, beforeRound uint64) {
	added := false
	removed := false
	for {
		time.Sleep(10 * time.Second)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// add a new plugin at round 5
		if round.Uint64() == 5 && added == false {
			err = copyFile("./plugins/crypto_plugins/crypto_uniswap", "./plugins/template_plugins/crypto_uniswap")
			require.NoError(t, err)
			added = true
		}

		// remove the plugin at round 8
		if round.Uint64() == 8 && removed == false {
			err = os.Remove("./plugins/template_plugins/crypto_uniswap")
			require.NoError(t, err)
			removed = true
		}

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		var rd []contract.IOracleRoundData
		symbols, err := o.GetSymbols(nil)
		require.NoError(t, err)

		// as current round is not finalized yet, thus the round data of it haven't being aggregate,
		// thus we will query the last round's data for the verification.
		lastRound := new(big.Int).SetUint64(round.Uint64() - 1)
		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, lastRound, s)
			require.NoError(t, err)
			require.Equal(t, true, d.Success)
			rd = append(rd, d)
		}

		break
	}
}

func testDisableAndEnablePlugin(t *testing.T, network *Network, o *contract.Oracle, beforeRound uint64) {
	disabled := false
	enabled := false
	for {
		time.Sleep(10 * time.Second)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// disable all the plugins at round 5
		if round.Uint64() == 5 && disabled == false {
			for _, n := range network.L2Nodes {
				conf, err := config.LoadServerConfig(n.OracleConf)
				require.NoError(t, err)

				for i := range conf.PluginConfigs {
					conf.PluginConfigs[i].Disabled = true
				}

				err = FlushServerConfig(conf, n.OracleConf)
				require.NoError(t, err)
			}
			disabled = true
		}

		if round.Uint64() == 8 && enabled == false {
			for _, n := range network.L2Nodes {
				conf, err := config.LoadServerConfig(n.OracleConf)
				require.NoError(t, err)

				for i := range conf.PluginConfigs {
					conf.PluginConfigs[i].Disabled = false
				}

				err = FlushServerConfig(conf, n.OracleConf)
				require.NoError(t, err)
			}
			enabled = true
		}

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		var rd []contract.IOracleRoundData
		symbols, err := o.GetSymbols(nil)
		require.NoError(t, err)

		// as current round is not finalized yet, thus the round data of it haven't being aggregate,
		// thus we will query the last round's data for the verification.
		lastRound := new(big.Int).SetUint64(round.Uint64() - 1)
		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, lastRound, s)
			require.NoError(t, err)
			require.Equal(t, true, d.Success)
			rd = append(rd, d)
		}

		break
	}
}

func testHappyCase(t *testing.T, o *contract.Oracle, beforeRound uint64, pricePrecision decimal.Decimal) {
	for {
		time.Sleep(10 * time.Second)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// continue to wait until end round.
		if round.Uint64() < beforeRound {
			continue
		}

		var rd []contract.IOracleRoundData
		symbols, err := o.GetSymbols(nil)
		require.NoError(t, err)

		// as current round is not finalized yet, thus the round data of it haven't being aggregate,
		// thus we will query the last round's data for the verification.
		lastRound := new(big.Int).SetUint64(round.Uint64() - 1)
		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, lastRound, s)
			require.NoError(t, err)
			require.Equal(t, true, d.Success)
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
		time.Sleep(10 * time.Second)
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

	newSymbols := append(legacySymbols, helpers.SymbolBTCETH)

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

		// as current round is not finalized yet, thus the round data of it haven't being aggregate,
		// thus we will query the last round's data for the verification.
		lastRound := new(big.Int).SetUint64(round.Uint64() - 1)

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, lastRound, s)
			require.NoError(t, err)
			rd = append(rd, d)
		}

		// verify result.
		for i, s := range symbols {
			price, err := decimal.NewFromString(rd[i].Price.String())
			require.NoError(t, err)
			require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
			require.Equal(t, true, rd[i].Success)
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

		// as current round is not finalized yet, thus the round data of it haven't being aggregate,
		// thus we will query the last round's data for the verification.
		lastRound := new(big.Int).SetUint64(round.Uint64() - 1)

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, lastRound, s)
			require.NoError(t, err)
			rd = append(rd, d)
		}

		// verify result.
		for i, s := range symbols {
			price, err := decimal.NewFromString(rd[i].Price.String())
			require.NoError(t, err)
			require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
			require.Equal(t, true, rd[i].Success)
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

		// as current round is not finalized yet, thus the round data of it haven't being aggregate,
		// thus we will query the last round's data for the verification.
		lastRound := new(big.Int).SetUint64(round.Uint64() - 1)

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, lastRound, s)
			require.NoError(t, err)
			rd = append(rd, d)
		}

		// verify result.
		for i, s := range symbols {
			price, err := decimal.NewFromString(rd[i].Price.String())
			require.NoError(t, err)
			require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
			require.Equal(t, true, rd[i].Success)
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

		// as current round is not finalized yet, thus the round data of it haven't being aggregate,
		// thus we will query the last round's data for the verification.
		lastRound := new(big.Int).SetUint64(round.Uint64() - 1)

		// get round data for each symbol.
		for _, s := range symbols {
			d, err := o.GetRoundData(nil, lastRound, s)
			require.NoError(t, err)
			rd = append(rd, d)
		}

		// verify result.
		for i, s := range symbols {
			price, err := decimal.NewFromString(rd[i].Price.String())
			require.NoError(t, err)
			require.True(t, true, price.Div(pricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
			require.Equal(t, true, rd[i].Success)
		}

		// todo: newly added validator shouldn't be slashed if they not omit any report.
		break
	}
}

func penalizeEventListener(t *testing.T, nodeAddress common.Address, done chan struct{}, oracle *contract.Oracle, endRound uint64) uint64 {
	// Subscribe to the penalize event with client address.
	chPenalizedEvent := make(chan *contract.OraclePenalized)
	t.Log("current node account", nodeAddress)
	subPenalizedEvent, err := oracle.WatchPenalized(new(bind.WatchOpts), chPenalizedEvent, []common.Address{nodeAddress})
	require.NoError(t, err)
	defer subPenalizedEvent.Unsubscribe()

	// Subscribe to the round event
	chRoundEvent := make(chan *contract.OracleNewRound)
	subRoundEvent, err := oracle.WatchNewRound(new(bind.WatchOpts), chRoundEvent)
	require.NoError(t, err)
	defer subRoundEvent.Unsubscribe()

	var penalized uint64
	for {
		select {
		case <-done:
			return penalized
		case err := <-subPenalizedEvent.Err():
			if err != nil {
				return penalized
			}
		case err := <-subRoundEvent.Err():
			if err != nil {
				return penalized
			}
		case penalizeEvent := <-chPenalizedEvent:
			t.Log("*****Oracle client get penalized as an outlier", "oracle node", penalizeEvent.Participant.String(),
				"currency symbol", penalizeEvent.Symbol, "median value", penalizeEvent.Median.String(), "reported value", penalizeEvent.Reported.String())
			penalized++
		case roundEvent := <-chRoundEvent:
			t.Log("round event received", "round", roundEvent.Round)
			if roundEvent.Round.Uint64() > endRound {
				return penalized
			}
		}
	}
}

func L2NodeResetEventLoop(t *testing.T, nodeIndex int, network *Network, done chan struct{}, oracle *contract.Oracle, endRound uint64, stopRound uint64, aut *autonity.Autonity, client *ethclient.Client) {
	// Watch height events, let it trigger some ATN transfers to let the TXN fee rewards(in ATN) to be shared.
	chHeadEvent := make(chan *types2.Header)
	subHeadEvent, err := client.SubscribeNewHead(context.Background(), chHeadEvent)
	require.NoError(t, err)
	defer subHeadEvent.Unsubscribe()

	// Subscribe to the epoch event
	chEpochEvent := make(chan *autonity.AutonityNewEpoch)
	subEpochEvent, err := aut.WatchNewEpoch(new(bind.WatchOpts), chEpochEvent)
	require.NoError(t, err)
	defer subEpochEvent.Unsubscribe()

	// Subscribe to the round event
	chRoundEvent := make(chan *contract.OracleNewRound)
	subRoundEvent, err := oracle.WatchNewRound(new(bind.WatchOpts), chRoundEvent)
	require.NoError(t, err)
	defer subRoundEvent.Unsubscribe()

	for {
		select {
		case <-done:
			return
		case headEvent := <-chHeadEvent:
			t.Log("new head event", headEvent.Number.Uint64())
			amount := new(big.Int).SetUint64(defaultTransferAmount)
			for _, v := range network.L1Nodes {
				err = transferATN(client, v.NodeKey.Key.PrivateKey, types.OracleContractAddress, amount, headEvent.BaseFee)
				require.NoError(t, err)
			}
		case epochEvent := <-chEpochEvent:
			t.Log("received epoch event", "epoch id", epochEvent.Epoch.Uint64())
			atnOracleContract, err := client.BalanceAt(context.Background(), types.OracleContractAddress, nil)
			require.NoError(t, err)
			t.Log("oracle contract have atn balance", atnOracleContract.Uint64())
			// bind client with autonity contract address
			ntnContract, err := erc20.NewErc20(types.AutonityContractAddress, client)
			require.NoError(t, err)
			ntnOracleContract, err := ntnContract.BalanceOf(nil, types.OracleContractAddress)
			require.NoError(t, err)
			t.Log("oracle contract have ntn balance", ntnOracleContract.Uint64())

			for _, n := range network.L2Nodes {
				account := n.Key.Key.Address
				atnBalance, err := client.BalanceAt(context.Background(), account, nil)
				require.NoError(t, err)
				t.Log("get oracle ATN reward", "address", account.Hex(), "atn balance", atnBalance.String())
				//require.Equal(t, true, atnBalance.Cmp(big.NewInt(0)) > 0)
			}

			for _, n := range network.L2Nodes {
				account := n.Key.Key.Address
				ntnBalance, err := ntnContract.BalanceOf(nil, account)
				require.NoError(t, err)
				t.Log("get oracle NTN reward", "address", account.Hex(), "ntn balance", ntnBalance.String())
				//require.Equal(t, true, ntnBalance.Cmp(big.NewInt(0)) > 0)
			}
		case roundEvent := <-chRoundEvent:
			t.Log("round event received", "round", roundEvent.Round)
			if roundEvent.Round.Uint64() == stopRound {
				t.Log("stopping oracle client", "node address", network.L2Nodes[nodeIndex].Key.Key.Address)
				network.StopL2Node(nodeIndex)
			}

			if roundEvent.Round.Uint64() > endRound {
				return
			}
		}
	}
}

func transferATN(client *ethclient.Client, privateKey *ecdsa.PrivateKey, receiverAddress common.Address, value *big.Int, baseFee *big.Int) error {

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return err
	}
	gasTip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return err
	}

	gasFeeCap := new(big.Int).Add(gasTip, new(big.Int).Mul(baseFee, big.NewInt(2)))

	signer := types2.NewLondonSigner(chainID)
	signedTX, err := newDynamicTX(nonce, gasTip, gasFeeCap, 21000, receiverAddress, value, signer, privateKey)
	if err != nil {
		return err
	}

	// Send the transaction
	err = client.SendTransaction(context.Background(), signedTX)
	if err != nil {
		panic(err)
	}

	return nil
}

func newDynamicTX(nonce uint64, tip *big.Int, feeCap *big.Int, gas uint64, to common.Address, value *big.Int,
	signer types2.Signer, sender *ecdsa.PrivateKey) (*types2.Transaction, error) {
	tx, err := types2.SignTx(types2.NewTx(&types2.DynamicFeeTx{
		Nonce:     nonce,
		GasTipCap: tip,
		GasFeeCap: feeCap,
		Gas:       gas,
		To:        &to,
		Value:     value,
	}), signer, sender)
	return tx, err
}
