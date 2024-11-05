package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/plugins/crypto_airswap/erc20"
	swaperc20 "autonity-oracle/plugins/crypto_airswap/swap_erc20"
	"autonity-oracle/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ecommon "github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hashicorp/go-hclog"
	ring "github.com/zfjagann/golang-ring"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	orderBookCapacity = 64
	version           = "v0.0.1"
	ATNUSDC           = "ATN-USDC"
	NTNUSDC           = "NTN-USDC"
	supportedSymbols  = []string{ATNUSDC, NTNUSDC}
	priceFile         = "airswap_prices.json"
	NTNTokenAddress   = types.AutonityContractAddress // Autonity contract is the protocol contract of NTN token
)

var defaultConfig = types.PluginConfig{
	Name:               "crypto_airswap",
	Scheme:             "wss",                             // todo: update this on redeployment of infra
	Endpoint:           "rpc1.piccadilly.autonity.org/ws", // todo: update this on redeployment of infra
	Timeout:            10,                                // 10s
	DataUpdateInterval: 30,                                // 30s
	// todo: update below protocol contract addresses on redeployment of protocols.
	ATNTokenAddress:   "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2", // Wrapped ATN ERC20 contract address on Autonity blockchain.
	USDCTokenAddress:  "0x3a60C03a86eEAe30501ce1af04a6C04Cf0188700", // USDC ERC20 contract address on Autonity blockchain.
	SwapAddress:       "0x28363983213F88C759b501E3a5888458178cD5E7", // AirSwap SwapERC20 contract address on Autonity blockchain.
	DataPointStoreDir: ".",
}

type Order struct {
	cryptoToken  ecommon.Address
	cryptoAmount *big.Int
	usdcAmount   *big.Int
}

type AirswapClient struct {
	conf   *types.PluginConfig
	client *ethclient.Client
	logger hclog.Logger

	atnContract *erc20.Erc20
	atnAddress  ecommon.Address

	ntnContract *erc20.Erc20
	ntnAddress  ecommon.Address

	usdcContract *erc20.Erc20
	usdcAddress  ecommon.Address

	swapContract *swaperc20.Swaperc20

	chSwapEvent  chan *swaperc20.Swaperc20SwapERC20
	subSwapEvent event.Subscription

	doneCh   chan struct{}
	ticker   *time.Ticker // the clock interval to recover L1 connectivity.
	lostSync bool

	ntnOrderBooks ring.Ring

	atnOrderBooks ring.Ring

	priceMutex sync.RWMutex

	lastAggregatedPrices map[ecommon.Address]common.Price
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

	atnAddress := ecommon.HexToAddress(conf.ATNTokenAddress)
	atnContract, err := erc20.NewErc20(atnAddress, client)
	if err != nil {
		logger.Error("cannot bind ATN ERC20 contract", "error", err)
		return nil, err
	}

	ntnAddress := NTNTokenAddress
	ntnContract, err := erc20.NewErc20(ntnAddress, client)
	if err != nil {
		logger.Error("cannot bind NTN ERC20 contract", "error", err)
		return nil, err
	}

	usdcAddress := ecommon.HexToAddress(conf.USDCTokenAddress)
	usdcContract, err := erc20.NewErc20(usdcAddress, client)
	if err != nil {
		logger.Error("cannot bind USDC ERC20 contract", "error", err)
		return nil, err
	}

	ac := &AirswapClient{
		conf:                 conf,
		client:               client,
		logger:               logger,
		atnContract:          atnContract,
		atnAddress:           atnAddress,
		ntnContract:          ntnContract,
		ntnAddress:           ntnAddress,
		usdcContract:         usdcContract,
		usdcAddress:          usdcAddress,
		swapContract:         swapContract,
		doneCh:               make(chan struct{}),
		ticker:               time.NewTicker(time.Minute),
		lastAggregatedPrices: map[ecommon.Address]common.Price{},
	}

	// load historic prices of ATN-USDC and NTN-USDC or the price of the genesis.
	ac.initPrices()

	ac.atnOrderBooks.SetCapacity(orderBookCapacity)
	ac.ntnOrderBooks.SetCapacity(orderBookCapacity)

	if err = ac.EventSubscription(); err != nil {
		return nil, err
	}

	return ac, nil
}

