package test

import (
	"autonity-oracle/config"
	"autonity-oracle/helpers"
	"autonity-oracle/http_server"
	"autonity-oracle/oracle_server"
	"autonity-oracle/reporter"
	contract "autonity-oracle/reporter/contract"
	"autonity-oracle/types"
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestOracleServerPluginTest(t *testing.T) {
	err := os.Unsetenv("ORACLE_HTTP_PORT")
	require.NoError(t, err)
	err = os.Unsetenv("ORACLE_CRYPTO_SYMBOLS")
	require.NoError(t, err)
	err = os.Setenv(types.EnvKeyFile, "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe")
	require.NoError(t, err)
	err = os.Setenv(types.EnvKeyFilePASS, "123")
	require.NoError(t, err)
	conf := config.MakeConfig()
	conf.PluginDIR = "../plugins/fakeplugin/bin"
	// create oracle service and start the ticker job.
	oracle := oracleserver.NewOracleServer(conf.Symbols, conf.PluginDIR)
	go oracle.Start()
	defer oracle.Stop()

	// create http service.
	srv := httpserver.NewHttpServer(oracle, conf.HTTPPort)
	srv.StartHTTPServer()

	// wait for the http service to be loaded.
	time.Sleep(25 * time.Second)

	testGetVersion(t, conf.HTTPPort)

	testListPlugins(t, conf.HTTPPort, conf.PluginDIR)

	testGetPrices(t, conf.HTTPPort)

	testReplacePlugin(t, conf.HTTPPort, conf.PluginDIR)

	testAddPlugin(t, conf.HTTPPort, conf.PluginDIR)

	defer srv.Shutdown(context.Background()) //nolint
}

// integration with l1 network, the reported data should be presented at l1 oracle contract.
func TestDataReportingHappyCase(t *testing.T) {
	network, err := createNetwork(false)
	require.NoError(t, err)
	defer network.Stop()

	client, err := ethclient.Dial(fmt.Sprintf("ws://%s:%d", network.Nodes[0].Host, network.Nodes[0].WSPort))
	require.NoError(t, err)

	// bind client with oracle contract address
	o, err := contract.NewOracle(reporter.ContractAddress, client)
	require.NoError(t, err)

	endingRound := uint64(5)
	for {
		time.Sleep(1 * time.Minute)
		round, err := o.GetRound(nil)
		require.NoError(t, err)

		// continue to wait until end round.
		if round.Uint64() < endingRound {
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
			require.True(t, true, price.Div(reporter.PricePrecision).Equal(helpers.ResolveSimulatedPrice(s)))
			require.Equal(t, uint64(0), rd[i].Status.Uint64())
		}
		break
	}
}

// todo: add new symbols
func TestAddNewSymbols(t *testing.T) {

}

// todo: remove symbols
func TestRemoveSymbols(t *testing.T) {

}

// todo: test add new validators in committee
func TestAddNewValidator(t *testing.T) {

}

// todo: test remove validators from committee.
func TestRemoveValidator(t *testing.T) {

}

// todo: missing report by omission voter
func TestOmissionVoter(t *testing.T) {

}
