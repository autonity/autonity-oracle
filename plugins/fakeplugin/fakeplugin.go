package main

import (
	"autonity-oracle/types"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shopspring/decimal"
	"os"
	"time"
)

var (
	version = "v0.0.1"
	pNTNUSD = decimal.RequireFromString("7.0")
	pNTNAUD = decimal.RequireFromString("9.856")
	pNTNCAD = decimal.RequireFromString("9.331")
	pNTNEUR = decimal.RequireFromString("6.8369")
	pNTNGBP = decimal.RequireFromString("5.691")
	pNTNJPY = decimal.RequireFromString("897.435")
	pNTNSEK = decimal.RequireFromString("72.163")
)

// FakePlugin Here is an implementation of a fake plugin which returns simulated data points.
type FakePlugin struct {
	logger hclog.Logger
}

func resolvePrice(s string) decimal.Decimal {
	defaultPrice := types.SimulatedPrice
	switch s {
	case "NTNUSD":
		defaultPrice = pNTNUSD
	case "NTNAUD":
		defaultPrice = pNTNAUD
	case "NTNCAD":
		defaultPrice = pNTNCAD
	case "NTNEUR":
		defaultPrice = pNTNEUR
	case "NTNGBP":
		defaultPrice = pNTNGBP
	case "NTNJPY":
		defaultPrice = pNTNJPY
	case "NTNSEK":
		defaultPrice = pNTNSEK
	}
	return defaultPrice
}

func (g *FakePlugin) FetchPrices(symbols []string) (types.PluginPriceReport, error) {
	var report types.PluginPriceReport
	for _, s := range symbols {

		price := resolvePrice(s)
		p := types.Price{
			Timestamp: time.Now().UnixMilli(),
			Symbol:    s,
			Price:     price,
		}
		report.Prices = append(report.Prices, p)
	}
	return report, nil
}

func (g *FakePlugin) GetVersion() (string, error) {
	return version, nil
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Debug,
		Output:     os.Stderr, // logging to stderr thus the framework can redirect the logs from plugin to plugin server.
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