func (e *AirswapClient) EventSubscription() error {
	// subscribe on-chain swap event of SwapERC20.
	chSwapEvent := make(chan *swaperc20.Swaperc20SwapERC20)
	subSwapEvent, err := e.swapContract.WatchSwapERC20(new(bind.WatchOpts), chSwapEvent, nil, nil)
	if err != nil {
		e.logger.Error("cannot watch swap event", "error", err)
		return err
	}
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
		case airSwapEvent := <-e.chSwapEvent:
			e.logger.Debug("receiving a SwapERC20 event", "event", airSwapEvent, "nonce", airSwapEvent.Nonce.Uint64())
			if err := e.handleSwapEvent(airSwapEvent.Raw.TxHash, airSwapEvent); err != nil {
				e.logger.Error("handle swap event failed", "error", err)
				continue
			}

		case <-e.ticker.C:
			e.checkHealth()
		}
	}
}

// handleSwapEvent, handles a single swap event at a time, if a txn contains multiple swap events, this function will
// be called with multiple times as the client subscribe every single swap event from L1. Processing one event at a
// time also make the logic simple and clear.
func (e *AirswapClient) handleSwapEvent(txnHash ecommon.Hash, swapEvent *swaperc20.Swaperc20SwapERC20) error {
	// pull the logs of the txn which issues the swap event.
	txnReceipt, err := e.client.TransactionReceipt(context.Background(), txnHash)
	if err != nil {
		e.logger.Error("cannot get transaction receipt", "error", err, "txnHash", txnHash)
		return err
	}

	order, err := e.extractOrder(txnReceipt, swapEvent)
	if err != nil {
		e.logger.Error("failed to extract the exchanges from txn receipts", "error", err, "txnHash", txnHash)
		return err
	}

	// then do the computing of price of ATN-USDC or NTN-USDC
	var orderBook *ring.Ring
	if order.cryptoToken == e.atnAddress {
		orderBook = &e.atnOrderBooks
	} else {
		orderBook = &e.ntnOrderBooks
	}

	lastAggregatedPrice, err := aggregatePrice(orderBook, order)
	if err != nil {
		e.logger.Error("failed to compute new price", "error", err, "txnHash", txnHash, "order", order)
		return err
	}

	// update the last aggregated price.
	e.updatePrice(order.cryptoToken, lastAggregatedPrice.FloatString(7))
	e.flushPrices() //nolint
	return nil
}

