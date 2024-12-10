package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/plugins/crypto_airswap/erc20"
	swaperc20 "autonity-oracle/plugins/crypto_airswap/swap_erc20"
	"autonity-oracle/types"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ecommon "github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hashicorp/go-hclog"
	ring "github.com/zfjagann/golang-ring"
	"math"
	"math/big"
	"os"
	"sync"
	"time"
)

var (
	orderBookCapacity     = 64
	version               = "v0.2.0"
	ATNPairToStableCoin   = "ATN-USDC"
	NTNPairToStableCoin   = "NTN-USDC"
	supportedSymbols      = []string{ATNPairToStableCoin, NTNPairToStableCoin, common.NTNATNSymbol}
	NTNTokenAddress       = types.AutonityContractAddress // Autonity contract is the protocol contract of NTN token
	usdStableCoinDecimals = common.USDCDecimals
)

// Todo: require a config PR to finalize below configuration for the target autonity network.
var defaultConfig = types.PluginConfig{
	Name:               "crypto_airswap",
	Scheme:             "wss",                             // todo: update this on redeployment of infra
	Endpoint:           "rpc1.piccadilly.autonity.org/ws", // todo: update this on redeployment of infra
	Timeout:            10,                                // 10s
	DataUpdateInterval: 30,                                // 30s
	// todo: update below protocol contract addresses on redeployment of protocols.
	NTNTokenAddress: NTNTokenAddress.Hex(),                        // NTN ERC20 token address on the target blockchain.
	ATNTokenAddress: "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2", // Wrapped ATN ERC20 contract address on the target blockchain.
	USDTokenAddress: "0x3a60C03a86eEAe30501ce1af04a6C04Cf0188700", // The USD stable coins' ERC20 contract address, could be USDC or USDX.
	SwapAddress:     "0x28363983213F88C759b501E3a5888458178cD5E7", // AirSwap SwapERC20 contract address on the target blockchain.
}

type Order struct {
	cryptoToken  ecommon.Address
	cryptoAmount *big.Int
	usdAmount    *big.Int
}

