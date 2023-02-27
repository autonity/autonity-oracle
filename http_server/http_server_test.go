package httpserver

import (
	"autonity-oracle/config"
	"autonity-oracle/oracle_server"
	"autonity-oracle/types"
	"bytes"
	"encoding/json"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
)

// todo: add tests for http api handlers.
func TestHttpServerAPIHandlers(t *testing.T) {
	err := os.Setenv(types.EnvKeyFile, "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe")
	require.NoError(t, err)
	err = os.Setenv(types.EnvKeyFilePASS, "123")
	require.NoError(t, err)
	conf := config.MakeConfig()
	t.Run("test get version", func(t *testing.T) {
		reqMsg := &types.JSONRPCMessage{
			ID:     json.RawMessage{0},
			Method: "get_version",
			Params: nil,
			Result: nil,
			Error:  "",
		}

		oracle := oracleserver.NewOracleServer(conf.Symbols, ".")
		hs := NewHttpServer(oracle, conf.HTTPPort)

		code, rspMsg := hs.getVersion(reqMsg)
		require.Equal(t, http.StatusOK, code)
		require.Equal(t, reqMsg.ID, rspMsg.ID)
		type Version struct {
			Version string
		}
		dec := json.NewDecoder(bytes.NewReader(rspMsg.Result))
		var v Version
		err := dec.Decode(&v)
		require.NoError(t, err)

		require.Equal(t, hs.oracle.Version(), v.Version)
	})

	t.Run("test get prices", func(t *testing.T) {
		reqMsg := &types.JSONRPCMessage{
			ID:     json.RawMessage{0},
			Method: "get_prices",
			Params: nil,
			Result: nil,
			Error:  "",
		}
		oracle := oracleserver.NewOracleServer(conf.Symbols, ".")
		for _, s := range conf.Symbols {
			price := types.Price{
				Timestamp: 0,
				Symbol:    s,
				Price:     decimal.RequireFromString("12.33"),
			}
			oracle.UpdatePrice(price)
		}
		hs := NewHttpServer(oracle, conf.HTTPPort)
		code, rspMsg := hs.getPrices(reqMsg)
		require.Equal(t, http.StatusOK, code)
		require.Equal(t, reqMsg.ID, rspMsg.ID)
		type PriceAndSymbol struct {
			Prices  types.PriceBySymbol
			Symbols []string
		}

		dec := json.NewDecoder(bytes.NewReader(rspMsg.Result))
		var prices PriceAndSymbol
		err := dec.Decode(&prices)
		require.NoError(t, err)
		for _, s := range conf.Symbols {
			require.Equal(t, s, prices.Prices[s].Symbol)
			require.Equal(t, true, prices.Prices[s].Price.Equals(decimal.RequireFromString("12.33")))
		}
		require.Equal(t, conf.Symbols, prices.Symbols)
	})
}