// extract order of the SwapERC20(nonce, signerWallet) event from the logs, as the functions which emits SwapERC20
// event can be called from any other contracts, thus it is not doable to parse the TXN's input
// data to get the direct inputs of the swap functions, thus we have to parse the ERC20 transfer events correspond to
// the airswap.SwapERC20(nonce, signerWallet) event to collect the exchange info.
// please visit:
// https://github.com/airswap/airswap-protocols/blob/develop/source/swap-erc20/contracts/SwapERC20.sol to find
// the details of the atomic swap between the two parties. The ERC20 and SwapERC20 events emitting follow in below
// patterns:
/*
log1: event: senderToken.Transfer(from: msg.sender, to: signerWallet, value: senderAmount);
log2: event: signerToken.Transfer(from: signerWallet, to: recipient/msg.sender, value: signerAmount);
log3, optional depends on if the fee > 0:
       if bonus > 0:
           event: signerToken.Transfer(from: signerWallet, to: msg.sender, bonus); // if the msg.sender is a staking node.
	       event: signerToken.Transfer(from: signerWallet, to: protocolFeeWallet, fee-bonus);
       else:
           event: signerToken.Transfer(from: signerWallet, to: protocolFeeWallet, fee);
log4, optional if swapEvent is emitted by a swapLight():
      event: signerToken.Transfer(from: signerWallet, to: protocolFeeWallet, lightFee);
log5, event: airswapERC20.SwapERC20(nonce, signerWallet);
*/
// From the pattern listed, we can see that log1 and log2 are the exchange of the two tokens, while log5 is the swap
// event that we subscribed, in between log2 and log5 there are multiple optional signerToken.Transfer events to pay
// the service fee/bonus in signerToken from signerWallet account to other parties (protocolFeeWallet, msg.Sender). As
// the fee/bonus are a small fraction of the signerAmount in signerToken of the exchange, thus we can filter out them
// from the log, and finally paired the log2 with log1 events as the exchange. With the signerWallet address is emitted
// by the SwapErc20(nonce, signerWallet) event, we can get the senderAmount of senderToken received by signerWallet, and
// the signerAmount of signerToken transfer by the signerWallet to get the exchange. In this plugin, we only care about
// the NTN token and the USDC token as our targeting liquidity market.
func (e *AirswapClient) extractOrder(receipt *types2.Receipt, targetSwapEvent *swaperc20.Swaperc20SwapERC20) (Order, error) {
	var order Order

	// todo: Jason, refine this implementation once the comment of https://github.com/airswap/airswap-protocols/issues/1341 is resolved.
	// iterate the logs to address the subscribed swapEvent,
	index := -1
	for i := len(receipt.Logs) - 1; i >= 0; i-- {
		// Check for the SwapERC20 event
		parsedSwap, err := e.swapContract.ParseSwapERC20(*receipt.Logs[i])
		if err != nil {
			e.logger.Debug("failed to parse log with swap event", "error", err)
			continue
		}

		// as the nonce is unique, check with nonce.
		if parsedSwap.Nonce == targetSwapEvent.Nonce {
			index = i
			break
		}
	}

	if index == -1 {
		return order, errors.New("failed to find matching swap in receipt")
	}

	var signerTokenAmount *big.Int
	var signerTokenAddress ecommon.Address
	var senderTokenAmount *big.Int
	var senderTokenAddress ecommon.Address

	// swap event is addressed, find the signerToken.Transfers and the senderToken.Transfer close to it.
	for i := index - 1; i >= 0; i-- {
		// just parse the ERC20 transfer events, the events could be ATN, NTN or USDC transfer events.
		transfer, err := e.ntnContract.ParseTransfer(*receipt.Logs[i])
		if err != nil {
			e.logger.Debug("failed to parse log with ERC20 transfer", "error", err)
			continue
		}

		eventEmitter := transfer.Raw.Address
		if eventEmitter != e.usdcAddress && eventEmitter != e.ntnAddress && eventEmitter != e.atnAddress {
			e.logger.Debug("skip none ATN, NTN & USDC swap event")
			return order, errors.New("skip none ATN, NTN & USDC swap event")
		}

		// now only transfers of ATN, NTN or USDC token can come to here.
		if transfer.From == targetSwapEvent.SignerWallet {
			if signerTokenAmount == nil || transfer.Value.Cmp(signerTokenAmount) > 0 {
				signerTokenAmount = transfer.Value
				signerTokenAddress = eventEmitter
			}
		} else {
			if transfer.To == targetSwapEvent.SignerWallet {
				senderTokenAmount = transfer.Value
				senderTokenAddress = eventEmitter
				break
			}
		}
	}

	if signerTokenAmount == nil || senderTokenAmount == nil {
		return order, errors.New("skip none ATN, NTN & USDC swap event")
	}

	if (signerTokenAddress == e.usdcAddress && senderTokenAddress == e.atnAddress) ||
		(signerTokenAddress == e.atnAddress && senderTokenAddress == e.usdcAddress) {
		order.cryptoToken = e.atnAddress
		if senderTokenAddress == e.atnAddress {
			order.cryptoAmount = senderTokenAmount
			order.usdcAmount = signerTokenAmount
		} else {
			order.cryptoAmount = signerTokenAmount
			order.usdcAmount = senderTokenAmount
		}
		return order, nil
	}

	if (signerTokenAddress == e.usdcAddress && senderTokenAddress == e.ntnAddress) ||
		(signerTokenAddress == e.ntnAddress && senderTokenAddress == e.usdcAddress) {
		order.cryptoToken = e.ntnAddress
		if senderTokenAddress == e.ntnAddress {
			order.cryptoAmount = senderTokenAmount
			order.usdcAmount = signerTokenAmount
		} else {
			order.cryptoAmount = signerTokenAmount
			order.usdcAmount = senderTokenAmount
		}
		return order, nil
	}

	// unexpected pairing, for example: ATN-NTN or NTN-ATN, they should be done by CDP, however if one want to create
	// ATN-NTN or NTN-ATN on aireswap, it should work. But we skip the pairing here.
	return order, errors.New("skip unexpected swap of ATN and NTN from airswap")
}

