package test

import (
	"autonity-oracle/config"
	"autonity-oracle/http_server"
	"autonity-oracle/oracle_server"
	"autonity-oracle/types"
	"context"
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

}