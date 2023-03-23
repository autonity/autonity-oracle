package e2e_test_test

import (
	"autonity-oracle/config"
	"autonity-oracle/helpers"
	"autonity-oracle/http_server"
	"autonity-oracle/oracle_server"
	"autonity-oracle/types"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestE2EAutonityOracleServer(t *testing.T) {
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

func testReplacePlugin(t *testing.T, port int, pluginDir string) {
	// get the plugins before replacement.
	reqMsg := &types.JSONRPCMessage{Method: "list_plugins"}
	respMsg, err := httpPost(t, reqMsg, port)
	require.NoError(t, err)
	var pluginsAtStart types.PluginByName
	err = json.Unmarshal(respMsg.Result, &pluginsAtStart)
	require.NoError(t, err)

	// do the replacement of plugins.
	err = replacePlugins(pluginDir)
	require.NoError(t, err)
	// wait for replaced plugins to be loaded.
	time.Sleep(10 * time.Second)

	respMsg, err = httpPost(t, reqMsg, port)
	require.NoError(t, err)
	var pluginsAfterReplace types.PluginByName
	err = json.Unmarshal(respMsg.Result, &pluginsAfterReplace)
	require.NoError(t, err)

	for k, p := range pluginsAfterReplace {
		require.Equal(t, p.Name, pluginsAtStart[k].Name)
		require.Equal(t, true, p.StartAt.After(pluginsAtStart[k].StartAt))
	}
}

func testAddPlugin(t *testing.T, port int, pluginDir string) {
	clonerPrefix := "clone"
	clonedPlugins, err := clonePlugins(pluginDir, clonerPrefix, pluginDir)
	defer func() {
		for _, f := range clonedPlugins {
			os.Remove(f) // nolint
		}
	}()

	require.NoError(t, err)
	require.Equal(t, true, len(clonedPlugins) > 0)
	// wait for cloned plugins to be loaded.
	time.Sleep(10 * time.Second)
	testListPlugins(t, port, pluginDir)
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
		require.Equal(t, true, ps.Prices[s].Price.Equal(helpers.ResolveSimulatedPrice(s)))
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

func testListPlugins(t *testing.T, port int, pluginDir string) {
	reqMsg := &types.JSONRPCMessage{Method: "list_plugins"}

	respMsg, err := httpPost(t, reqMsg, port)
	require.NoError(t, err)
	var plugins types.PluginByName
	err = json.Unmarshal(respMsg.Result, &plugins)
	require.NoError(t, err)
	files, err := listPluginDir(pluginDir)
	require.NoError(t, err)
	require.Equal(t, len(files), len(plugins))
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

func replacePlugins(pluginDir string) error {
	rawPlugins, err := listPluginDir(pluginDir)
	if err != nil {
		return err
	}

	clonePrefix := "clone"
	clonedPlugins, err := clonePlugins(pluginDir, clonePrefix, fmt.Sprintf("%s/..", pluginDir))
	if err != nil {
		return err
	}

	for _, file := range clonedPlugins {
		for _, info := range rawPlugins {
			if strings.Contains(file, info.Name()) {
				err := os.Rename(file, fmt.Sprintf("%s/%s", pluginDir, info.Name()))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func clonePlugins(pluginDIR string, clonePrefix string, destDir string) ([]string, error) {

	var clonedPlugins []string
	files, err := listPluginDir(pluginDIR)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// read srcFile
		srcContent, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", pluginDIR, file.Name()))
		if err != nil {
			return clonedPlugins, err
		}

		// create dstFile and copy the content
		newPlugin := fmt.Sprintf("%s/%s%s", destDir, clonePrefix, file.Name())
		err = ioutil.WriteFile(newPlugin, srcContent, file.Mode())
		if err != nil {
			return clonedPlugins, err
		}
		clonedPlugins = append(clonedPlugins, newPlugin)
	}
	return clonedPlugins, nil
}

func listPluginDir(pluginDIR string) ([]fs.FileInfo, error) {
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
