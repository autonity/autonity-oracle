package common

import (
	"autonity-oracle/config"
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
	Version = "v0.2.6"
	apiPath = "api/v3/ticker/price"
	symbol  = "symbols"
)

type SIMClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewSIMClient(conf *config.PluginConfig) *SIMClient {
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

	for i := range prices {
		prices[i].Volume = types.DefaultVolume.String()
	}

	return prices, nil
}

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
