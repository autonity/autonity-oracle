package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"io"
	"net/url"
	"os"
	"time"
)

const (
	version = "v0.0.2"
	apiPath = "api/v3/ticker/price"
	symbol  = "symbols"
)

var defaultConfig = types.PluginConfig{
	Key:                "",
	Scheme:             "https",
	Endpoint:           "api.binance.us",
	Timeout:            10, //10s
	DataUpdateInterval: 30, //10s
}

type BIClient struct {
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewBIClient(conf *types.PluginConfig) *BIClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	if client == nil {
		panic("cannot create https client for api.binance.us")
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &BIClient{conf: conf, client: client, logger: logger}
}

func (bi *BIClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices
	u, err := bi.buildURL(symbols)
	if err != nil {
		return nil, err
	}

	res, err := bi.client.Conn.Request(bi.conf.Scheme, u)
	if err != nil {
		bi.logger.Error("https get", "error", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		bi.logger.Error("io read", "error", err.Error())
		return nil, err
	}

	err = json.Unmarshal(body, &prices)
	if err != nil {
		return nil, err
	}

	return prices, nil
}

func (bi *BIClient) AvailableSymbols() ([]string, error) {
	var res []string
	prices, err := bi.FetchPrice(nil)
	if err != nil {
		return nil, err
	}

	for _, p := range prices {
		res = append(res, p.Symbol)
	}
	return res, nil
}

func (bi *BIClient) Close() {
	bi.client.Conn.Close()
}

func (bi *BIClient) buildURL(symbols []string) (*url.URL, error) {
	endpoint := &url.URL{}
	endpoint.Path = apiPath

	if len(symbols) != 0 {
		parameters, err := json.Marshal(symbols)
		if err != nil {
			return nil, err
		}

		query := endpoint.Query()
		query.Set(symbol, string(parameters))
		endpoint.RawQuery = query.Encode()
	}

	return endpoint, nil
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewBIClient(conf), version)
	defer adapter.Close()

	common.PluginServe(adapter)
}
