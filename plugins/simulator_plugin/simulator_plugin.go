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

// todo: refine the plugins code structures, some of them share duplicated code.

var version = "v0.0.1"

const SimulatorDataURL = "http://127.0.0.1:50991/api/v3/ticker/price"
const UnknownErr = "Unknown Error"
const DataLegalErr = "StatusUnavailableForLegalReasons"

var (
	FetchCounter         uint64
	SyncSymbolsInterval  = uint64(6) // on every 6 fetches, a symbol synchronization will be triggered.
	LatestBinanceSymbols = make(map[string]types.Price)
	Timeout              = time.Second * 5
)

// Price is the basic data structure returned by SimulatorPlugin.
type Price struct {
	Symbol string `json:"symbol,omitempty"`
	Price  string `json:"price,omitempty"`
}

type Prices []Price

// SimulatorPlugin Here is an implementation of a fake plugin which returns simulated data points.
type SimulatorPlugin struct {
	logger hclog.Logger
	client *http.Client
}

func (g *SimulatorPlugin) FetchPrices(symbols []string) (types.PluginPriceReport, error) {
	var report types.PluginPriceReport

	if FetchCounter%SyncSymbolsInterval == 0 {
		FetchCounter++
		return g.FetchPricesWithSymbolSync(symbols)
	}
	FetchCounter++
	goodSym, badSym := resolveSymbols(symbols)
	parameters, err := json.Marshal(goodSym)
	if err != nil {
		g.logger.Error("json marshal symbols", "error", err.Error())
		return report, err
	}

	if g.client == nil {
		g.client = &http.Client{}
		g.client.Timeout = Timeout
	}

	req, err := http.NewRequest(http.MethodGet, SimulatorDataURL, nil)
	if err != nil {
		g.logger.Error("http new request", "error", err.Error())
		return report, err
	}

	req.Header.Set("accept", "application/json")
	// appending to existing query args
	q := req.URL.Query()
	q.Add("symbols", string(parameters))
	// assign encoded query string to http request.
	req.URL.RawQuery = q.Encode()

	resp, err := g.client.Do(req)
	if err != nil {
		g.logger.Error("send http request", "error", err.Error())
		return report, err
	}

	//g.logger.Debug("Get HTTP response status code: ", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		msg := UnknownErr
		if resp.StatusCode == http.StatusUnavailableForLegalReasons {
			msg = "StatusUnavailableForLegalReasons"
		}
		if resp.StatusCode == http.StatusBadRequest {
			return g.FetchPricesWithSymbolSync(symbols)
		}
		return report, fmt.Errorf("ErrorCode: %d, msg: %s", resp.StatusCode, msg)
	}

	var prices Prices
	err = json.NewDecoder(resp.Body).Decode(&prices)
	if err != nil {
		g.logger.Error("decode http body", "err", err.Error())
		return report, err
	}
	//g.logger.Debug("sampled data points: ", prices)

	now := time.Now().Unix()
	for _, v := range prices {
		dec, err := decimal.NewFromString(v.Price)
		if err != nil {
			g.logger.Error("cannot convert price string to decimal: ", v.Price, err)
			continue
		}
		report.Prices = append(report.Prices, types.Price{
			Timestamp: now,
			Symbol:    v.Symbol,
			Price:     dec,
		})
	}
	report.BadSymbols = badSym

	return report, nil
}

func resolveSymbols(symbols []string) ([]string, []string) {
	var badSymbols []string
	var goodSymbols []string
	for _, s := range symbols {
		_, ok := LatestBinanceSymbols[s]
		if !ok {
			badSymbols = append(badSymbols, s)
		} else {
			goodSymbols = append(goodSymbols, s)
		}
	}
	return goodSymbols, badSymbols
}

// FetchPricesWithSymbolSync fetch all prices of supported symbols from binance, and filter out invalid symbols.
func (g *SimulatorPlugin) FetchPricesWithSymbolSync(symbols []string) (report types.PluginPriceReport, e error) {
	if g.client == nil {
		g.client = &http.Client{}
		g.client.Timeout = Timeout
	}

	// without specifying the query parameter, binance will return all its symbols' price.
	req, err := http.NewRequest(http.MethodGet, SimulatorDataURL, nil)
	if err != nil {
		g.logger.Error("http new request", "error", err.Error())
		return report, err
	}
	req.Header.Set("accept", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		g.logger.Error("send http request", "error", err.Error())
		return report, err
	}

	//g.logger.Debug("Get HTTP response status code: ", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		msg := "unknown error"
		if resp.StatusCode == http.StatusUnavailableForLegalReasons {
			msg = DataLegalErr
		}
		return report, fmt.Errorf("ErrorCode: %d, msg: %s", resp.StatusCode, msg)
	}

	var prices Prices
	err = json.NewDecoder(resp.Body).Decode(&prices)
	if err != nil {
		g.logger.Error("decode http body", "error", err.Error())
		return report, err
	}

	now := time.Now().Unix()
	LatestBinanceSymbols = make(map[string]types.Price)
	for _, v := range prices {
		dec, err := decimal.NewFromString(v.Price)
		if err != nil {
			g.logger.Error("cannot convert price string to decimal: ", v.Price, err)
			continue
		}
		LatestBinanceSymbols[v.Symbol] = types.Price{
			Timestamp: now,
			Symbol:    v.Symbol,
			Price:     dec,
		}
	}

	for _, s := range symbols {
		p, ok := LatestBinanceSymbols[s]
		if !ok {
			report.BadSymbols = append(report.BadSymbols, s)
		} else {
			report.Prices = append(report.Prices, p)
		}
	}

	return report, nil
}

func (g *SimulatorPlugin) GetVersion() (string, error) {
	return version, nil
}

func (g *SimulatorPlugin) Close() {
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

	adapter := &SimulatorPlugin{
		logger: logger,
	}

	_, err := adapter.FetchPricesWithSymbolSync(nil)
	if err != nil {
		logger.Error("Init symbols failed: ", err)
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
