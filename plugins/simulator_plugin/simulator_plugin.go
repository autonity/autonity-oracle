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
	version = "v0.2.0"
	apiPath = "api/v3/ticker/price"
	symbol  = "symbols"
)

// set to bakerloo feed service by default.
var defaultEndpoint = "simfeed.bakerloo.autonity.org"

var defaultConfig = types.PluginConfig{
	Name:               "simulator_plugin",
	Key:                "",
	Scheme:             "https",
	Endpoint:           defaultEndpoint,
	Timeout:            10, //10s
	DataUpdateInterval: 10, //10s
}

type SIMClient struct {
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewSIMClient(conf *types.PluginConfig) *SIMClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &SIMClient{conf: conf, client: client, logger: logger}
}

func (bi *SIMClient) KeyRequired() bool {
	return false
}

func (bi *SIMClient) FetchPrice(symbols []string) (common.Prices, error) {
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

	if err = common.CheckHTTPStatusCode(res.StatusCode); err != nil {
		return nil, err
	}

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

// AvailableSymbols get available symbols from simulator service.
// It could return: ATN-USDC, NTN-USDC, NTN-ATN, ATN-USDX, NTN-USDX.
func (bi *SIMClient) AvailableSymbols() ([]string, error) {
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

func (bi *SIMClient) Close() {
	bi.client.Conn.Close()
}

func (bi *SIMClient) buildURL(symbols []string) (*url.URL, error) {
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
	adapter := common.NewPlugin(conf, NewSIMClient(conf), version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
