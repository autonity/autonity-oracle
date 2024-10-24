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
	"strings"
	"time"
)

const (
	version = "v0.0.1"
	path    = "v2/prices/USDC-USD/spot"
)

var defaultConfig = types.PluginConfig{
	Name:               "usdc_coinbase",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "api.coinbase.com",
	Timeout:            10,  // 10s
	DataUpdateInterval: 120, // 120s
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
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewCoinBaseClient(conf *types.PluginConfig) *CoinBaseClient {
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

func (c *CoinBaseClient) FetchPrice(symbols []string) (common.Prices, error) {
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

	for _, s := range symbols {
		p, err := c.toPrice(s, &data)
		if err != nil {
			c.logger.Error("error filling USDC-USD price data", "err", err.Error())
			continue
		}
		prices = append(prices, p)
	}

	return prices, nil
}

func (c *CoinBaseClient) toPrice(symbol string, res *Response) (common.Price, error) {
	var price common.Price
	sep := common.ResolveSeparator(symbol)
	codes := strings.Split(symbol, sep)
	if len(codes) != 2 {
		return price, fmt.Errorf("invalid symbol %s", symbol)
	}

	from := codes[0]
	to := codes[1]
	if to != res.Data.Currency {
		return price, fmt.Errorf("wrong base %s", to)
	}

	if from != res.Data.Base {
		return price, fmt.Errorf("wrong currency %s", from)
	}

	price.Symbol = symbol
	price.Price = res.Data.Amount
	return price, nil
}

func (c *CoinBaseClient) AvailableSymbols() ([]string, error) {
	return common.DefaultUSDCSymbols, nil
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
	adapter := common.NewPlugin(conf, NewCoinBaseClient(conf), version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
