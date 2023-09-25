package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"io"
	"net/url"
	"os"
	"time"
)

const (
	version               = "v0.0.2"
	defaultScheme         = "https"
	defaultHost           = "simfeed.bakerloo.autonity.org"
	apiPath               = "api/v3/ticker/price"
	symbol                = "symbols"
	defaultTimeout        = 10 // 10s
	defaultUpdateInterval = 10 // 10s
	defaultKey            = ""
)

type R4CESClient struct {
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewR4CESClient(conf *types.PluginConfig) *R4CESClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	if client == nil {
		panic("cannot create client for exchange rate api")
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	return &R4CESClient{conf: conf, client: client, logger: logger}
}

func (bi *R4CESClient) FetchPrice(symbols []string) (common.Prices, error) {
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

	bi.logger.Info("binance", "data", prices)
	return prices, nil
}

func (bi *R4CESClient) AvailableSymbols() ([]string, error) {
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

func (bi *R4CESClient) Close() {
	bi.client.Conn.Close()
}

func (bi *R4CESClient) buildURL(symbols []string) (*url.URL, error) {
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

func resolveConf(conf *types.PluginConfig) {

	if conf.Timeout == 0 {
		conf.Timeout = defaultTimeout
	}

	if conf.DataUpdateInterval == 0 {
		conf.DataUpdateInterval = defaultUpdateInterval
	}

	if len(conf.Scheme) == 0 {
		conf.Scheme = defaultScheme
	}

	if len(conf.Endpoint) == 0 {
		conf.Endpoint = defaultHost
	}

	if len(conf.Key) == 0 {
		conf.Key = defaultKey
	}
}

func main() {
	conf, err := common.LoadPluginConf(os.Args[0])
	if err != nil {
		println("cannot load conf: ", err.Error(), os.Args[0])
		os.Exit(-1)
	}

	resolveConf(&conf)

	client := NewR4CESClient(&conf)
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
