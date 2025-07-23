package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	version = "v0.2.4"
)

var defaultConfig = config.PluginConfig{
	Name:               "forex_wise",
	Key:                "0x123",
	Scheme:             "https",
	Endpoint:           "api.transferwise.com",
	Timeout:            10, // Timeout in seconds
	DataUpdateInterval: 30,
}

type WiseClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

type WRResult struct {
	Rate   decimal.Decimal `json:"rate"`
	Source string          `json:"source"`
	Target string          `json:"target"`
	Time   string          `json:"time"`
}

func (wc *WiseClient) buildURL(source, target string) *url.URL {
	endpoint := &url.URL{
		Scheme: "https",
		Host:   wc.conf.Endpoint,
		Path:   "/v1/rates",
	}

	query := endpoint.Query()
	query.Set("source", source)
	query.Set("target", target)

	endpoint.RawQuery = query.Encode()

	return endpoint
}

func NewWiseClient(conf *config.PluginConfig) *WiseClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &WiseClient{
		conf:   conf,
		client: client,
		logger: logger,
	}
}

func (wc *WiseClient) KeyRequired() bool {
	return true
}

// FetchPrice fetches forex prices for given symbols.
func (wc *WiseClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices

	for _, symbol := range symbols {
		parts := strings.Split(symbol, "-")
		if len(parts) != 2 {
			wc.logger.Warn("Invalid symbol format, expected SOURCE-TARGET", "symbol", symbol)
			continue
		}

		source := parts[0]
		target := parts[1]

		reqURL := wc.buildURL(target, source)

		req, err := http.NewRequest("GET", reqURL.String(), nil)
		if err != nil {
			wc.logger.Error("Failed to create request", "error", err)
			continue
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", wc.conf.Key))
		resp, err := wc.client.Conn.Do(req)
		if err != nil {
			wc.logger.Error("Request to Wise API failed", "error", err)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			wc.logger.Error("API response returned non-200 status code", "status", resp.Status, "symbol", symbol)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			wc.logger.Error("Failed to read response body", "error", err)
			continue
		}

		var result []WRResult
		if err := json.Unmarshal(body, &result); err != nil {
			wc.logger.Error("Failed to parse JSON response", "error", err)
			continue
		}

		for i := range result {

			p, err := wc.symbolsToPrice(symbol, &result[i])
			if err != nil {
				wc.logger.Error("symbol to price", "error", err.Error())
				continue
			}
			prices = append(prices, p)

		}
	}

	return prices, nil
}

func (wc *WiseClient) symbolsToPrice(s string, res *WRResult) (common.Price, error) {
	var price common.Price
	sep := common.ResolveSeparator(s)
	codes := strings.Split(s, sep)
	if len(codes) != 2 {
		return price, fmt.Errorf("invalid symbol %s", s)
	}

	from := codes[0]
	to := codes[1]
	if to != res.Source {
		return price, fmt.Errorf("wrong base %s", to)
	}

	price.Symbol = s
	price.Volume = types.DefaultVolume.String()
	switch from {
	case "EUR":
		price.Price = decimal.NewFromInt(1).Div(res.Rate).String()
	case "JPY":
		price.Price = decimal.NewFromInt(1).Div(res.Rate).String()
	case "GBP":
		price.Price = decimal.NewFromInt(1).Div(res.Rate).String()
	case "AUD":
		price.Price = decimal.NewFromInt(1).Div(res.Rate).String()
	case "CAD":
		price.Price = decimal.NewFromInt(1).Div(res.Rate).String()
	case "SEK":
		price.Price = decimal.NewFromInt(1).Div(res.Rate).String()
	default:
		return price, fmt.Errorf("unknown symbol %s", from)
	}
	return price, nil
}

// AvailableSymbols returns the supported symbols.
func (wc *WiseClient) AvailableSymbols() ([]string, error) {
	return common.DefaultForexSymbols, nil
}

func (wc *WiseClient) Close() {
	wc.client.Conn.Close()
}

func main() {
	// Plugin configuration

	conf := common.ResolveConf(os.Args[0], &defaultConfig)

	adapter := common.NewPlugin(conf, NewWiseClient(conf), version, types.SrcCEX, nil)
	defer adapter.Close()

	common.PluginServe(adapter)
}
