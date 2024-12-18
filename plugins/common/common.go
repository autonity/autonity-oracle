package common

import (
	"autonity-oracle/config"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shopspring/decimal"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	Zero                 = big.NewInt(0)
	DefaultForexSymbols  = []string{"EUR-USD", "JPY-USD", "GBP-USD", "AUD-USD", "CAD-USD", "SEK-USD"}
	NTNATNSymbol         = "NTN-ATN"
	DefaultCryptoSymbols = []string{"ATN-USDC", "NTN-USDC", NTNATNSymbol}
	DefaultUSDCSymbol    = "USDC-USD"
	ErrDataNotAvailable  = fmt.Errorf("data is not available")
	ErrKnownSymbols      = fmt.Errorf("the data source does not have all the data asked by oracle server")
	ErrAccessLimited     = fmt.Errorf("access rate is limited, please check your subscription from data provider")
)

const (
	DefaultAMMDataUpdateInterval = 1
	AutonityCryptoDecimals       = 18 // both NTN and the Wrapped ATN take 18 as the decimal.
	USDCDecimals                 = 6  // the decimal of USDC coin in autonity L1 network.
	CryptoToUsdcDecimals         = 18 // the data precision in oracle contract.
)

type Price struct {
	Symbol string `json:"symbol,omitempty"`
	Price  string `json:"price,omitempty"`
}

type Prices []Price

type Plugin struct {
	version          string
	availableSymbols map[string]struct{}
	symbolSeparator  string // "|", "/", "-", ",", "." or with a no separator "".
	logger           hclog.Logger
	client           DataSourceClient
	conf             *config.PluginConfig
	cachePrices      map[string]types.Price
}

