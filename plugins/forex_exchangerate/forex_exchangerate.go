package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
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
	version   = "v0.0.1"
	exVersion = "v6"
)

var defaultConfig = types.PluginConfig{
	Key:                "411f04e4775bb86c20296530",
	Scheme:             "https",
	Endpoint:           "v6.exchangerate-api.com",
	Timeout:            10, //10s
	DataUpdateInterval: 30, //30s
}

type EXResult struct {
	Result             string          `json:"result"`
	Documentation      string          `json:"documentation"`
	Term               string          `json:"terms_of_use"`
	TimeLastUpdateUnix int64           `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string          `json:"time_last_update_utc"`
	TimeNextUpdateUnix int64           `json:"time_next_update_unix"`
	TimeNextUpdateUTC  string          `json:"time_next_update_utc"`
	Base               string          `json:"base_code"`
	Rates              ConversionRates `json:"conversion_rates"`
}

type ConversionRates struct {
	EUR decimal.Decimal `json:"EUR"`
	JPY decimal.Decimal `json:"JPY"`
	GBP decimal.Decimal `json:"GBP"`
	AUD decimal.Decimal `json:"AUD"`
	CAD decimal.Decimal `json:"CAD"`
	SEK decimal.Decimal `json:"SEK"`
}

type EXClient struct {
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewEXClient(conf *types.PluginConfig) *EXClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	if client == nil {
		panic(fmt.Sprintf("cannot create client for %s", conf.Endpoint))
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "ExchangeClient",
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &EXClient{conf: conf, client: client, logger: logger}
}

func (ex *EXClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices
	u := ex.buildURL(ex.conf.Key)

	res, err := ex.client.Conn.Request(ex.conf.Scheme, u)
	if err != nil {
		ex.logger.Error("request", "error", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	if err = common.CheckHTTPStatusCode(res.StatusCode); err != nil {
		ex.logger.Error("data source return error", "error", err.Error())
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		ex.logger.Error("read", "error", err.Error())
		return nil, err
	}

	var result EXResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		ex.logger.Error("unmarshal", "error", err.Error())
		return nil, err
	}

	if result.Result != "success" {
		ex.logger.Error("data source returns", "data", string(body))
		return nil, common.ErrDataNotAvailable
	}

	for _, s := range symbols {
		p, err := ex.symbolsToPrice(s, &result)
		if err != nil {
			ex.logger.Error("unify price format", "error", err.Error())
			continue
		}
		prices = append(prices, p)
	}

	return prices, nil
}

// AvailableSymbols returns the adapted symbols for current data source.
func (ex *EXClient) AvailableSymbols() ([]string, error) {
	return common.DefaultForexSymbols, nil
}
func (ex *EXClient) Close() {
	ex.client.Conn.Close()
}

func (ex *EXClient) symbolsToPrice(s string, res *EXResult) (common.Price, error) {
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

func (ex *EXClient) buildURL(apiKey string) *url.URL {
	endpoint := &url.URL{}
	endpoint.Path = exVersion + fmt.Sprintf("/%s/latest/USD", apiKey)
	return endpoint
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewEXClient(conf), version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
