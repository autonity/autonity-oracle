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
	version   = "v0.1.0"
	pathLive  = "live"
	accessKey = "access_key"
)

var defaultConfig = types.PluginConfig{
	Name:               "forex_nodesphere",
	Key:                "sandbox",
	Scheme:             "https",
	Endpoint:           "api-currency.nodesphere.net",
	Timeout:            10, //10s
	DataUpdateInterval: 30, //30s
}

type NPResult struct {
	Success   bool   `json:"success"`
	Timestamp int64  `json:"timestamp"`
	Source    string `json:"source"`
	Quotes    Quotes `json:"quotes"`
}

type Quotes struct {
	EURUSD decimal.Decimal `json:"EURUSD"`
	JPYUSD decimal.Decimal `json:"JPYUSD"`
	GBPUSD decimal.Decimal `json:"GBPUSD"`
	AUDUSD decimal.Decimal `json:"AUDUSD"`
	CADUSD decimal.Decimal `json:"CADUSD"`
	SEKUSD decimal.Decimal `json:"SEKUSD"`
}

type NPClient struct {
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewNPClient(conf *types.PluginConfig) *NPClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &NPClient{conf: conf, client: client, logger: logger}
}

func (cl *NPClient) KeyRequired() bool {
	return true
}

func (cl *NPClient) FetchPrice(symbols []string) (common.Prices, error) {
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

	var result NPResult
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
func (cl *NPClient) AvailableSymbols() ([]string, error) {
	return common.DefaultForexSymbols, nil
}

func (cl *NPClient) Close() {
	cl.client.Conn.Close()
}

func (cl *NPClient) symbolsToPrice(s string, res *NPResult) (common.Price, error) {
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
		price.Price = res.Quotes.EURUSD.String()
	case "JPY":
		price.Price = res.Quotes.JPYUSD.String()
	case "GBP":
		price.Price = res.Quotes.GBPUSD.String()
	case "AUD":
		price.Price = res.Quotes.AUDUSD.String()
	case "CAD":
		price.Price = res.Quotes.CADUSD.String()
	case "SEK":
		price.Price = res.Quotes.SEKUSD.String()
	default:
		return price, fmt.Errorf("unknown symbol %s", from)
	}
	return price, nil
}

func (cl *NPClient) buildURL(apiKey string) *url.URL {
	endpoint := &url.URL{}
	endpoint.Path = pathLive

	query := endpoint.Query()
	query.Set(accessKey, apiKey)

	endpoint.RawQuery = query.Encode()
	return endpoint
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewNPClient(conf), version, types.SrcCEX, nil)
	defer adapter.Close()
	common.PluginServe(adapter)
}
