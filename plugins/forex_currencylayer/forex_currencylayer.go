package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
)

const (
	version   = "v0.2.7"
	pathLive  = "live"
	accessKey = "access_key"
)

var defaultConfig = config.PluginConfig{
	Name:               "forex_currencylayer",
	Key:                "",
	Scheme:             "http",
	Endpoint:           "api.currencylayer.com",
	Timeout:            10, //10s
	DataUpdateInterval: 30, //30s
}

type CLResult struct {
	Success   bool   `json:"success"`
	Terms     string `json:"terms"`
	Privacy   string `json:"privacy"`
	Timestamp int64  `json:"timestamp"`
	Source    string `json:"source"`
	Quotes    Quotes `json:"quotes"`
}

type Quotes struct {
	USDEUR decimal.Decimal `json:"USDEUR"`
	USDJPY decimal.Decimal `json:"USDJPY"`
	USDGBP decimal.Decimal `json:"USDGBP"`
	USDAUD decimal.Decimal `json:"USDAUD"`
	USDCAD decimal.Decimal `json:"USDCAD"`
	USDSEK decimal.Decimal `json:"USDSEK"`
}

type CLClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewCLClient(conf *config.PluginConfig) *CLClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &CLClient{conf: conf, client: client, logger: logger}
}

func (cl *CLClient) KeyRequired() bool {
	return true
}

func (cl *CLClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices
	u := cl.buildURL(cl.conf.Key)

	res, err := cl.client.Conn.Request(cl.conf.Scheme, u)
	if err != nil {
		cl.logger.Error("http request", "error", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	if err = common.CheckHTTPStatusCode(res.StatusCode); err != nil {
		cl.logger.Error("data source return error", "error", err.Error())
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		cl.logger.Error("io read", "error", err.Error())
		return nil, err
	}

	var result CLResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		cl.logger.Error("unmarshal price", "error", err.Error())
		return nil, err
	}

	if !result.Success {
		cl.logger.Error("fetch price", "error", string(body))
		return nil, fmt.Errorf("data source return error: %s", string(body))
	}

	for _, s := range symbols {
		p, err := cl.symbolsToPrice(s, &result)
		if err != nil {
			cl.logger.Error("symbol to prices", "error", err.Error())
			continue
		}
		prices = append(prices, p)
	}

	return prices, nil
}

// AvailableSymbols returns the adapted symbols for current data source.
func (cl *CLClient) AvailableSymbols() ([]string, error) {
	return common.DefaultForexSymbols, nil
}

func (cl *CLClient) Close() {
	cl.client.Conn.Close()
}

func (cl *CLClient) symbolsToPrice(s string, res *CLResult) (common.Price, error) {
	var price common.Price
	sep := common.ResolveSeparator(s)
	codes := strings.Split(s, sep)
	if len(codes) != 2 {
		return price, fmt.Errorf("invalid symbol %s", s)
	}

	from := codes[0]
	to := codes[1]
	if to != res.Source {
		return price, fmt.Errorf("wrong base %s", to)
	}

	price.Symbol = s
	price.Volume = types.DefaultVolume.String()
	switch from {
	case "EUR":
		price.Price = decimal.NewFromInt(1).Div(res.Quotes.USDEUR).String()
	case "JPY":
		price.Price = decimal.NewFromInt(1).Div(res.Quotes.USDJPY).String()
	case "GBP":
		price.Price = decimal.NewFromInt(1).Div(res.Quotes.USDGBP).String()
	case "AUD":
		price.Price = decimal.NewFromInt(1).Div(res.Quotes.USDAUD).String()
	case "CAD":
		price.Price = decimal.NewFromInt(1).Div(res.Quotes.USDCAD).String()
	case "SEK":
		price.Price = decimal.NewFromInt(1).Div(res.Quotes.USDSEK).String()
	default:
		return price, fmt.Errorf("unknown symbol %s", from)
	}
	return price, nil
}

func (cl *CLClient) buildURL(apiKey string) *url.URL {
	endpoint := &url.URL{}
	endpoint.Path = pathLive

	query := endpoint.Query()
	query.Set(accessKey, apiKey)

	endpoint.RawQuery = query.Encode()
	return endpoint
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewCLClient(conf), version, types.SrcCEX, nil)
	defer adapter.Close()
	common.PluginServe(adapter)
}
