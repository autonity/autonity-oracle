package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	version      = "v0.2.0"
	path         = "api/v3/simple/price"
	ids          = "ids"
	vsCurrencies = "vs_currencies"
	base         = "usd-coin"
	quote        = "usd"
)

var defaultConfig = config.PluginConfig{
	Name:               "crypto_coingecko",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "api.coingecko.com",
	Timeout:            10, // 10s
	DataUpdateInterval: 30, // 30s, tested and passed the rate limit policy of public data service of coin-gecko.
}

type CoinData struct {
	USD float64 `json:"usd"`
}

type Response struct {
	USDCoin CoinData `json:"usd-coin"`
}

type CoinGeckoClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewCoinGeckoClient(conf *config.PluginConfig) *CoinGeckoClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &CoinGeckoClient{conf: conf, client: client, logger: logger}
}

func (c *CoinGeckoClient) KeyRequired() bool {
	return false
}

func (c *CoinGeckoClient) FetchPrice(_ []string) (common.Prices, error) {
	var prices common.Prices
	u := c.buildURL()
	res, err := c.client.Conn.Request(c.conf.Scheme, u)
	if err != nil {
		c.logger.Error("https request", "error", err.Error())
		return nil, err
	}
	defer res.Body.Close()
	if err = common.CheckHTTPStatusCode(res.StatusCode); err != nil {
		c.logger.Error("data source return error", "error", err.Error())
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.logger.Error("io read", "error", err.Error())
		return nil, err
	}

	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		c.logger.Error("unmarshal result", "error", err.Error())
		return nil, err
	}

	prices = append(prices, common.Price{
		Symbol: common.DefaultUSDCSymbol,
		Price:  strconv.FormatFloat(result.USDCoin.USD, 'f', 6, 64),
	})

	return prices, nil
}

func (c *CoinGeckoClient) AvailableSymbols() ([]string, error) {
	return []string{common.DefaultUSDCSymbol}, nil
}

func (c *CoinGeckoClient) Close() {
	c.client.Conn.Close()
}

func (c *CoinGeckoClient) buildURL() *url.URL {
	endpoint := &url.URL{}
	endpoint.Path = path
	query := endpoint.Query()
	query.Set(ids, base)
	query.Set(vsCurrencies, quote)
	endpoint.RawQuery = query.Encode()
	return endpoint
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewCoinGeckoClient(conf), version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
