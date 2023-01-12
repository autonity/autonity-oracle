package main

import (
	"autonity-oracle/types"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
	"time"
)

// FakePlugin Here is an implementation of a fake plugin which returns simulated data points.
type FakePlugin struct {
	logger hclog.Logger
}

func (g *FakePlugin) FetchPrices(symbols []string) []types.Price {
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
	return prices
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Trace,
		Output:     os.Stderr,
		JSONFormat: true,
	})

	adapter := &FakePlugin{
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
