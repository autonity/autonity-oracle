package test

import (
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

// integration with l1 network, the reported data should be presented at l1 oracle contract.
func TestDataReporting(t *testing.T) {
	network, err := createNetwork(false)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.Nodes[0].Host, network.Nodes[0].WSPort))
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

	// first test happy case.
	testHappyCaseEndRound := uint64(5)
	testHappyCase(t, o, testHappyCaseEndRound, pricePrecision)

	// test to add new symbols.
	testAddSymbolEndRound := testHappyCaseEndRound + 4
	testAddNewSymbols(t, network, client, o, testAddSymbolEndRound, pricePrecision)

	// test to remove symbols.
	testRMSymbolsEndRound := testAddSymbolEndRound + 4
	testRMSymbols(t, network, client, o, testRMSymbolsEndRound, pricePrecision)

	// test to remove validator from current committee.
	testRMValidatorEndRound := testRMSymbolsEndRound + 40
	testRMValidatorFromCommittee(t, network, client, o, aut, testRMValidatorEndRound, pricePrecision)

	// test to add validator into current committee.
	testNewValidatorAddedEndRound := testRMValidatorEndRound + 40
	testNewValidatorJoinToCommittee(t, network, client, o, aut, testNewValidatorAddedEndRound, pricePrecision)
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

func testAddNewSymbols(t *testing.T, network *Network, client *ethclient.Client, o *contract.Oracle, beforeRound uint64,
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

	newCommitteeSize := int64(len(network.Nodes) / 2)
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

	newCommitteeSize := int64(len(network.Nodes))
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

		// todo: newly added validator shouldn't be slashed if they does not omit any report.
		break
	}
}

// todo: missing report by omission voter, the faulty node should be slashed.
func TestOmissionVoter(t *testing.T) {

}
