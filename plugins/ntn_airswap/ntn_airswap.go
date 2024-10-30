package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/plugins/ntn_airswap/erc20"
	swaperc20 "autonity-oracle/plugins/ntn_airswap/swap_erc20"
	"autonity-oracle/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hashicorp/go-hclog"
	lru "github.com/hashicorp/golang-lru"
	ring "github.com/zfjagann/golang-ring"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	orderBookCapacity   = 64
	swapEventCacheLimit = 1024
	version             = "v0.0.1"
	NTNUSDC             = "NTN-USDC"
	supportedSymbols    = []string{NTNUSDC}
	priceFile           = "NTN-USDC-Price.json"
	// todo: set an initial price of NTN-USDC on genesis, there is no price at all on airswapERC20 on genesis until there
	// is an exchange/swap happens. Otherwise, the oracle client might be slashed due to the missing of price of a symbol.
	priceOnGenesis = ""
)

var defaultConfig = types.PluginConfig{
	Name:               "ntn_airswap",
	Key:                "",
	Scheme:             "ws", // todo: set the protocol to connect to L1 blockchain node. ws or https
	Endpoint:           "",   // todo: set the host name or IP address and port for the service endpoint.
	Timeout:            10,   // 10s
	DataUpdateInterval: 30,   // todo: resolve the interval by according to the rate limit policy of the service end point.
	BaseTokenAddress:   "0x", // todo: set the wrapped NTN erc20 token address
	QuoteTokenAddress:  "0x", // todo: set the USDC erc20 token address
	SwapAddress:        "0x", // todo: set the airwapERC20 contract address
	DataPointStoreDir:  ".",  // the default directory to save historic aggregated prices of NTN-USDC from orders of airswapERC20.
}

type Order struct {
	ntnAmount  *big.Int
	usdcAmount *big.Int
	timeStamp  int64
}

type SwapEvent struct {
	ntnTransfer  *erc20.Erc20Transfer
	usdcTransfer *erc20.Erc20Transfer
	airSwap      *swaperc20.Swaperc20SwapERC20
}

func (d *SwapEvent) fill(src *SwapEvent) {
	if src.airSwap != nil {
		d.airSwap = src.airSwap
	}

	if src.ntnTransfer != nil {
		d.ntnTransfer = src.ntnTransfer
	}

	if src.usdcTransfer != nil {
		d.usdcTransfer = src.ntnTransfer
	}
}

func (d *SwapEvent) full() bool {
	return d.airSwap != nil && d.ntnTransfer != nil && d.usdcTransfer != nil
}

type AirswapClient struct {
	conf   *types.PluginConfig
	client *ethclient.Client
	logger hclog.Logger

	ntnContract  *erc20.Erc20
	usdcContract *erc20.Erc20
	swapContract *swaperc20.Swaperc20

	chNTNTransferEvent  chan *erc20.Erc20Transfer
	subNTNTransferEvent event.Subscription

	chUSDCTransferEvent  chan *erc20.Erc20Transfer
	subUSDCTransferEvent event.Subscription

	chSwapEvent  chan *swaperc20.Swaperc20SwapERC20
	subSwapEvent event.Subscription

	doneCh   chan struct{}
	ticker   *time.Ticker // the clock interval to recover L1 connectivity.
	lostSync bool

	swapEventCache *lru.Cache
	orderBooks     ring.Ring

	priceMutex          sync.RWMutex
	lastAggregatedPrice common.Price
}

func NewAirswapClient(conf *types.PluginConfig, logger hclog.Logger) (*AirswapClient, error) {
	url := conf.Scheme + "://" + conf.Endpoint
	client, err := ethclient.Dial(url)
	if err != nil {
		logger.Error("cannot dial to L1 node", "error", err)
		return nil, err
	}

	swapContract, err := swaperc20.NewSwaperc20(ecommon.HexToAddress(conf.SwapAddress), client)
	if err != nil {
		logger.Error("cannot bind airswapERC20 contract", "error", err)
		return nil, err
	}

	ntnContract, err := erc20.NewErc20(ecommon.HexToAddress(conf.BaseTokenAddress), client)
	if err != nil {
		logger.Error("cannot bind NTN ERC20 contract", "error", err)
		return nil, err
	}

	usdcContract, err := erc20.NewErc20(ecommon.HexToAddress(conf.QuoteTokenAddress), client)
	if err != nil {
		logger.Error("cannot bind USDC ERC20 contract", "error", err)
		return nil, err
	}

	swapEventCache, err := lru.New(swapEventCacheLimit)
	if err != nil {
		logger.Error("cannot create swapEventCache", "", err)
		return nil, err
	}

	ac := &AirswapClient{
		conf:           conf,
		client:         client,
		logger:         logger,
		ntnContract:    ntnContract,
		usdcContract:   usdcContract,
		swapContract:   swapContract,
		doneCh:         make(chan struct{}),
		swapEventCache: swapEventCache,
		ticker:         time.NewTicker(time.Minute),
	}

	ac.initPrice()

	ac.orderBooks.SetCapacity(orderBookCapacity)
	if err = ac.EventSubscription(); err != nil {
		return nil, err
	}

	return ac, nil
}

