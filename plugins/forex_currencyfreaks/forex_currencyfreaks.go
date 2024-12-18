package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	version    = "v0.2.0"
	apiVersion = "v2.0/rates/latest"
	apiKey     = "apikey"
)

var defaultConfig = config.PluginConfig{
	Name:               "forex_currencyfreaks",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "api.currencyfreaks.com",
	Timeout:            10, //10s
	DataUpdateInterval: 30, //30s
}

type CFResult struct {
	Date  string  `json:"date"`
	Base  string  `json:"base"`
	Rates CFRates `json:"rates"`
}

type CFRates struct {
	EUR string `json:"EUR"`
	JPY string `json:"JPY"`
	GBP string `json:"GBP"`
	AUD string `json:"AUD"`
	CAD string `json:"CAD"`
	SEK string `json:"SEK"`
}

type CFClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewCFClient(conf *config.PluginConfig) *CFClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &CFClient{conf: conf, client: client, logger: logger}
}

func (cf *CFClient) KeyRequired() bool {
	return true
}

func (cf *CFClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices
	u := cf.buildURL(cf.conf.Key)
	res, err := cf.client.Conn.Request(cf.conf.Scheme, u)
	if err != nil {
		cf.logger.Error("https request", "error", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	if err = common.CheckHTTPStatusCode(res.StatusCode); err != nil {
		cf.logger.Error("data source return error", "error", err.Error())
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		cf.logger.Error("io read", "error", err.Error())
		return nil, err
	}

	var result CFResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		cf.logger.Error("unmarshal price", "error", err.Error())
		return nil, err
	}

	if result.Date == "" {
		cf.logger.Error("data source returns", "data", string(body))
		return nil, common.ErrDataNotAvailable
	}

	for _, s := range symbols {
		p, err := cf.symbolsToPrice(s, &result)
		if err != nil {
			cf.logger.Error("symbol to price", "error", err.Error())
			continue
		}
		prices = append(prices, p)
	}
	return prices, nil
}

// AvailableSymbols returns the adapted symbols for current data source.
func (cf *CFClient) AvailableSymbols() ([]string, error) {
	return common.DefaultForexSymbols, nil
}

func (cf *CFClient) Close() {
	cf.client.Conn.Close()
}

func (cf *CFClient) symbolsToPrice(s string, res *CFResult) (common.Price, error) {
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
		pUE, err := decimal.NewFromString(res.Rates.EUR)
		if err != nil {
			return price, err
		}
		price.Price = decimal.NewFromInt(1).Div(pUE).String()
	case "JPY":
		pUJ, err := decimal.NewFromString(res.Rates.JPY)
		if err != nil {
			return price, err
		}
		price.Price = decimal.NewFromInt(1).Div(pUJ).String()
	case "GBP":
		pUG, err := decimal.NewFromString(res.Rates.GBP)
		if err != nil {
			return price, err
		}
		price.Price = decimal.NewFromInt(1).Div(pUG).String()
	case "AUD":
		pUA, err := decimal.NewFromString(res.Rates.AUD)
		if err != nil {
			return price, err
		}
		price.Price = decimal.NewFromInt(1).Div(pUA).String()
	case "CAD":
		pUC, err := decimal.NewFromString(res.Rates.CAD)
		if err != nil {
			return price, err
		}
		price.Price = decimal.NewFromInt(1).Div(pUC).String()
	case "SEK":
		pUS, err := decimal.NewFromString(res.Rates.SEK)
		if err != nil {
			return price, err
		}
		price.Price = decimal.NewFromInt(1).Div(pUS).String()
	default:
		return price, fmt.Errorf("unknown symbol %s", from)
	}
	return price, nil
}

func (cf *CFClient) buildURL(key string) *url.URL {
	endpoint := &url.URL{}
	endpoint.Path = apiVersion

	query := endpoint.Query()
	query.Set(apiKey, key)
	endpoint.RawQuery = query.Encode()
	return endpoint
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewCFClient(conf), version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
