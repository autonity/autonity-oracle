package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
)

const (
	version = "v0.2.7"
	path    = "/v6/finance/quote"
)

var defaultConfig = config.PluginConfig{
	Name:               "forex_yahoofinance",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "yfapi.net",
	Timeout:            10,  // Timeout in seconds
	DataUpdateInterval: 600, // 10 minutes for yahoo finance free data plan.
}

type YahooFinanceClient struct {
	conf   *config.PluginConfig
	client *common.Client
	logger hclog.Logger
}

type YahooResponse struct {
	QuoteResponse struct {
		Result []struct {
			Symbol string  `json:"symbol"`
			Price  float64 `json:"regularMarketPrice"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"quoteResponse"`
}

func NewYahooClient(conf *config.PluginConfig) *YahooFinanceClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &YahooFinanceClient{
		conf:   conf,
		client: client,
		logger: logger,
	}
}

func (yh *YahooFinanceClient) KeyRequired() bool {
	return true
}

func (yh *YahooFinanceClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices

	var symbolMap = make(map[string]string)
	convSymbols := make([]string, len(symbols))
	for i, symbol := range symbols {
		if _, ok := common.ForexCurrencies[symbol]; ok {
			yhSym := strings.Join(strings.Split(symbol, "-"), "")
			convSymbols[i] = yhSym + "=X"
			symbolMap[convSymbols[i]] = symbol
			continue
		}
		symbolMap[symbol] = symbol
		convSymbols[i] = symbol
	}

	pairParams := strings.Join(convSymbols, ",")

	reqURL := yh.buildURL(pairParams)
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		yh.logger.Error("Error building request", "error", err)
		return prices, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", yh.conf.Key)
	resp, err := yh.client.Conn.Do(req)
	if err != nil {
		yh.logger.Error("Error making request", "error", err)
		return prices, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		yh.logger.Error("Error making request", "status", resp.Status)
		return prices, err
	}

	var response YahooResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		yh.logger.Error("Error parsing JSON response", "error", err)
		return prices, err
	}

	if err = json.Unmarshal(body, &response); err != nil {
		yh.logger.Error("cannot unmarshall JSON", "error", err)
		return prices, err
	}

	if response.QuoteResponse.Error != nil {
		yh.logger.Error("Error parsing JSON response", "error", response.QuoteResponse.Error)
		return prices, fmt.Errorf("%v", response.QuoteResponse.Error)
	}

	for _, result := range response.QuoteResponse.Result {
		srcSymbol := symbolMap[result.Symbol]
		price := result.Price

		prices = append(prices, common.Price{
			Symbol: srcSymbol,
			Price:  fmt.Sprintf("%.8f", price),
			Volume: types.DefaultVolume.String(),
		})
	}

	return prices, nil
}

func (yh *YahooFinanceClient) AvailableSymbols() ([]string, error) {
	return append(common.DefaultForexSymbols, common.DefaultUSDCSymbol), nil
}

func (yh *YahooFinanceClient) Close() {
	yh.client.Conn.Close()
}

func (yh *YahooFinanceClient) buildURL(symbols string) *url.URL {
	endpoint := &url.URL{
		Scheme: yh.conf.Scheme,
		Host:   yh.conf.Endpoint,
		Path:   path,
	}

	query := endpoint.Query()
	query.Set("symbols", symbols)
	endpoint.RawQuery = query.Encode()
	return endpoint
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewYahooClient(conf), version, types.SrcCEX, nil)
	defer adapter.Close()
	common.PluginServe(adapter)
}
