# To write a new plugin
The oracle server provides a plugin framework for the development and management of data adaptors for loading data from external data source providers. Runtime management of plugins is dynamic; the server discovers and loads new or updated data source adaptors during runtime, so there is no need to restart the oracle server to detect data adaptor changes.

## The plugin configuration
Every plugin has a unified configuration, it is configured by oracle server in the plugins-conf.yml file. When you implement a new plugin, you will need to prepare the default values for the plugin, for example:
```go
var defaultConfig = types.PluginConfig{
	Key:                "",
	Scheme:             "https",
	Endpoint:           "127.0.0.1:8080",
	Timeout:            10, //10s
	DataUpdateInterval: 30, //30s
}
```
where:
- Key is a string contains the data access service key of your data provider, keep it empty if the provider does not require one.
- Scheme is either "http" or "https" depends on your data provider service.
- Endpoint is the host endpoint of your data provider's service.
- Timeout is the timer in seconds to cancel a single data request RPC.
- DataUpdateInterval is the interval in seconds to fetch data from the data provider, it is very useful for those rate limited data service.

## Interface
The interface in between the oracle server and the plugin are simple:
```go
// Adapter is the interface that we're exposing as a plugin.
type Adapter interface {
	// FetchPrices is called by oracle server to fetch data points of symbols required by the protocol contract, some
	// symbols in the protocol's symbols set might not be recognisable by a data source, thus in the plugin implementation,
	// one need to filter invalid symbols to make sure valid symbol's data be collected. The return PluginPriceReport
	// contains the valid symbols' prices and a set of invalid symbols.
	FetchPrices(symbols []string) (PluginPriceReport, error)
	// State is called by oracle server to instantiate a plugin, it returns the plugin's version and a set of symbol that
	// the data source support, this information is just used for logging at the bootstrap phase of a plugin.
	State() (PluginState, error)
}

// PluginPriceReport is the returned data samples from adapters which carry the prices and those symbols of no data if
// there are any unrecognisable symbols from the data source side.
type PluginPriceReport struct {
	Prices                []Price
	UnRecognizableSymbols []string
}

// in autonity-oracle/types/types.go, there is a type Price:
// Price is the structure contains the exchange rate of a symbol with a timestamp at which the sampling happens.
type Price struct {
	Timestamp int64 // TS on when the data is being sampled in time's seconds since Jan 1 1970 (Unix time).
	Symbol    string
	Price     decimal.Decimal
}

// PluginState is the returned data when the oracle host want to initialise the plugin with basic information: version,
// and available symbols that the data source support.
type PluginState struct {
	KeyRequired      bool
	Version          string
	AvailableSymbols []string
}
```
where the statements in PluginState:
- KeyRequired states if the plugin needs a service key to be configured. When the plugin requires a key, oracle server will not start the plugin if a key is missing from the configuration.
- Version states the version of the plugin.
- AvailableSymbols states the available symbols of current data plugin.

## Implement a plugin
Create a directory for your plugin under the autonity-oracle/plugins directory. There is a template_plugin directory
which contains a go source code file template_plugin.go that can be used as a template. To implement a new plugin, there are 4 steps described beneath.
### Create a Plugin structure
```go
// TemplatePlugin Here is an implementation of a plugin which returns simulated data points.
type TemplatePlugin struct {
	version          string
	availableSymbols map[string]struct{}
	symbolSeparator  string
	logger           hclog.Logger
	client           common.DataSourceClient
	conf             *types.PluginConfig
	cachePrices      map[string]types.Price
}
```
### Implement the 2 interfaces for the Plugin structure
In this template plugin, we leave only 3 todos to be implemented by plugin developer to adapt with real data vendor.
```go
func (g *TemplatePlugin) FetchPrices(symbols []string) (types.PluginPriceReport, error) {
	var report types.PluginPriceReport

	availableSymbols, badSymbols, availableSymMap := g.resolveSymbols(symbols)
	if len(availableSymbols) == 0 {
		g.logger.Warn("no available symbols from plugin", "plugin", g.conf.Name)
		report.BadSymbols = badSymbols
		return report, fmt.Errorf("no available symbols")
	}

	res, err := g.client.FetchPrice(availableSymbols)
	if err != nil {
		return report, err
	}

	g.logger.Debug("sampled data points", res)

	now := time.Now().Unix()
	for _, v := range res {
		dec, err := decimal.NewFromString(v.Price)
		if err != nil {
			g.logger.Error("cannot convert price string to decimal: ", v.Price, err)
			continue
		}
		report.Prices = append(report.Prices, types.Price{
			Timestamp: now,
			Symbol:    availableSymMap[v.Symbol], // convert the symbol to raw symbol style in oracle server side.
			Price:     dec,
		})
	}
	report.BadSymbols = badSymbols
	return report, nil
}

func (g *TemplatePlugin) State() (types.PluginState, error) {
	var state types.PluginState

	symbols, err := g.client.AvailableSymbols()
	if err != nil {
		return state, err
	}

	if len(g.availableSymbols) != 0 {
		for k := range g.availableSymbols {
		symbol := k
		delete(g.availableSymbols, symbol)
		}
	}

	for _, s := range symbols {
		g.availableSymbols[s] = struct{}{}
	}

	for _, symbol := range symbols {
		if len(symbol) != 0 {
			g.symbolSeparator = common.ResolveSeparator(symbol)
			break
		}
	}

	state.KeyRequired = g.client.KeyRequired()
	state.Version = g.version
	state.AvailableSymbols = symbols

	return state, nil
}

// KeyRequired returns true if the service key is required to access the data source.
func (tc *TemplateClient) KeyRequired() bool {
	return false // return true if your data provider asked for a service key.
}

```
### Instantiate the Plugin and Register it.
In the main() function, which is the entry point of your plugin, initialize the plugin structure, and register it in the go-plugin framework:
```go
func main() {
	conf, err := common.LoadPluginConf(os.Args[0])
	if err != nil {
		println("cannot load conf: ", err.Error(), os.Args[0])
		os.Exit(-1)
	}

	resolveConf(&conf)

	client := NewTemplateClient(&conf)
	adapter := NewTemplatePlugin(&conf, client, version)
	defer adapter.Close()

	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{Impl: adapter},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
```
Handshake config is in autonity-oracle/types/plugin_spec.go
```go
// HandshakeConfig are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user-friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}
```

