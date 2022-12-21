package crypto_provider

import (
	"autonity-oralce/types"
	"errors"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var ErrInvalidResponse = errors.New("invalid http response")

type BinanceAdapter struct {
	tradePool types.TradePool
	config    *types.AdapterConfig

	aliveAt    int64
	sentPingAt int64
	alive      bool
	symbols    []string
	wsUrl      string
	wsCon      *websocket.Conn
}

func NewBinanceAdapter() *BinanceAdapter {
	return &BinanceAdapter{}
}

func (ba *BinanceAdapter) Name() string {
	return "Binance"
}

func (ba *BinanceAdapter) Version() string {
	return "0.0.1"
}

func (ba *BinanceAdapter) Initialize(config *types.AdapterConfig, tradePool types.TradePool) error {
	// todo: check config.
	ba.config = config
	ba.tradePool = tradePool
	return nil
}

func (ba *BinanceAdapter) Start() error {
	for _, s := range ba.symbols {
		trs, err := ba.fetchLatestTrades(s)
		if err != nil {
			//todo: log warnings.
		}
		if len(trs) > 0 {
			ba.tradePool.PushTrades(ba.Name(), s, trs, false)
		}
	}

	//todo: start the web socket routine to connect to the ws endpoint and get those trades being notified.
	return nil
}

func (ba *BinanceAdapter) fetchLatestTrades(symbol string) (types.Trades, error) {

	// Get candles from Binance
	// reference: https://binance-docs.github.io/apidocs/spot/en/#kline-candlestick-data
	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/klines", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("symbol", strings.ReplaceAll(symbol, "/", ""))
	q.Add("interval", "1m")
	q.Add("limit", "10")
	req.URL.RawQuery = q.Encode()
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.ContentLength < 1 || resp.StatusCode != 200 {
		return nil, ErrInvalidResponse
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	responseString := string(responseData)
	// todo: take off data from response.

	return nil, nil
}

func (ba *BinanceAdapter) Stop() error {
	// todo: stop the ticker routine
	return nil
}

func (ba *BinanceAdapter) Symbols() []string {
	return ba.symbols
}

func (ba *BinanceAdapter) Alive() bool {
	return ba.alive
}

func (ba *BinanceAdapter) Config() *types.AdapterConfig {
	return ba.config
}

// TODO: implement the adapter which fetch trades via pull mode and a push mode from binance endpoint.
