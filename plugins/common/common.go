package common

import (
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Price struct {
	Symbol string `json:"symbol,omitempty"`
	Price  string `json:"price,omitempty"`
}

type Prices []Price

var DefaultForexSymbols = []string{"EUR/USD", "JPY/USD", "GBP/USD", "AUD/USD", "CAD/USD", "SEK/USD"}

type Plugin struct {
	version          string
	availableSymbols map[string]struct{}
	separatedStyle   bool
	logger           hclog.Logger
	client           DataSourceClient
	conf             *types.PluginConfig
	cachePrices      map[string]types.Price
}

func NewPlugin(conf *types.PluginConfig, client DataSourceClient, version string) *Plugin {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:       conf.Name,
		Level:      hclog.Debug,
		Output:     os.Stderr, // logging into stderr thus the go-plugin can redirect the logs to plugin server.
		JSONFormat: true,
	})

	return &Plugin{
		version:          version,
		logger:           logger,
		client:           client,
		conf:             conf,
		availableSymbols: make(map[string]struct{}),
		cachePrices:      make(map[string]types.Price),
	}
}

func (p *Plugin) FetchPrices(symbols []string) (types.PluginPriceReport, error) {
	var report types.PluginPriceReport

	availableSymbols, badSymbols, availableSymMap := p.resolveSymbols(symbols)
	if len(availableSymbols) == 0 {
		p.logger.Warn("no available symbols from plugin", "plugin", p.conf.Name)
		report.BadSymbols = badSymbols
		return report, fmt.Errorf("no available symbols")
	}

	cPRs, err := p.fetchPricesFromCache(availableSymbols)
	if err == nil {
		report.Prices = cPRs
		report.BadSymbols = badSymbols
		return report, nil
	}

	// fetch data from data source.
	res, err := p.client.FetchPrice(availableSymbols)
	if err != nil {
		return report, err
	}

	p.logger.Debug("sampled data points", res)

	now := time.Now().Unix()
	for _, v := range res {
		dec, err := decimal.NewFromString(v.Price)
		if err != nil {
			p.logger.Error("cannot convert price string to decimal: ", v.Price, err)
			continue
		}

		pr := types.Price{
			Timestamp: now,
			Symbol:    availableSymMap[v.Symbol], // set the symbol with the symbol style used in oracle server side.
			Price:     dec,
		}
		p.cachePrices[v.Symbol] = pr
		report.Prices = append(report.Prices, pr)
	}
	report.BadSymbols = badSymbols
	return report, nil
}

func (p *Plugin) State() (types.PluginState, error) {
	var state types.PluginState

	symbols, err := p.client.AvailableSymbols()
	if err != nil {
		return state, err
	}

	if len(p.availableSymbols) != 0 {
		for _, s := range symbols {
			delete(p.availableSymbols, s)
		}
	}

	for _, s := range symbols {
		p.availableSymbols[s] = struct{}{}
	}

	for k := range p.availableSymbols {
		if strings.Contains(k, "/") {
			p.separatedStyle = true
			break
		}
		p.separatedStyle = false
		break
	}

	state.Version = p.version
	state.AvailableSymbols = symbols

	return state, nil
}

func (p *Plugin) Close() {
	if p.client != nil {
		p.client.Close()
	}
}

// resolveSymbols resolve available symbols of provider, and it converts symbols from `/` separated pattern to none `/`
// separated pattern if the provider uses the none `/` separated pattern of symbols.
func (p *Plugin) resolveSymbols(symbols []string) ([]string, []string, map[string]string) {
	var available []string
	var badSymbols []string

	availableSymbolMap := make(map[string]string)

	for _, raw := range symbols {

		if p.separatedStyle {
			if _, ok := p.availableSymbols[raw]; !ok {
				badSymbols = append(badSymbols, raw)
				continue
			}
			available = append(available, raw)
			availableSymbolMap[raw] = raw
		} else {

			nSymbol, err := helpers.NoneSeparatedSymbol(raw)
			if err != nil {
				badSymbols = append(badSymbols, raw)
			}

			if _, ok := p.availableSymbols[nSymbol]; !ok {
				badSymbols = append(badSymbols, raw)
				continue
			}
			available = append(available, nSymbol)
			availableSymbolMap[nSymbol] = raw
		}
	}
	return available, badSymbols, availableSymbolMap
}

func (p *Plugin) fetchPricesFromCache(availableSymbols []string) ([]types.Price, error) {
	var prices []types.Price
	now := time.Now().Unix()
	for _, s := range availableSymbols {
		pr, ok := p.cachePrices[s]
		if !ok {
			return nil, fmt.Errorf("no data buffered")
		}

		if now-pr.Timestamp > int64(p.conf.DataUpdateInterval) {
			return nil, fmt.Errorf("data is too old")
		}

		prices = append(prices, pr)
	}
	return prices, nil
}

// LoadPluginConf is called from plugin main() to load plugin's conf from system env.
func LoadPluginConf(cmd string) (types.PluginConfig, error) {
	name := filepath.Base(cmd)
	conf := os.Getenv(name)
	var c types.PluginConfig
	err := json.Unmarshal([]byte(conf), &c)
	if err != nil {
		return c, err
	}
	return c, nil
}
