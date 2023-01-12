package main_test

import (
	"autonity-oracle/config"
	"autonity-oracle/http_server"
	"autonity-oracle/oracle_server"
	"autonity-oracle/types"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAutonityOracleAPIs(t *testing.T) {
	//t.Skip("it fails in ci job environment, but works in local environment, todo e2e test via docker.")
	err := os.Unsetenv("ORACLE_HTTP_PORT")
	require.NoError(t, err)
	err = os.Unsetenv("ORACLE_CRYPTO_SYMBOLS")
	require.NoError(t, err)

	conf := config.MakeConfig()
	// create oracle service and start the ticker job.
	oracle := oracleserver.NewOracleServer(conf.Symbols, "./e2e_test_plugins/")
	go oracle.Start()
	defer oracle.Stop()

	// create http service.
	srv := httpserver.NewHttpServer(oracle, conf.HTTPPort)
	srv.StartHTTPServer()

	// wait for the http service to be loaded.
	time.Sleep(25 * time.Second)

	testGetVersion(t, conf.HTTPPort)

	testGetPrices(t, conf.HTTPPort)

	testUpdateSymbols(t, conf.HTTPPort)

	defer srv.Shutdown(context.Background()) //nolint
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
		require.Equal(t, true, ps.Prices[s].Price.Equal(types.SimulatedPrice))
	}
}

func testUpdateSymbols(t *testing.T, port int) {
	newSymbols := []string{"NTNETH", "NTNBTC", "NTNBNB"}
	encSymbols, err := json.Marshal(newSymbols)
	require.NoError(t, err)

	reqMsg := &types.JSONRPCMessage{
		Method: "update_symbols",
		Params: encSymbols,
	}

	respMsg, err := httpPost(t, reqMsg, port)
	require.NoError(t, err)
	var symbols []string
	err = json.Unmarshal(respMsg.Result, &symbols)
	require.NoError(t, err)
	require.Equal(t, newSymbols, symbols)
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