## The full code
```go
package main

import (
	"autonity-oracle/helpers"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shopspring/decimal"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	version       = "v0.0.2"
	defaultScheme = "https"
	defaultHost   = "127.0.0.1:8080"
	//apiPath               = "api/v3/ticker/price"
	defaultTimeout        = 10 // 10s
	defaultUpdateInterval = 30 // 30s
	defaultKey            = ""
)

var (
	defaultForex  = []string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"}
	defaultCrypto = []string{"ATN-USD", "NTN-USD"}
)

// TemplatePlugin Here is an implementation of a plugin which returns simulated data points.
type TemplatePlugin struct {
	version          string
	availableSymbols map[string]struct{}
	symbolSeparator  string
	logger           hclog.Logger
	client           common.DataSourceClient
	conf             *types.PluginConfig
	cachePrices      map[string]types.Price
}

func NewTemplatePlugin(conf *types.PluginConfig, client common.DataSourceClient, version string) *TemplatePlugin {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       conf.Name,
		Level:      hclog.Debug,
		Output:     os.Stderr, // logging into stderr thus the go-plugin can redirect the logs to plugin server.
		JSONFormat: true,
	})

	return &TemplatePlugin{
		version:          version,
		logger:           logger,
		client:           client,
		conf:             conf,
		availableSymbols: make(map[string]struct{}),
	}
}

func (g *TemplatePlugin) FetchPrices(symbols []string) (types.PluginPriceReport, error) {
	var report types.PluginPriceReport

	availableSymbols, badSymbols, availableSymMap := g.resolveSymbols(symbols)
	if len(availableSymbols) == 0 {
		g.logger.Warn("no available symbols from plugin", "plugin", g.conf.Name)
		report.BadSymbols = badSymbols
		return report, fmt.Errorf("no available symbols")
	}

	res, err := g.client.FetchPrice(availableSymbols)
	if err != nil {
		return report, err
	}

	g.logger.Debug("sampled data points", res)

	now := time.Now().Unix()
	for _, v := range res {
		dec, err := decimal.NewFromString(v.Price)
		if err != nil {
			g.logger.Error("cannot convert price string to decimal: ", v.Price, err)
			continue
		}
		report.Prices = append(report.Prices, types.Price{
			Timestamp: now,
			Symbol:    availableSymMap[v.Symbol], // convert the symbol to raw symbol style in oracle server side.
			Price:     dec,
		})
	}
	report.BadSymbols = badSymbols
	return report, nil
}

func (g *TemplatePlugin) State() (types.PluginState, error) {
	var state types.PluginState

	symbols, err := g.client.AvailableSymbols()
	if err != nil {
		return state, err
	}

	if len(g.availableSymbols) != 0 {
		for k := range g.availableSymbols {
			symbol := k
			delete(g.availableSymbols, symbol)
		}
	}

	for _, s := range symbols {
		g.availableSymbols[s] = struct{}{}
	}

	for _, symbol := range symbols {
		if len(symbol) != 0 {
			g.symbolSeparator = common.ResolveSeparator(symbol)
			break
		}
	}

	state.Version = g.version
	state.AvailableSymbols = symbols

	return state, nil
}

func (g *TemplatePlugin) Close() {
	if g.client != nil {
		g.client.Close()
	}
}

// resolveSymbols resolve supported symbols of provider, and it builds the mapping of symbols from `-` separated pattern to those
// pattens supported by data providers, and filter outs those un-supported symbols.
func (g *TemplatePlugin) resolveSymbols(askedSymbols []string) ([]string, []string, map[string]string) {
	var supported []string
	var unSupported []string

	symbolsMapping := make(map[string]string)

	for _, askedSym := range askedSymbols {

		converted := common.ConvertSymbol(askedSym, g.symbolSeparator)
		if _, ok := g.availableSymbols[converted]; !ok {
			unSupported = append(unSupported, askedSym)
			continue
		}
		supported = append(supported, converted)
		symbolsMapping[converted] = askedSym
	}
	return supported, unSupported, symbolsMapping
}

func resolveConf(conf *types.PluginConfig) {

	if conf.Timeout == 0 {
		conf.Timeout = defaultTimeout
	}

	if conf.DataUpdateInterval == 0 {
		conf.DataUpdateInterval = defaultUpdateInterval
	}

	if len(conf.Scheme) == 0 {
		conf.Scheme = defaultScheme
	}

	if len(conf.Endpoint) == 0 {
		conf.Endpoint = defaultHost
	}

	if len(conf.Key) == 0 {
		conf.Key = defaultKey
	}
}

type TemplateClient struct {
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewTemplateClient(conf *types.PluginConfig) *TemplateClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	if client == nil {
		panic("cannot create client for exchange rate api")
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	return &TemplateClient{conf: conf, client: client, logger: logger}
}

//FetchPrice is the function fetch prices of the available symbols from data vendor.
func (tc *TemplateClient) FetchPrice(symbols []string) (common.Prices, error) {
	// todo: implement this function by plugin developer.
	/*
		var prices common.Prices
		u, err := tc.buildURL(symbols)
		if err != nil {
			return nil, err
		}

		res, err := tc.client.Conn.Request(tc.conf.Scheme, u)
		if err != nil {
			tc.logger.Error("https get", "error", err.Error())
			return nil, err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			tc.logger.Error("io read", "error", err.Error())
			return nil, err
		}

		err = json.Unmarshal(body, &prices)
		if err != nil {
			return nil, err
		}

		tc.logger.Info("binance", "data", prices)
	*/

	// in this template, we just return fix values.
	var prices common.Prices
	for _, s := range symbols {
		var price common.Price
		price.Symbol = s
		price.Price = helpers.ResolveSimulatedPrice(s).String()
		prices = append(prices, price)
	}

	return prices, nil
}

//AvailableSymbols is the function to resolve the available symbols from your data vendor.
func (tc *TemplateClient) AvailableSymbols() ([]string, error) {
	// todo: implement this function by plugin developer.
	/*
		var res []string
		prices, err := tc.FetchPrice(nil)
		if err != nil {
			return nil, err
		}

		for _, p := range prices {
			res = append(res, p.Symbol)
		}*/
	// in this template, we just return the simulated symbols inside this plugin.
	res := append(defaultForex, defaultCrypto...)
	return res, nil
}

func (tc *TemplateClient) Close() {
	if tc.client != nil && tc.client.Conn != nil {
		tc.client.Conn.Close()
	}
}

// this is the function build the url to access your remote data provider's data api.
func (tc *TemplateClient) buildURL(symbols []string) (*url.URL, error) { //no lint
	// todo: implement this function by plugin developer.
	/*
		endpoint := &url.URL{}
		endpoint.Path = apiPath

		if len(symbols) != 0 {
			parameters, err := json.Marshal(symbols)
			if err != nil {
				return nil, err
			}

			query := endpoint.Query()
			query.Set("symbol", string(parameters))
			endpoint.RawQuery = query.Encode()
		}*/

	// in this template, we just return a default url since in this template we just return simulated values rather than
	// rise http request to get real data from a data provider.
	endpoint := &url.URL{}
	return endpoint, nil
}

func main() {
	conf, err := common.LoadPluginConf(os.Args[0])
	if err != nil {
		println("cannot load conf: ", err.Error(), os.Args[0])
		os.Exit(-1)
	}

	resolveConf(&conf)

	client := NewTemplateClient(&conf)
	adapter := NewTemplatePlugin(&conf, client, version)
	defer adapter.Close()

	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{Impl: adapter},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
	})
}

```
## Build it
In the autonity-oracle directory, run:
```shell
go build -o ./build/bin/plugins/template_plugin ./plugins/template_plugin/template_plugin.go
```
You will find a binary named `template_plugin` under the directory: ./build/bin/plugins
## Use it
In production, after you have built the plugin binary, then just copy it in to the plugins directory that is scanned by the oracle server. It will be discovered and loaded automatically.
