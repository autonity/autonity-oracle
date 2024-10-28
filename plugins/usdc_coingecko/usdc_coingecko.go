package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	version      = "v0.0.1"
	path         = "api/v3/simple/price"
	ids          = "ids"
	vsCurrencies = "vs_currencies"
	base         = "usd-coin"
	quote        = "usd"
)

var defaultConfig = types.PluginConfig{
	Name:               "usdc_coingecko",
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
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewCoinGeckoClient(conf *types.PluginConfig) *CoinGeckoClient {
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

func (c *CoinGeckoClient) FetchPrice(symbols []string) (common.Prices, error) {
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

	for _, s := range symbols {
		p, err := c.toPrice(s, &result)
		if err != nil {
			c.logger.Error("error filling USDC-USD price data", "err", err.Error())
			continue
		}
		prices = append(prices, p)
	}

	return prices, nil
}

func (c *CoinGeckoClient) toPrice(symbol string, res *Response) (common.Price, error) {
	var price common.Price
	sep := common.ResolveSeparator(symbol)
	codes := strings.Split(symbol, sep)
	if len(codes) != 2 {
		return price, fmt.Errorf("invalid symbol %s", symbol)
	}

	from := codes[0]
	to := codes[1]
	if to != "USD" {
		return price, fmt.Errorf("wrong base %s", to)
	}

	if from != "USDC" {
		return price, fmt.Errorf("wrong currency %s", from)
	}

	price.Symbol = symbol
	value := strconv.FormatFloat(res.USDCoin.USD, 'f', 6, 64)
	price.Price = value
	return price, nil
}

func (c *CoinGeckoClient) AvailableSymbols() ([]string, error) {
	return common.DefaultUSDCSymbols, nil
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
