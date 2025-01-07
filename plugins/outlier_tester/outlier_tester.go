package main

import (
	"autonity-oracle/config"
	"autonity-oracle/helpers"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"time"
)

const (
	version = "v0.2.0"
)

var defaultConfig = config.PluginConfig{
	Name:               "outlier_tester_plugin",
	Timeout:            10, //10s
	DataUpdateInterval: 30, //30s
}

// OutlierTesterPlugin is only used for internal e2e testing,
// it simulates invalid data points for all the symbols in the protocol.
type OutlierTesterPlugin struct {
	version          string
	availableSymbols map[string]struct{}
	symbolSeparator  string
	logger           hclog.Logger
	client           common.DataSourceClient
	conf             *config.PluginConfig
	cachePrices      map[string]types.Price
}

func NewOutlierPlugin(conf *config.PluginConfig, client common.DataSourceClient, version string) *OutlierTesterPlugin {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       conf.Name,
		Level:      hclog.Info,
		Output:     os.Stderr, // logging into stderr thus the go-plugin can redirect the logs to plugin server.
		JSONFormat: true,
	})

	return &OutlierTesterPlugin{
		version:          version,
		logger:           logger,
		client:           client,
		conf:             conf,
		availableSymbols: make(map[string]struct{}),
		cachePrices:      make(map[string]types.Price),
	}
}

func (g *OutlierTesterPlugin) FetchPrices(symbols []string) (types.PluginPriceReport, error) {
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
		decPrice, err := decimal.NewFromString(v.Price)
		if err != nil {
			g.logger.Error("cannot convert price string to decimal: ", "price", v.Price, "error", err.Error())
			continue
		}

		decVol, ok := new(big.Int).SetString(v.Volume, 0)
		if !ok {
			g.logger.Error("cannot convert volume to big.Int: ", "volume", v.Volume)
			continue
		}

		pr := types.Price{
			Timestamp:        now,
			Symbol:           availableSymMap[v.Symbol], // set the symbol with the symbol style used in oracle server side.
			Price:            decPrice,
			RecentVolInUsdcx: decVol,
		}
		g.cachePrices[v.Symbol] = pr
		report.Prices = append(report.Prices, pr)
	}
	report.UnRecognizableSymbols = unRecognizeSymbols
	return report, nil
}

func (g *OutlierTesterPlugin) State(_ int64) (types.PluginStatement, error) {
	var state types.PluginStatement

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
	state.DataSource = g.conf.Scheme + "://" + g.conf.Endpoint
	state.DataSourceType = types.SrcCEX
	return state, nil
}

func (g *OutlierTesterPlugin) Close() {
	if g.client != nil {
		g.client.Close()
	}
}

func (g *OutlierTesterPlugin) fetchPricesFromCache(availableSymbols []string) ([]types.Price, error) {
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
func (g *OutlierTesterPlugin) resolveSymbols(askedSymbols []string) ([]string, []string, map[string]string) {
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

type OutlierClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewOutlierClient(conf *config.PluginConfig) *OutlierClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	return &OutlierClient{conf: conf, client: client, logger: logger}
}

// KeyRequired returns true if the service key is required to access the data source.
func (tc *OutlierClient) KeyRequired() bool {
	return false
}

// FetchPrice is the function fetch prices of the available symbols from data vendor.
func (tc *OutlierClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices
	for _, s := range symbols {
		var price common.Price
		price.Volume = common.DefaultVolume.String()
		price.Symbol = s
		// it is a malicious behaviour to set the price into an outlier range by multiply with 3.0
		p := helpers.ResolveSimulatedPrice(s).Mul(decimal.RequireFromString("3.0"))
		price.Price = p.String()
		prices = append(prices, price)
	}

	return prices, nil
}

// AvailableSymbols is the function to resolve the available symbols from your data vendor.
func (tc *OutlierClient) AvailableSymbols() ([]string, error) {
	res := append(common.DefaultForexSymbols, common.DefaultCryptoSymbols...)
	res = append(res, common.DefaultUSDCSymbol)
	res = append(res, helpers.SymbolBTCETH)
	res = append(res, []string{"ATN-USD", "NTN-USD"}...)
	return res, nil
}

func (tc *OutlierClient) Close() {
	if tc.client != nil && tc.client.Conn != nil {
		tc.client.Conn.Close()
	}
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := NewOutlierPlugin(conf, NewOutlierClient(conf), version)
	defer adapter.Close()

	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{Impl: adapter},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
