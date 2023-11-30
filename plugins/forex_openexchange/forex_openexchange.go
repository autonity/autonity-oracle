package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shopspring/decimal"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	version = "v0.0.1"
	api     = "api/latest.json"
	base    = "base"
	appID   = "app_id"
)

var defaultConfig = types.PluginConfig{
	Key:                "0be02ca33c4843ee968c4cedd2686f01",
	Scheme:             "https",
	Endpoint:           "openexchangerates.org",
	Timeout:            10, //10s
	DataUpdateInterval: 30, //30s
}

type ConversionRates struct {
	EUR decimal.Decimal `json:"EUR"`
	JPY decimal.Decimal `json:"JPY"`
	GBP decimal.Decimal `json:"GBP"`
	AUD decimal.Decimal `json:"AUD"`
	CAD decimal.Decimal `json:"CAD"`
	SEK decimal.Decimal `json:"SEK"`
}

type OEResult struct {
	Disclaimer string          `json:"disclaimer"`
	License    string          `json:"license"`
	Timestamp  int64           `json:"timestamp"`
	Base       string          `json:"base"`
	Rates      ConversionRates `json:"rates"`
}

type OXClient struct {
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewOXClient(conf *types.PluginConfig) *OXClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	if client == nil {
		panic("cannot create client for open exchange rate api")
	}
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "OpenExchangeRate",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	return &OXClient{
		conf:   conf,
		client: client,
		logger: logger,
	}
}

func (oe *OXClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices
	u := oe.buildURL(oe.conf.Key)
	res, err := oe.client.Conn.Request(oe.conf.Scheme, u)
	if err != nil {
		oe.logger.Error("https request", "error", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		oe.logger.Error("io read", "error", err.Error())
		return nil, err
	}

	var result OEResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		oe.logger.Error("unmarshal price", "error", err.Error())
		return nil, err
	}

	if result.Timestamp == 0 {
		oe.logger.Error("data source returns", "data", string(body))
		return nil, common.ErrDataNotAvailable
	}

	for _, s := range symbols {
		p, err := oe.symbolsToPrice(s, &result)
		if err != nil {
			oe.logger.Error("symbol to price", "error", err.Error())
			continue
		}
		prices = append(prices, p)
	}
	oe.logger.Info("open exchange api", "data", prices)
	return prices, nil
}

// AvailableSymbols returns the adapted symbols for current data source.
func (oe *OXClient) AvailableSymbols() ([]string, error) {
	return common.DefaultForexSymbols, nil
}
func (oe *OXClient) Close() {
	oe.client.Conn.Close()
}

func (oe *OXClient) symbolsToPrice(s string, res *OEResult) (common.Price, error) {
	var price common.Price
	sep := common.ResolveSeparator(s)
	codes := strings.Split(s, sep)
	if len(codes) != 2 {
		return price, fmt.Errorf("invalid symbol %s", s)
	}

	from := codes[0]
	to := codes[1]
	if to != res.Base {
		return price, fmt.Errorf("wrong base %s", to)
	}
	price.Symbol = s
	switch from {
	case "EUR":
		price.Price = decimal.NewFromInt(1).Div(res.Rates.EUR).String()
	case "JPY":
		price.Price = decimal.NewFromInt(1).Div(res.Rates.JPY).String()
	case "GBP":
		price.Price = decimal.NewFromInt(1).Div(res.Rates.GBP).String()
	case "AUD":
		price.Price = decimal.NewFromInt(1).Div(res.Rates.AUD).String()
	case "CAD":
		price.Price = decimal.NewFromInt(1).Div(res.Rates.CAD).String()
	case "SEK":
		price.Price = decimal.NewFromInt(1).Div(res.Rates.SEK).String()
	default:
		return price, fmt.Errorf("unknown symbol %s", from)
	}
	return price, nil
}

func (oe *OXClient) buildURL(apiKey string) *url.URL {
	endpoint := &url.URL{}
	endpoint.Path = api

	query := endpoint.Query()
	query.Set(base, "USD")
	query.Set(appID, apiKey)
	endpoint.RawQuery = query.Encode()
	return endpoint
}

func main() {
	conf, err := common.LoadPluginConf(os.Args[0])
	if err != nil {
		println("cannot load conf: ", err.Error(), os.Args[0])
		os.Exit(-1)
	}
	common.ResolveConf(&conf, &defaultConfig)

	client := NewOXClient(&conf)
	adapter := common.NewPlugin(&conf, client, version)
	defer adapter.Close()

	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{Impl: adapter},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
