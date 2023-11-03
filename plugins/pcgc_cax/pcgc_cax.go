package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shopspring/decimal"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

// This plugin is only used for autonity round 4 game purpose, the data of NTN-USD & ATN-USD come from a simulated
// exchange service build by Clearmatics.
const (
	version   = "v0.0.1"
	orderbook = "orderbooks"
	quote     = "quote"
	NTNATN    = "NTN-ATN"
	NTNUSD    = "NTN-USD"
	ATNUSD    = "ATN-USD"
)

var defaultConfig = types.PluginConfig{
	Key:                "",
	Scheme:             "https",
	Endpoint:           "cax.devnet.clearmatics.network", // todo: replace it with cax.piccadilly.autonity.org for piccadilly network.
	Timeout:            10,                               //10s
	DataUpdateInterval: 3600,                             //3600s, todo: if there is a rate limit for the CAX service, we'd need to set it properly.
}

type CAXQuote struct {
	Timestamp string `json:"timestamp"`
	BidPrice  string `json:"bid_price"`
	BidAmount string `json:"bid_amount"`
	AskPrice  string `json:"ask_price"`
	AskAmount string `json:"ask_amount"`
}

type CAXClient struct {
	conf   *types.PluginConfig
	client *common.Client
	logger hclog.Logger
}

func NewCAXClient(conf *types.PluginConfig) *CAXClient {
	client := common.NewClient(conf.Key, time.Second*time.Duration(conf.Timeout), conf.Endpoint)
	if client == nil {
		panic("cannot create client for open exchange rate api")
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "AutonityR4CAX",
		Level:  hclog.Debug,
		Output: os.Stdout,
	})

	return &CAXClient{
		conf:   conf,
		client: client,
		logger: logger,
	}
}

func (cc *CAXClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices
	priceMap := make(map[string]common.Price)

	for _, s := range symbols {
		p, err := cc.fetchPrice(s)
		if err != nil {
			cc.logger.Error("query price", "error", err.Error())
			continue
		}
		priceMap[s] = p
		prices = append(prices, p)
	}

	if len(prices) == 0 {
		return nil, common.ErrDataNotAvailable
	}

	// for autonity round4 game, the price of "NTN-ATN" is derived from the price of "NTN-USD" and "ATN-USD"
	if _, ok := priceMap[NTNATN]; !ok && len(priceMap) == 2 {
		// since only 3 symbols are supported, thus we assume the collected two symbols are NTN-USD and ATN-USD.
		pNTNATN, err := cc.computeDerivedPrice(priceMap[NTNUSD], priceMap[ATNUSD])
		if err != nil {
			cc.logger.Error("compute derived price NTN-ATN", "error", err.Error())
			return prices, nil
		}
		prices = append(prices, pNTNATN)
	}

	return prices, nil
}

func (cc *CAXClient) fetchPrice(symbol string) (common.Price, error) {
	var price common.Price
	u := cc.buildURL(symbol)
	res, err := cc.client.Conn.Request(cc.conf.Scheme, u)
	if err != nil {
		cc.logger.Error("https request", "error", err.Error())
		return price, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		cc.logger.Error("io read", "error", err.Error())
		return price, err
	}

	var result CAXQuote
	err = json.Unmarshal(body, &result)
	if err != nil {
		cc.logger.Error("unmarshal quote", "error", err.Error())
		return price, err
	}

	if result.Timestamp == "" {
		cc.logger.Error("data source returns", "data", string(body))
		return price, common.ErrDataNotAvailable
	}

	askPrice, err := decimal.NewFromString(result.AskPrice)
	if err != nil {
		cc.logger.Error("invalid askPrice value", "error", err)
		return price, err
	}

	bidPrice, err := decimal.NewFromString(result.BidPrice)
	if err != nil {
		cc.logger.Error("invalid bidPrice value", "error", err)
		return price, err
	}

	// the aggregated price takes the average value of ask and bid prices.
	price.Price = askPrice.Add(bidPrice).Div(decimal.NewFromInt(2)).String()
	price.Symbol = symbol

	return price, nil
}

func (cc *CAXClient) buildURL(symbol string) *url.URL {
	endpoint := &url.URL{}
	endpoint.Path = strings.Join([]string{orderbook, symbol, quote}, "/")
	return endpoint
}

// for autonity round4 game, "NTN-ATN" is derived from NTN-USD and ATN-USD.
func (cc *CAXClient) computeDerivedPrice(ntnUSD, atnUSD common.Price) (common.Price, error) {
	var priceNTNATN common.Price
	pNTN, err := decimal.NewFromString(ntnUSD.Price)
	if err != nil {
		return priceNTNATN, err
	}

	pATN, err := decimal.NewFromString(atnUSD.Price)
	if err != nil {
		return priceNTNATN, err
	}

	priceNTNATN.Symbol = NTNATN
	priceNTNATN.Price = pNTN.Div(pATN).String()
	return priceNTNATN, nil
}

func (cc *CAXClient) AvailableSymbols() ([]string, error) {
	return common.DefaultCryptoSymbols, nil
}

func (cc *CAXClient) Close() {
	cc.client.Conn.Close()
}

func main() {
	conf, err := common.LoadPluginConf(os.Args[0])
	if err != nil {
		println("cannot load conf: ", err.Error(), os.Args[0])
		os.Exit(-1)
	}
	common.ResolveConf(&conf, &defaultConfig)
	client := NewCAXClient(&conf)
	adapter := common.NewPlugin(&conf, client, version)
	defer adapter.Close()

	var pluginMap = map[string]plugin.Plugin{
		"adapter": &types.AdapterPlugin{Impl: adapter},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: types.HandshakeConfig,
		Plugins:         pluginMap,
	})
}