func (e *AirswapClient) EventSubscription() error {
	// subscribe on-chain transfer event of NTN.
	chNTNTransferEvent := make(chan *erc20.Erc20Transfer)
	subNTNTransferEvent, err := e.ntnContract.WatchTransfer(new(bind.WatchOpts), chNTNTransferEvent, nil, nil)
	if err != nil {
		e.logger.Error("cannot watch NTN transfer event", "error", err)
		return err
	}
	// subscribe on-chain transfer event of USDC.
	chUSDCTransferEvent := make(chan *erc20.Erc20Transfer)
	subUSDCTransferEvent, err := e.usdcContract.WatchTransfer(new(bind.WatchOpts), chUSDCTransferEvent, nil, nil)
	if err != nil {
		e.logger.Error("cannot watch USDC transfer event", "error", err)
		return err
	}
	// subscribe on-chain swap event of airsapERC20.
	chSwapEvent := make(chan *swaperc20.Swaperc20SwapERC20)
	subSwapEvent, err := e.swapContract.WatchSwapERC20(new(bind.WatchOpts), chSwapEvent, nil, nil)
	if err != nil {
		e.logger.Error("cannot watch swap event", "error", err)
		return err
	}
	e.chUSDCTransferEvent = chUSDCTransferEvent
	e.subUSDCTransferEvent = subUSDCTransferEvent
	e.chNTNTransferEvent = chNTNTransferEvent
	e.subNTNTransferEvent = subNTNTransferEvent
	e.chSwapEvent = chSwapEvent
	e.subSwapEvent = subSwapEvent
	return nil
}

func (e *AirswapClient) StartWatcher() {
	for {
		select {
		case <-e.doneCh:
			e.ticker.Stop()
			e.logger.Info("air-swap events watcher stopped")
			return
		case err := <-e.subSwapEvent.Err():
			if err != nil {
				e.logger.Info("subscription error of swap event", "error", err)
				e.handleConnectivityError()
				e.subSwapEvent.Unsubscribe()
			}
		case err := <-e.subUSDCTransferEvent.Err():
			if err != nil {
				e.logger.Info("subscription error of usdc transfer event", "error", err)
				e.handleConnectivityError()
				e.subUSDCTransferEvent.Unsubscribe()
			}
		case err := <-e.subSwapEvent.Err():
			if err != nil {
				e.logger.Info("subscription error of swap event", "error", err)
				e.handleConnectivityError()
				e.subSwapEvent.Unsubscribe()
			}
		case ntnTransferEvent := <-e.chNTNTransferEvent:
			e.logger.Debug("receiving ntn transfer event", "event", ntnTransferEvent)
			transfer := &SwapEvent{ntnTransfer: ntnTransferEvent}
			e.handleEvent(ntnTransferEvent.Raw.TxHash, transfer)
		case usdcTransferEvent := <-e.chUSDCTransferEvent:
			e.logger.Debug("receiving usdc transfer event", "event", usdcTransferEvent)
			transfer := &SwapEvent{usdcTransfer: usdcTransferEvent}
			e.handleEvent(usdcTransferEvent.Raw.TxHash, transfer)
		case airSwapEvent := <-e.chSwapEvent:
			e.logger.Debug("receiving air swap event", "event", airSwapEvent)
			swap := &SwapEvent{airSwap: airSwapEvent}
			e.handleEvent(airSwapEvent.Raw.TxHash, swap)
		case <-e.ticker.C:
			e.checkHealth()
		}
	}
}

func (e *AirswapClient) handleEvent(txnHash ecommon.Hash, swapEvent *SwapEvent) {

	// if event was cached, fill the sub events, and check if we can push it to the aggregator.
	if cached, ok := e.swapEventCache.Get(txnHash); ok {
		swap := cached.(*SwapEvent)
		swap.fill(swapEvent)
		// ready for price aggregation.
		if swap.full() {
			e.logger.Info("a full swap event received, going to aggregate the price", "event", swap)
			order := Order{
				ntnAmount:  swap.ntnTransfer.Value,
				usdcAmount: swap.usdcTransfer.Value,
				timeStamp:  time.Now().Unix(),
			}
			e.computePrice(order)
			e.swapEventCache.Remove(txnHash)
			return
		}
		return
	}

	// not cached yet, put it in the cache.
	e.swapEventCache.Add(txnHash, swapEvent)
}

