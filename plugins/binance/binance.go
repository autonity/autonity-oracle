package main

import (
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shopspring/decimal"
	"net/http"
	"os"
	"time"
)

var version = "v0.0.1"

const BinanceMarketDataURL = "https://api.binance.com/api/v3/price"

// Price is the basic data structure returned by Binance.
type Price struct {
	Symbol string `json:"symbol,omitempty"`
	Price  string `json:"price,omitempty"`
}

type Prices []Price

type BadRequest struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

// Binance Here is an implementation of a fake plugin which returns simulated data points.
type Binance struct {
	logger hclog.Logger
	client *http.Client
}

func (g *Binance) FetchPrices(symbols []string) ([]types.Price, error) {
	g.logger.Debug("fetching price for symbols: ", symbols)
	parameters, err := json.Marshal(symbols)
	if err != nil {
		return nil, err
	}

	if g.client == nil {
		g.client = &http.Client{}
	}

	req, err := http.NewRequest(http.MethodGet, BinanceMarketDataURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	// appending to existing query args
	q := req.URL.Query()
	q.Add("symbols", string(parameters))
	// assign encoded query string to http request.
	req.URL.RawQuery = q.Encode()

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}

	g.logger.Debug("Get HTTP response status code: ", resp.StatusCode)

	if resp.StatusCode == http.StatusUnavailableForLegalReasons {
		// https://dev.binance.vision/t/api-error-451-unavailable-for-legal-reasons/13828/4
		return nil, fmt.Errorf("StatusUnavailableForLegalReasons")
	}

	if resp.StatusCode == http.StatusBadRequest {
		var badReq BadRequest
		err = json.NewDecoder(resp.Body).Decode(&badReq)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("BadRequest: %s", badReq.Msg)
	}

	if resp.StatusCode == http.StatusOK {
		var prices Prices
		err = json.NewDecoder(resp.Body).Decode(&prices)
		if err != nil {
			return nil, err
		}
		g.logger.Debug("data points: ", prices)

		var results []types.Price
		now := time.Now().UnixMilli()
		for _, v := range prices {
			dec, err := decimal.NewFromString(v.Price)
			if err != nil {
				g.logger.Error("cannot convert price string to decimal: ", v.Price, err)
				continue
			}
			results = append(results, types.Price{
				Timestamp: now,
				Symbol:    v.Symbol,
				Price:     dec,
			})
		}

		return results, nil
	}
	return nil, nil
}

func (g *Binance) GetVersion() (string, error) {
	return version, nil
}

func (g *Binance) Close() {
	if g.client != nil {
		g.client.CloseIdleConnections()
	}
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Debug,
		Output:     os.Stderr, // logging to stderr thus the framework can redirect the logs from plugin to plugin server.
		JSONFormat: true,
	})

	adapter := &Binance{
		logger: logger,
	}
	defer adapter.Close()
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{Impl: adapter},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
