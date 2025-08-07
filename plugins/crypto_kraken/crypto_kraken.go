package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"io"
	"net/url"
	"os"
	"time"
)

const (
	version         = "v0.2.5"
	path            = "0/public/Ticker"
	queryParam      = "pair"
	supportedSymbol = "USDCUSD"
)

var defaultConfig = config.PluginConfig{
	Name:               "crypto_kraken",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "api.kraken.com",
	Timeout:            10, // 10s
	DataUpdateInterval: 30, // 30s, tested and passed the rate limit policy of public data service of kraken.
}

type Result struct {
	A []string `json:"a"` // ask price [price, whole lot volume, lot volume]
	B []string `json:"b"` // bid price [price, whole lot volume, lot volume]
	C []string `json:"c"` // last trade closed [price, lot volume]
	V []string `json:"v"` // volume [today, last 24 hours]
	P []string `json:"p"` // volume weighted average price [today, last 24 hours]
	T []int64  `json:"t"` // num of trades [today, last 24 hours]
	L []string `json:"l"` // low [today, last 24 hours]
	H []string `json:"h"` // high [today, last 24 hours]
	O string   `json:"o"` // today's opening price
}

type Response struct {
	Error  []string          `json:"error"`
	Result map[string]Result `json:"result"`
}

type KrakenClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewKrakenClient(conf *config.PluginConfig) *KrakenClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &KrakenClient{conf: conf, client: client, logger: logger}
}

func (k *KrakenClient) KeyRequired() bool {
	return false
}

func (k *KrakenClient) FetchPrice(_ []string) (common.Prices, error) {
	var prices common.Prices
	u := k.buildURL()
	res, err := k.client.Conn.Request(k.conf.Scheme, u)
	if err != nil {
		k.logger.Error("https request", "error", err.Error())
		return nil, err
	}
	defer res.Body.Close()
	if err = common.CheckHTTPStatusCode(res.StatusCode); err != nil {
		k.logger.Error("data source return error", "error", err.Error())
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		k.logger.Error("io read", "error", err.Error())
		return nil, err
	}

	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		k.logger.Error("unmarshal result", "error", err.Error())
		return nil, err
	}

	if len(result.Error) > 0 {
		k.logger.Error("data source return error", "error", result.Error)
		return nil, fmt.Errorf("%s", result.Error[0])
	}

	p, err := k.toPrice(common.DefaultUSDCSymbol, &result)
	if err != nil {
		k.logger.Error("error filling USDC-USD price data", "err", err.Error())
		return nil, err
	}

	prices = append(prices, p)
	return prices, nil
}

func (k *KrakenClient) toPrice(symbol string, res *Response) (common.Price, error) {
	var price common.Price

	usdcResult, ok := res.Result[supportedSymbol]
	if !ok {
		return price, fmt.Errorf("symbol %s not found", symbol)
	}

	if len(usdcResult.P) == 0 {
		return price, fmt.Errorf("%s price not found", symbol)
	}

	price.Symbol = symbol
	price.Price = usdcResult.P[0] // take the volume weighted average price of today.
	price.Volume = types.DefaultVolume.String()
	return price, nil
}

func (k *KrakenClient) AvailableSymbols() ([]string, error) {
	return []string{common.DefaultUSDCSymbol}, nil
}

func (k *KrakenClient) Close() {
	k.client.Conn.Close()
}

func (k *KrakenClient) buildURL() *url.URL {
	endpoint := &url.URL{}
	endpoint.Path = path
	query := endpoint.Query()
	query.Set(queryParam, supportedSymbol)
	endpoint.RawQuery = query.Encode()
	return endpoint
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewKrakenClient(conf), version, types.SrcCEX, nil)
	defer adapter.Close()
	common.PluginServe(adapter)
}
