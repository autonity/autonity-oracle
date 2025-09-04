package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/hashicorp/go-hclog"
)

const (
	version = "v0.2.7"
	path    = "v2/prices/USDC-USD/spot"
)

var defaultConfig = config.PluginConfig{
	Name:               "crypto_coinbase",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "api.coinbase.com",
	Timeout:            10, // 10s
	DataUpdateInterval: 30, // 30s, tested and passed the rate limit policy of public data service of coinbase.
}

type PriceData struct {
	Amount   string `json:"amount"`
	Base     string `json:"base"`
	Currency string `json:"currency"`
}

type Response struct {
	Data PriceData `json:"data"`
}

type CoinBaseClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewCoinBaseClient(conf *config.PluginConfig) *CoinBaseClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &CoinBaseClient{conf: conf, client: client, logger: logger}
}

func (c *CoinBaseClient) KeyRequired() bool {
	return false
}

func (c *CoinBaseClient) FetchPrice(_ []string) (common.Prices, error) {
	var prices common.Prices
	u := c.buildURL()
	res, err := c.client.Conn.Request(c.conf.Scheme, u)
	if err != nil {
		c.logger.Error("Error fetching USDC-USD price data", "err", err.Error())
		return nil, err
	}

	defer res.Body.Close()
	if err = common.CheckHTTPStatusCode(res.StatusCode); err != nil {
		c.logger.Error("data source return error", "error", err.Error())
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.logger.Error("Error reading USDC-USD price data", "err", err.Error())
		return nil, err
	}

	var data Response
	err = json.Unmarshal(body, &data)
	if err != nil {
		c.logger.Error("unable to parse USDC-USD price data", "err", err.Error())
		return nil, err
	}

	prices = append(prices, common.Price{
		Symbol: common.DefaultUSDCSymbol,
		Price:  data.Data.Amount,
		Volume: types.DefaultVolume.String(),
	})

	return prices, nil
}

func (c *CoinBaseClient) AvailableSymbols() ([]string, error) {
	return []string{common.DefaultUSDCSymbol}, nil
}

func (c *CoinBaseClient) Close() {
	c.client.Conn.Close()
}

func (c *CoinBaseClient) buildURL() *url.URL {
	endpoint := &url.URL{}
	endpoint.Path = path
	query := endpoint.Query()
	endpoint.RawQuery = query.Encode()
	return endpoint
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewCoinBaseClient(conf), version, types.SrcCEX, nil)
	defer adapter.Close()
	common.PluginServe(adapter)
}
