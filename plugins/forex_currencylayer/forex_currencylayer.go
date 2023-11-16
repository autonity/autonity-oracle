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
	version   = "v0.0.1"
	pathLive  = "live"
	accessKey = "access_key"
)

var defaultConfig = types.PluginConfig{
	Key:                "705af082ac7f7d150c87303d4e2f049e",
	Scheme:             "http", // todo: replace the scheme with https once we purchase the service plan.
	Endpoint:           "api.currencylayer.com",
	Timeout:            10,   //10s
	DataUpdateInterval: 3600, //3600s
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
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewCLClient(conf *types.PluginConfig) *CLClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	if client == nil {
		panic("cannot create client for currency layer")
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	return &CLClient{conf: conf, client: client, logger: logger}
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

	if result.Timestamp == 0 {
		cl.logger.Error("data source returns", "data", string(body))
		return nil, common.ErrDataNotAvailable
	}

	if !result.Success {
		cl.logger.Error("fetch price", "error", "not success")
		return nil, fmt.Errorf("source return not success")
	}

	for _, s := range symbols {
		p, err := cl.symbolsToPrice(s, &result)
		if err != nil {
			cl.logger.Error("symbol to prices", "error", err.Error())
			continue
		}
		prices = append(prices, p)
	}

	cl.logger.Info("currency layer", "data", prices)
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
	conf, err := common.LoadPluginConf(os.Args[0])
	if err != nil {
		println("cannot load conf: ", err.Error(), os.Args[0])
		os.Exit(-1)
	}

	common.ResolveConf(&conf, &defaultConfig)

	client := NewCLClient(&conf)
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
