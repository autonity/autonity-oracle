package main

import (
	"autonity-oracle/types"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/modern-go/reflect2"
	"os"
	"time"
)

// Binance Here is an implementation of a fake plugin which returns simulated data points.
type Binance struct {
	logger hclog.Logger
}

func (g *Binance) FetchPrices(symbols []string) []types.Price {
	// todo: fetch prices for symbols from binance http endpoint, for the time being, we just simulate fake data.
	g.logger.Debug("message from Binance.FetchPrices")
	// some fake data is simulated here since none data plugin_client is clarified.
	g.logger.Debug("fetching data prices from plugin_client: ", reflect2.TypeOfPtr(g).String())
	var prices []types.Price
	for _, s := range symbols {
		p := types.Price{
			Timestamp: time.Now().UnixMilli(),
			Symbol:    s,
			Price:     types.SimulatedPrice,
		}
		prices = append(prices, p)
	}

	return prices
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