// compute new price once new settled order comes.
func (e *AirswapClient) computePrice(order Order) {
	e.orderBooks.Enqueue(order)
	orders := e.orderBooks.Values()
	aggregatedPrice, err := volumeWeightedPrice(orders)
	if err != nil {
		return
	}
	e.updatePrice(aggregatedPrice.FloatString(7))
	e.flushPrice()
}

func (e *AirswapClient) flushPrice() error {
	e.priceMutex.RLock()
	defer e.priceMutex.RUnlock()

	// Construct the full file path
	filePath := filepath.Join(e.conf.DataPointStoreDir, priceFile)
	// Open the file for writing (create if it doesn't exist)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		e.logger.Error("cannot open price file", "error", err)
		return err
	}
	defer file.Close()

	// Marshal the Price struct into JSON
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(e.lastAggregatedPrice); err != nil {
		e.logger.Error("failed to flush last aggregated price", "error", err)
		return err
	}

	return nil
}

// readPriceFromFile reads the Price struct from a specified JSON file.
func (e *AirswapClient) readPriceFromFile(dir string) (common.Price, error) {
	var price common.Price
	// Construct the full file path
	filePath := filepath.Join(dir, priceFile)

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		e.logger.Error("No historic aggregated price can be find", "error", err)
		return common.Price{}, err
	}
	defer file.Close()

	// Decode the JSON data into the Price struct
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&price); err != nil {
		e.logger.Error("Cannot decode historic price", "error", err)
		return price, err
	}

	return price, nil
}

func (e *AirswapClient) initPrice() {
	e.lastAggregatedPrice.Symbol = NTNUSDC

	historicPrice, err := e.readPriceFromFile(priceFile)
	if err != nil {
		e.logger.Info("cannot find historic price, going to apply the genesis exchange ratio of NTN-USDC")
		e.lastAggregatedPrice.Price = priceOnGenesis
		return
	}

	if historicPrice.Symbol != NTNUSDC {
		e.logger.Info("wrong symbol address from historic price, going to apply the genesis exchange ratio of NTN-USDC")
		e.lastAggregatedPrice.Price = priceOnGenesis
		return
	}

	e.lastAggregatedPrice.Price = historicPrice.Price
}

// volumeWeightedPrice calculates the volume-weighted exchange ratio of NTN to USDC.
func volumeWeightedPrice(orders []interface{}) (*big.Rat, error) {
	// Initialize total NTN and USDC amounts
	totalNTN := new(big.Int)
	totalUSDC := new(big.Int)

	// Iterate through the orders to sum up the amounts
	for _, orderInterface := range orders {
		// Type assert to Order
		order, ok := orderInterface.(Order)
		if !ok {
			return nil, fmt.Errorf("invalid order type")
		}

		totalNTN.Add(totalNTN, order.ntnAmount)
		totalUSDC.Add(totalUSDC, order.usdcAmount)
	}

	// Check if totalUSDC is zero to avoid division by zero
	if totalUSDC.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("total USDC amount is zero, cannot compute ratio")
	}

	// Calculate the weighted ratio as a fraction
	weightedRatio := new(big.Rat).SetFrac(totalNTN, totalUSDC)

	return weightedRatio, nil
}

func (e *AirswapClient) updatePrice(price string) {
	e.priceMutex.Lock()
	defer e.priceMutex.Unlock()
	e.lastAggregatedPrice.Price = price
}

func (e *AirswapClient) checkHealth() {
	if e.lostSync {
		err := e.EventSubscription()
		if err != nil {
			e.logger.Info("rebuilding WS connectivity with L1 node", "error", err)
			return
		}
		e.lostSync = false
		return
	}

	h, err := e.client.BlockNumber(context.Background())
	if err != nil {
		e.logger.Error("get block number", "error", err.Error())
		return
	}

	e.logger.Debug("checking heart beat", "current height", h)
}

func (e *AirswapClient) handleConnectivityError() {
	e.lostSync = true
}

func (e *AirswapClient) KeyRequired() bool {
	return false
}

func (e *AirswapClient) FetchPrice(_ []string) (common.Prices, error) {
	e.priceMutex.RLock()
	defer e.priceMutex.RUnlock()
	var prices common.Prices
	prices = append(prices, e.lastAggregatedPrice)
	return prices, nil
}

func (e *AirswapClient) AvailableSymbols() ([]string, error) {
	return supportedSymbols, nil
}

func (e *AirswapClient) Close() {
	if e.client != nil {
		e.client.Close()
	}
	e.subSwapEvent.Unsubscribe()
	e.subNTNTransferEvent.Unsubscribe()
	e.subUSDCTransferEvent.Unsubscribe()
	e.doneCh <- struct{}{}
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	client, err := NewAirswapClient(conf, logger)
	if err != nil {
		return
	}

	// start the swap order book watching for price aggregation of NTN-USDC
	go client.StartWatcher()

	adapter := common.NewPlugin(conf, client, version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
