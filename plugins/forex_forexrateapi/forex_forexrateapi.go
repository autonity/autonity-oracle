package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
)

const (
	version = "v0.2.6"
	path    = "/v1/latest"
)

var defaultConfig = config.PluginConfig{
	Name:               "forex_rateapi",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "api.forexrateapi.com",
	Timeout:            10,  // 10 seconds
	DataUpdateInterval: 300, // 5 minutes
}

type ForexRateAPIClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewForexRateAPIClient(conf *config.PluginConfig) *ForexRateAPIClient {
	client := common.NewClient("", time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	if client == nil {
		panic("cannot create common client")
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:  conf.Name,
		Level: hclog.Info,
	})

	return &ForexRateAPIClient{conf: conf, client: client, logger: logger}
}

func (c *ForexRateAPIClient) KeyRequired() bool {
	return true
}

func (c *ForexRateAPIClient) FetchPrice(symbols []string) (common.Prices, error) {
	var allPrices common.Prices

	requests := c.groupSymbolsByBase(symbols)
	c.logger.Debug("Grouped symbol requests", "requests", requests)

	for base, symbolsToFetch := range requests {
		u, err := c.buildURL(base, symbolsToFetch)
		if err != nil {
			c.logger.Error("Failed to build URL", "base", base, "error", err)
			continue
		}

		res, err := c.client.Conn.Request(c.conf.Scheme, u)
		if err != nil {
			c.logger.Error("HTTP request failed", "base", base, "error", err)
			continue
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			c.logger.Error("Failed to read response body", "base", base, "error", err)
			continue
		}

		if res.StatusCode != 200 {
			c.logger.Error("API request failed", "status", res.StatusCode, "body", string(body))
			continue
		}

		type ForexLatestResponse struct {
			Success bool               `json:"success"`
			Base    string             `json:"base"`
			Rates   map[string]float64 `json:"rates"`
			Error   *struct {
				Info string `json:"info"`
			} `json:"error"`
		}

		var apiResponse ForexLatestResponse
		if err := json.Unmarshal(body, &apiResponse); err != nil {
			c.logger.Error("Failed to unmarshal JSON response", "base", base, "error", err)
			continue
		}

		if !apiResponse.Success && apiResponse.Error != nil {
			c.logger.Error("API returned an error", "base", base, "message", apiResponse.Error.Info)
			continue
		}

		for quoteCurrency, rate := range apiResponse.Rates {
			if rate == 0 {
				c.logger.Warn("Received zero rate for symbol, skipping", "symbol", quoteCurrency)
				continue
			}

			invertedRate := decimal.NewFromInt(1).Div(decimal.NewFromFloat(rate))
			originalSymbol := fmt.Sprintf("%s-%s", quoteCurrency, base)

			allPrices = append(allPrices, common.Price{
				Symbol: originalSymbol,
				Price:  invertedRate.StringFixed(8),
				Volume: types.DefaultVolume.String(),
			})
		}
	}

	return allPrices, nil
}

func (c *ForexRateAPIClient) AvailableSymbols() ([]string, error) {
	return common.DefaultForexSymbols, nil
}

func (c *ForexRateAPIClient) Close() {
	if c.client != nil && c.client.Conn != nil {
		c.client.Conn.Close()
	}
}

func (c *ForexRateAPIClient) groupSymbolsByBase(symbols []string) map[string][]string {
	requests := make(map[string][]string)
	for _, symbol := range symbols {
		parts := strings.Split(symbol, "-")
		if len(parts) != 2 {
			continue
		}
		quote := parts[0]
		base := parts[1]
		requests[base] = append(requests[base], quote)
	}
	return requests
}

func (c *ForexRateAPIClient) buildURL(base string, symbols []string) (*url.URL, error) {
	endpoint := &url.URL{
		Scheme: c.conf.Scheme,
		Host:   c.conf.Endpoint,
		Path:   path,
	}

	query := endpoint.Query()
	query.Set("api_key", c.conf.Key)
	query.Set("base", base)
	query.Set("currencies", strings.Join(symbols, ","))

	endpoint.RawQuery = query.Encode()
	return endpoint, nil
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewForexRateAPIClient(conf), version, types.SrcCEX, nil)
	defer adapter.Close()
	common.PluginServe(adapter)
}