func NewPlugin(conf *config.PluginConfig, client DataSourceClient, version string) *Plugin {
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

	availableSymbols, unRecognizableSymbols, availableSymMap := p.resolveSymbols(symbols)
	if len(availableSymbols) == 0 {
		report.UnRecognizableSymbols = unRecognizableSymbols
		return report, ErrKnownSymbols
	}

	cPRs, err := p.fetchPricesFromCache(availableSymbols)
	if err == nil {
		report.Prices = cPRs
		report.UnRecognizableSymbols = unRecognizableSymbols
		return report, nil
	}

	// fetch data from data source.
	res, err := p.client.FetchPrice(availableSymbols)
	if err != nil {
		return report, err
	}

	p.logger.Info("sampled data", "data", res)

	now := time.Now().Unix()
	for _, v := range res {
		dec, err := decimal.NewFromString(v.Price)
		if err != nil {
			p.logger.Error("cannot convert price string to decimal: ", "price", v.Price, "error", err.Error())
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
	report.UnRecognizableSymbols = unRecognizableSymbols
	return report, nil
}

func (p *Plugin) State() (types.PluginState, error) {
	var state types.PluginState

	symbols, err := p.client.AvailableSymbols()
	if err != nil {
		return state, err
	}

	if len(p.availableSymbols) != 0 {
		for k := range p.availableSymbols {
			symbol := k
			delete(p.availableSymbols, symbol)
		}
	}

	for _, s := range symbols {
		p.availableSymbols[s] = struct{}{}
	}

	for _, symbol := range symbols {
		if len(symbol) != 0 {
			p.symbolSeparator = ResolveSeparator(symbol)
			break
		}
	}

	state.Version = p.version
	state.AvailableSymbols = symbols
	state.KeyRequired = p.client.KeyRequired()
	return state, nil
}

func (p *Plugin) Close() {
	if p.client != nil {
		p.client.Close()
	}
}

// resolveSymbols resolve supported symbols of provider, and it builds the mapping of symbols from `-` separated pattern to those
// pattens supported by data providers, and filter outs those un-supported symbols.
func (p *Plugin) resolveSymbols(askedSymbols []string) ([]string, []string, map[string]string) {
	var supported []string
	var unRecognizable []string

	symbolsMapping := make(map[string]string)

	for _, askedSym := range askedSymbols {
		converted := ConvertSymbol(askedSym, p.symbolSeparator)
		if _, ok := p.availableSymbols[converted]; !ok {
			unRecognizable = append(unRecognizable, askedSym)
			continue
		}
		supported = append(supported, converted)
		symbolsMapping[converted] = askedSym
	}
	return supported, unRecognizable, symbolsMapping
}

func (p *Plugin) fetchPricesFromCache(availableSymbols []string) ([]types.Price, error) {
	var prices []types.Price
	now := time.Now().Unix()
	for _, s := range availableSymbols {
		pr, ok := p.cachePrices[s]
		if !ok {
			return nil, fmt.Errorf("no data buffered")
		}

		if now-pr.Timestamp >= int64(p.conf.DataUpdateInterval) {
			return nil, fmt.Errorf("data is too old")
		}

		prices = append(prices, pr)
	}
	return prices, nil
}

// LoadPluginConf is called from plugin main() to load plugin's conf from system env.
func LoadPluginConf(cmd string) (*config.PluginConfig, error) {
	name := filepath.Base(cmd)
	conf := os.Getenv(name)
	var c config.PluginConfig
	err := json.Unmarshal([]byte(conf), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func ResolveSeparator(symbol string) string {
	if i := strings.IndexAny(symbol, "|/-,."); i != -1 {
		chars := strings.Split(symbol, "")
		return chars[i]
	}
	return ""
}

// ConvertSymbol Convert source symbol to symbol with new separator, it assumes that the source symbol has one of these
// "|/-,." separators, otherwise it returns the source symbol without converting.
func ConvertSymbol(src string, toSep string) string {
	srcSep := ResolveSeparator(src)
	if srcSep == "" {
		return src
	}
	subs := strings.Split(src, srcSep)
	return strings.Join(subs, toSep)
}

func ResolveConf(cmd string, defConf *config.PluginConfig) *config.PluginConfig {

	conf, err := LoadPluginConf(cmd)
	if err != nil {
		println("cannot load conf: ", err.Error(), cmd)
		os.Exit(-1)
	}

	if conf.Timeout == 0 {
		conf.Timeout = defConf.Timeout
	}

	if conf.DataUpdateInterval == 0 {
		conf.DataUpdateInterval = defConf.DataUpdateInterval
	}

	if len(conf.Scheme) == 0 {
		conf.Scheme = defConf.Scheme
	}

	if len(conf.Endpoint) == 0 {
		conf.Endpoint = defConf.Endpoint
	}

	if len(conf.Key) == 0 {
		conf.Key = defConf.Key
	}

	if len(conf.Name) == 0 {
		conf.Name = defConf.Name
	}

	if len(conf.NTNTokenAddress) == 0 {
		conf.NTNTokenAddress = defConf.NTNTokenAddress
	}

	if len(conf.ATNTokenAddress) == 0 {
		conf.ATNTokenAddress = defConf.ATNTokenAddress
	}

	if len(conf.USDCTokenAddress) == 0 {
		conf.USDCTokenAddress = defConf.USDCTokenAddress
	}

	if len(conf.SwapAddress) == 0 {
		conf.SwapAddress = defConf.SwapAddress
	}

	return conf
}

// PluginServe doesn't return until the plugin is done being executed.
func PluginServe(p *Plugin) {
	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{Impl: p},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
	})
}

func CheckHTTPStatusCode(code int) error {
	if code != http.StatusOK {
		switch code {
		case http.StatusForbidden:
			fallthrough
		case http.StatusTooManyRequests:
			return ErrAccessLimited
		default:
			return fmt.Errorf("error return from data source, status code: %d", code)
		}
	}
	return nil
}

func ComputeDerivedPrice(ntnUSD, atnUSD string) (Price, error) {
	var priceNTNATN Price
	pNTN, err := decimal.NewFromString(ntnUSD)
	if err != nil {
		return priceNTNATN, err
	}

	pATN, err := decimal.NewFromString(atnUSD)
	if err != nil {
		return priceNTNATN, err
	}

	if pATN.IsZero() {
		return priceNTNATN, fmt.Errorf("div with zero of ATN price")
	}

	priceNTNATN.Symbol = NTNATNSymbol
	priceNTNATN.Price = pNTN.Div(pATN).String()
	return priceNTNATN, nil
}