type AirswapClient struct {
	conf   *types.PluginConfig
	client *ethclient.Client
	logger hclog.Logger

	atnAddress ecommon.Address
	usdAddress ecommon.Address // the USD stable coins' address. Could be usdc or usdx erc 20 token address.
	ntnAddress ecommon.Address

	// ERC20 Transfer event parser.
	erc20LogParser *erc20.Erc20

	// SwapERC20 event watcher and log parser.
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

	erc20LogParser, err := erc20.NewErc20(NTNTokenAddress, client)
	if err != nil {
		logger.Error("cannot bind NTN ERC20 contract", "error", err)
		return nil, err
	}

	ac := &AirswapClient{
		conf:                 conf,
		client:               client,
		logger:               logger,
		atnAddress:           ecommon.HexToAddress(conf.ATNTokenAddress),
		ntnAddress:           NTNTokenAddress,
		usdAddress:           ecommon.HexToAddress(conf.USDTokenAddress),
		erc20LogParser:       erc20LogParser,
		swapContract:         swapContract,
		doneCh:               make(chan struct{}),
		ticker:               time.NewTicker(time.Minute),
		lastAggregatedPrices: map[ecommon.Address]common.Price{},
	}

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

	logs := txnReceipt.Logs
	order, err := e.extractOrder(logs, swapEvent)
	if err != nil {
		e.logger.Error("failed to extract the exchanges from txn receipts", "error", err, "txnHash", txnHash)
		return err
	}

	// then do the computing of price of supported symbols.
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
	e.updatePrice(order.cryptoToken, lastAggregatedPrice.FloatString(common.CryptoToUsdcDecimals))
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
// the NTN token and the USDC(X) token as our targeting liquidity market.
func (e *AirswapClient) extractOrder(logs []*types2.Log, targetSwapEvent *swaperc20.Swaperc20SwapERC20) (Order, error) {
	var order Order

	// todo: Jason, refine this implementation once the comment of https://github.com/airswap/airswap-protocols/issues/1341 is resolved.
	// iterate the logs to address the subscribed swapEvent,
	index := -1
	for i := len(logs) - 1; i >= 0; i-- {
		// Check for the SwapERC20 event
		parsedSwap, err := e.swapContract.ParseSwapERC20(*logs[i])
		if err != nil {
			e.logger.Debug("failed to parse log with swap event", "error", err)
			continue
		}

		// as the nonce is unique, check with nonce and signer wallet.
		if parsedSwap.Nonce.Cmp(targetSwapEvent.Nonce) == 0 && parsedSwap.SignerWallet == targetSwapEvent.SignerWallet {
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
		// just parse the ERC20 transfer events, the events could be ATN, NTN or the pegged USD stable coin transfer events.
		transfer, err := e.erc20LogParser.ParseTransfer(*logs[i])
		if err != nil {
			e.logger.Debug("failed to parse log with ERC20 transfer", "error", err)
			continue
		}

		eventEmitter := transfer.Raw.Address
		if eventEmitter != e.usdAddress && eventEmitter != e.ntnAddress && eventEmitter != e.atnAddress {
			e.logger.Debug("skip none ATN, NTN & pegged USD stable coin swap event")
			return order, errors.New("skip none ATN, NTN & pegged USD stable coin swap event")
		}

		// now only transfers of ATN, NTN or USDC(X) token can come to here.
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
		return order, errors.New("skip none ATN, NTN & pegged USD stable coin swap event")
	}

	if (signerTokenAddress == e.usdAddress && senderTokenAddress == e.atnAddress) ||
		(signerTokenAddress == e.atnAddress && senderTokenAddress == e.usdAddress) {
		order.cryptoToken = e.atnAddress
		if senderTokenAddress == e.atnAddress {
			order.cryptoAmount = senderTokenAmount
			order.usdAmount = signerTokenAmount
		} else {
			order.cryptoAmount = signerTokenAmount
			order.usdAmount = senderTokenAmount
		}
		return order, nil
	}

	if (signerTokenAddress == e.usdAddress && senderTokenAddress == e.ntnAddress) ||
		(signerTokenAddress == e.ntnAddress && senderTokenAddress == e.usdAddress) {
		order.cryptoToken = e.ntnAddress
		if senderTokenAddress == e.ntnAddress {
			order.cryptoAmount = senderTokenAmount
			order.usdAmount = signerTokenAmount
		} else {
			order.cryptoAmount = signerTokenAmount
			order.usdAmount = senderTokenAmount
		}
		return order, nil
	}

	// exchange of ATN and NTN is not watched, we skip the order.
	return order, errors.New("skip process swap event of ATN and NTN from airswap")
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

// volumeWeightedPrice calculates the volume-weighted exchange ratio of ATN or NTN to the pegged USD stable coin.
func volumeWeightedPrice(orders []interface{}) (*big.Rat, error) {
	// Initialize total crypto and the pegged USD stable coin amounts
	totalCrypto := new(big.Int)
	totalUSDStableCoin := new(big.Int)

	// Iterate through the orders to sum up the amounts
	for _, orderInterface := range orders {
		// Type assert to Order
		order, ok := orderInterface.(Order)
		if !ok {
			return nil, fmt.Errorf("invalid order type")
		}

		totalCrypto.Add(totalCrypto, order.cryptoAmount)
		totalUSDStableCoin.Add(totalUSDStableCoin, order.usdAmount)
	}

	// Check if totalUSDStableCoin is zero to avoid division by zero
	if totalUSDStableCoin.Cmp(common.Zero) == 0 {
		return nil, fmt.Errorf("total USDC(X) amount is zero, cannot compute ratio")
	}

	// Scale the totals according to their decimals
	scaledTotalCrypto := new(big.Int).Div(totalCrypto, big.NewInt(int64(math.Pow(10, float64(common.AutonityCryptoDecimals)))))
	scaledTotalUSD := new(big.Int).Div(totalUSDStableCoin, big.NewInt(int64(math.Pow(10, float64(usdStableCoinDecimals)))))

	// Calculate the weighted ratio as a fraction
	if scaledTotalUSD.Cmp(common.Zero) == 0 {
		return nil, fmt.Errorf("scaled total USDC amount is zero, cannot compute ratio")
	}

	weightedRatio := new(big.Rat).SetFrac(scaledTotalCrypto, scaledTotalUSD)
	return weightedRatio, nil
}

func (e *AirswapClient) updatePrice(tokenAddress ecommon.Address, price string) {
	e.priceMutex.Lock()
	defer e.priceMutex.Unlock()

	symbol := ATNPairToStableCoin
	if tokenAddress == e.ntnAddress {
		symbol = NTNPairToStableCoin
	}

	e.lastAggregatedPrices[tokenAddress] = common.Price{
		Symbol: symbol,
		Price:  price,
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

	// both ATN-USDC(X) and NTN-USDC(X) price are collected, compute NTN-ATN price.
	if len(prices) == 2 {
		atnPrice := e.lastAggregatedPrices[e.atnAddress]
		ntnPrice := e.lastAggregatedPrices[e.ntnAddress]
		ntnATNPrice, err := common.ComputeDerivedPrice(ntnPrice.Price, atnPrice.Price)
		if err != nil {
			e.logger.Error("cannot compute NTN-ATN price", "error", err.Error())
			return prices, nil
		}
		prices = append(prices, ntnATNPrice)
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

	// start the SwapERC20 event watching for price aggregation of supported symbols.
	go client.StartWatcher()

	adapter := common.NewPlugin(conf, client, version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
