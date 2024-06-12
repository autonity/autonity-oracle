package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

// This plugin is only used for autonity round 4 game purpose, the data of NTN-USDC & ATN-USDC come from a simulated
// exchange service build by Clearmatics.
const (
	version = "v0.0.1"
	quote   = "quote"
	NTNATN  = "NTN-ATN"
	NTNUSDC = "NTN-USDC"
	ATNUSDC = "ATN-USDC"
)

// take piccadilly setup as the default setting
var routers = "api/orderbooks"
var defaultEndpoint = "cax.piccadilly.autonity.org"

var defaultConfig = types.PluginConfig{
	Key:                "",
	Scheme:             "https",
	Endpoint:           defaultEndpoint,
	Timeout:            10, //10s
	DataUpdateInterval: 30, //30s,
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
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "AutonityR4CAX",
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	return &CAXClient{
		conf:   conf,
		client: client,
		logger: logger,
	}
}

func (cc *CAXClient) KeyRequired() bool {
	return false
}

func (cc *CAXClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices
	priceMap := make(map[string]common.Price)

	for _, s := range symbols {
		// all CAXs of Autonity test networks do not provice NTN-ATN price, thus the price of NTN-ATN is derived from
		// the price of NTN-USDC and ATN-USDC for the time being.
		if s == NTNATN {
			continue
		}
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

	// for autonity round4 game, the price of "NTN-ATN" is derived from the price of "NTN-USDC" and "ATN-USDC"
	if _, ok := priceMap[NTNATN]; !ok && len(priceMap) == 2 {
		pNTN, ok := priceMap[NTNUSDC]
		if !ok {
			cc.logger.Error("missing NTN-USDC data to compute derived price: NTN-ATN")
			return prices, nil
		}

		pATN, ok := priceMap[ATNUSDC]
		if !ok {
			cc.logger.Error("missing ATN-USDC data to compute derived price: NTN-ATN")
			return prices, nil
		}

		// since only 3 symbols are supported, thus we assume the collected two symbols are NTN-USDC and ATN-USDC.
		pNTNATN, err := cc.computeDerivedPrice(pNTN, pATN)
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

	if err = common.CheckHTTPStatusCode(res.StatusCode); err != nil {
		return price, err
	}

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
	endpoint.Path = strings.Join([]string{routers, symbol, quote}, "/")
	return endpoint
}

// for autonity round4 game, "NTN-ATN" is derived from NTN-USDC and ATN-USDC.
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
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, NewCAXClient(conf), version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