// compute new price once new settled order comes.
func aggregatePrice(orderBook *ring.Ring, order Order) (*big.Rat, error) {
	orderBook.Enqueue(order)
	recentOrders := orderBook.Values()
	aggregatedPrice, err := volumeWeightedPrice(recentOrders)
	if err != nil {
		return nil, err
	}
	return aggregatedPrice, nil
}

// flushPrices marshals the latest aggregated Price structs and writes them to a JSON file.
func (e *AirswapClient) flushPrices() error {
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

	// Marshal the slice of Price structs into JSON
	encoder := json.NewEncoder(file)

	var prices []common.Price
	for _, v := range e.lastAggregatedPrices {
		prices = append(prices, v)
	}

	if err = encoder.Encode(prices); err != nil {
		e.logger.Error("failed to flush prices", "error", err)
		return err
	}

	return nil
}

// historicPricesFromFile reads a slice of Price structs from a specified JSON file.
func (e *AirswapClient) historicPricesFromFile(dir string) ([]common.Price, error) {
	var prices []common.Price
	// Construct the full file path
	filePath := filepath.Join(dir, priceFile)

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		e.logger.Error("No historic aggregated prices can be found", "error", err)
		return nil, err
	}
	defer file.Close()

	// Decode the JSON data into the slice of Price structs
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&prices); err != nil {
		e.logger.Error("Cannot decode historic prices", "error", err)
		return nil, err
	}

	return prices, nil
}

func (e *AirswapClient) initPrices() {

	historicPrices, err := e.historicPricesFromFile(priceFile)
	if err != nil {
		e.logger.Error("cannot find historic price, waiting for realtime order book event")
		return
	}

	for _, p := range historicPrices {
		price := p
		if price.Symbol == NTNUSDC {
			e.lastAggregatedPrices[e.ntnAddress] = price
			continue
		}

		if price.Symbol == ATNUSDC {
			e.lastAggregatedPrices[e.atnAddress] = price
		}
	}
}

// volumeWeightedPrice calculates the volume-weighted exchange ratio of ATN or NTN to USDC.
func volumeWeightedPrice(orders []interface{}) (*big.Rat, error) {
	// Initialize total crypto and USDC amounts
	totalCrypto := new(big.Int)
	totalUSDC := new(big.Int)

	// Iterate through the orders to sum up the amounts
	for _, orderInterface := range orders {
		// Type assert to Order
		order, ok := orderInterface.(Order)
		if !ok {
			return nil, fmt.Errorf("invalid order type")
		}

		totalCrypto.Add(totalCrypto, order.cryptoAmount)
		totalUSDC.Add(totalUSDC, order.usdcAmount)
	}

	// Check if totalUSDC is zero to avoid division by zero
	if totalUSDC.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("total USDC amount is zero, cannot compute ratio")
	}

	// Calculate the weighted ratio as a fraction
	weightedRatio := new(big.Rat).SetFrac(totalCrypto, totalUSDC)

	return weightedRatio, nil
}

func (e *AirswapClient) updatePrice(tokenAddress ecommon.Address, price string) {
	e.priceMutex.Lock()
	defer e.priceMutex.Unlock()

	symbol := ATNUSDC
	if tokenAddress == e.ntnAddress {
		symbol = NTNUSDC
	}

	e.lastAggregatedPrices[tokenAddress] = common.Price{
		Symbol:    symbol,
		Price:     price,
		Timestamp: time.Now().Unix(),
	}
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

	for _, p := range e.lastAggregatedPrices {
		prices = append(prices, p)
	}

	if len(prices) == 0 {
		return prices, errors.New("dex-pluign hasn't receive any realtime swap event yet")
	}

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

	// start the SwapERC20 event watching for price aggregation of NTN-USDC & ATN-USDC
	go client.StartWatcher()

	adapter := common.NewPlugin(conf, client, version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
