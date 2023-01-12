package main

import (
	"autonity-oracle/types"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
	"time"
)

var version = "v0.0.1"

// Binance Here is an implementation of a fake plugin which returns simulated data points.
type Binance struct {
	logger hclog.Logger
}

func (g *Binance) FetchPrices(symbols []string) ([]types.Price, error) {
	// todo: fetch prices for symbols from binance http endpoint, for the time being, we just simulate fake data.
	g.logger.Debug("receive request from oracle service, send data response.")
	var prices []types.Price
	for _, s := range symbols {
		p := types.Price{
			Timestamp: time.Now().UnixMilli(),
			Symbol:    s,
			Price:     types.SimulatedPrice,
		}
		prices = append(prices, p)
	}

	g.logger.Debug("", prices)
	return prices, nil
}

func (g *Binance) GetVersion() (string, error) {
	return version, nil
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	adapter := &Binance{
		logger: logger,
	}
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{Impl: adapter},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
