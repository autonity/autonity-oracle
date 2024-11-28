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
	"time"
)

const (
	version = "v0.2.0"
)

var defaultConfig = types.PluginConfig{
	Name:               "pluginBinaryName",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "127.0.0.1:8080",
	Timeout:            10, //10s
	DataUpdateInterval: 30, //30s
}

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
		Level:      hclog.Info,
		Output:     os.Stderr, // logging into stderr thus the go-plugin can redirect the logs to plugin server.
		JSONFormat: true,
	})

	return &TemplatePlugin{
		version:          version,
		logger:           logger,
		client:           client,
		conf:             conf,
		availableSymbols: make(map[string]struct{}),
		cachePrices:      make(map[string]types.Price),
	}
}

func (g *TemplatePlugin) FetchPrices(symbols []string) (types.PluginPriceReport, error) {
	var report types.PluginPriceReport

	availableSymbols, unRecognizeSymbols, availableSymMap := g.resolveSymbols(symbols)
	if len(availableSymbols) == 0 {
		report.UnRecognizableSymbols = unRecognizeSymbols
		return report, common.ErrKnownSymbols
	}

	cPRs, err := g.fetchPricesFromCache(availableSymbols)
	if err == nil {
		report.Prices = cPRs
		report.UnRecognizableSymbols = unRecognizeSymbols
		return report, nil
	}

	// fetch data from data source.
	res, err := g.client.FetchPrice(availableSymbols)
	if err != nil {
		return report, err
	}

	now := time.Now().Unix()
	for _, v := range res {
		dec, err := decimal.NewFromString(v.Price)
		if err != nil {
			g.logger.Error("cannot convert price string to decimal: ", "price", v.Price, "error", err.Error())
			continue
		}

		pr := types.Price{
			Timestamp: now,
			Symbol:    availableSymMap[v.Symbol], // set the symbol with the symbol style used in oracle server side.
			Price:     dec,
		}
		g.cachePrices[v.Symbol] = pr
		report.Prices = append(report.Prices, pr)
	}
	report.UnRecognizableSymbols = unRecognizeSymbols
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

func (g *TemplatePlugin) Close() {
	if g.client != nil {
		g.client.Close()
	}
}

func (g *TemplatePlugin) fetchPricesFromCache(availableSymbols []string) ([]types.Price, error) {
	var prices []types.Price
	now := time.Now().Unix()
	for _, s := range availableSymbols {
		pr, ok := g.cachePrices[s]
		if !ok {
			return nil, fmt.Errorf("no data buffered")
		}

		if now-pr.Timestamp > int64(g.conf.DataUpdateInterval) {
			return nil, fmt.Errorf("data is too old")
		}

		prices = append(prices, pr)
	}
	return prices, nil
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

type TemplateClient struct {
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewTemplateClient(conf *types.PluginConfig) *TemplateClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	return &TemplateClient{conf: conf, client: client, logger: logger}
}

// KeyRequired returns true if the service key is required to access the data source.
func (tc *TemplateClient) KeyRequired() bool {
	return false
}

// FetchPrice is the function fetch prices of the available symbols from data vendor.
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

// AvailableSymbols is the function to resolve the available symbols from your data vendor.
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
	// Put all the supported symbols here, as this template plugin is used by simulations and e2e testing,
	// we add some symbols required for the test as well.
	res := append(common.DefaultForexSymbols, common.DefaultCryptoSymbols...)
	res = append(res, common.DefaultUSDCSymbol)
	res = append(res, types.SymbolBTCETH)
	return res, nil
}

func (tc *TemplateClient) Close() {
	if tc.client != nil && tc.client.Conn != nil {
		tc.client.Conn.Close()
	}
}

// this is the function build the url to access your remote data provider's data api.
func (tc *TemplateClient) buildURL(symbols []string) (*url.URL, error) { //nolint
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
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := NewTemplatePlugin(conf, NewTemplateClient(conf), version)
	defer adapter.Close()

	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{Impl: adapter},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
