package common

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"autonity-oracle/plugins/crypto_uniswap/contracts/factory"
	"autonity-oracle/plugins/crypto_uniswap/contracts/pair"
	"autonity-oracle/types"
	"fmt"
	"math"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	ring "github.com/zfjagann/golang-ring"
)

var (
	orderBookCapacity = 64
	Version           = "v0.2.7"
	supportedSymbols  = common.DefaultCryptoSymbols
)

type Order struct {
	cryptoToUsdcPrice decimal.Decimal // ATN-USDCx or NTN-USDCx ratio.
	volume            *big.Int        // trade volume in usdc of per swap event.
}

type UniswapClient struct {
	conf   *config.PluginConfig
	logger hclog.Logger

	atnUSDCPairContract *WrappedPair
	ntnUSDCPairContract *WrappedPair
}

func NewUniswapClient(conf *config.PluginConfig) (*UniswapClient, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	// just load config and logger for uniswap client, as the crypto-pair markets can be
	// resolved during runtime now on-demand.
	return &UniswapClient{
		conf:   conf,
		logger: logger,
	}, nil
}

func (e *UniswapClient) KeyRequired() bool {
	return false
}

func (e *UniswapClient) AvailableSymbols() ([]string, error) {
	return supportedSymbols, nil
}

func (e *UniswapClient) tryFetch(pair *WrappedPair, symbol string) (common.Price, error) {
	// if the pair haven't been bind with the marketplace, try to bind it.
	if pair == nil {
		url := e.conf.Scheme + "://" + e.conf.Endpoint
		usdcTokenAddress := ecommon.HexToAddress(e.conf.USDCTokenAddress)
		factoryAddress := ecommon.HexToAddress(e.conf.SwapAddress)

		if symbol == common.ATNUSDCSymbol {
			atnTokenAddress := ecommon.HexToAddress(e.conf.ATNTokenAddress)
			atnUsdcPair, err := NewWrappedPair(symbol, atnTokenAddress, usdcTokenAddress, factoryAddress, url, e.logger)
			if err != nil {
				return common.Price{}, err
			}
			e.atnUSDCPairContract = atnUsdcPair
			return atnUsdcPair.aggregatedPrice()
		}

		// pair for NTN-USDC comes here
		ntnTokenAddress := ecommon.HexToAddress(e.conf.NTNTokenAddress)
		ntnUsdcPair, err := NewWrappedPair(symbol, ntnTokenAddress, usdcTokenAddress, factoryAddress, url, e.logger)
		if err != nil {
			return common.Price{}, err
		}
		e.ntnUSDCPairContract = ntnUsdcPair
		return ntnUsdcPair.aggregatedPrice()
	}

	// pair contract already bind, try to get recent aggregated price.
	return pair.aggregatedPrice()
}

// FetchPrice fetch the price of the supported symbols of this plugin.
func (e *UniswapClient) FetchPrice(symbols []string) (common.Prices, error) {
	var prices common.Prices

	var atnUSDCPrice common.Price
	var ntnUSDCPrice common.Price
	var err error
	for _, symbol := range symbols {
		if symbol == common.ATNUSDCSymbol {
			atnUSDCPrice, err = e.tryFetch(e.atnUSDCPairContract, symbol)
			if err != nil {
				e.logger.Info("fetch price", "symbol", symbol, "err", err)
				continue
			}

			prices = append(prices, atnUSDCPrice)
			continue
		}

		if symbol == common.NTNUSDCSymbol {
			ntnUSDCPrice, err = e.tryFetch(e.ntnUSDCPairContract, symbol)
			if err != nil {
				e.logger.Info("fetch price", "symbol", symbol, "err", err)
				continue
			}
			prices = append(prices, ntnUSDCPrice)
			continue
		}
	}

	if len(prices) == 2 {
		ntnATNPrice, err := common.ComputeDerivedPrice(ntnUSDCPrice.Price, atnUSDCPrice.Price)
		if err != nil {
			e.logger.Error("failed to compute NTN-ATN price", "error", err)
			return prices, nil
		}
		ntnATNPrice.Volume = atnUSDCPrice.Volume
		prices = append(prices, ntnATNPrice)
	}

	return prices, nil
}

func (e *UniswapClient) Close() {
	if e.ntnUSDCPairContract != nil {
		e.ntnUSDCPairContract.Close()
	}

	if e.atnUSDCPairContract != nil {
		e.atnUSDCPairContract.Close()
	}
}

type WrappedPair struct {
	logger              hclog.Logger
	symbol              string
	baseTokenAddress    ecommon.Address // the base token address, it is ATN or NTN.
	client              *ethclient.Client
	chSwapEvent         chan *pair.PairSwap // chan of the swap event of the tracked pair
	subSwapEvent        event.Subscription  // subscription of the swap event of the tracked pair.
	doneCh              chan struct{}
	ticker              *time.Ticker
	lostSync            bool
	orderBooks          ring.Ring
	priceMutex          sync.RWMutex
	lastAggregatedPrice *common.Price

	pairContract   *pair.Pair
	pairAddress    ecommon.Address
	token0         ecommon.Address
	token1         ecommon.Address
	token0Reserves *big.Int
	token1Reserves *big.Int
}

func NewWrappedPair(symbol string, baseTokenAddress ecommon.Address, quoteTokenAddress ecommon.Address, factoryAddress ecommon.Address,
	url string, logger hclog.Logger) (*WrappedPair, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		logger.Error("cannot dial to L1 validator node", "error", err)
		return nil, err
	}
	// bind uniswap factory contract, it manages the pair contracts in the AMM.
	factoryContract, err := factory.NewFactory(factoryAddress, client)
	if err != nil {
		logger.Info("connect to uniswap factory contract", "error", err, "address", factoryAddress.String())
		client.Close()
		return nil, err
	}

	pairAddress, err := factoryContract.GetPair(nil, baseTokenAddress, quoteTokenAddress)
	if err != nil {
		logger.Info("connect to token pair contract in uniswap factory", "error", err, "token1", baseTokenAddress, "token2", quoteTokenAddress)
		client.Close()
		return nil, err
	}

	if pairAddress == (ecommon.Address{}) {
		logger.Info("pair contract is not created yet in uniswap factory", "token1", baseTokenAddress, "token2", quoteTokenAddress)
		client.Close()
		return nil, fmt.Errorf("pair contract is not created yet, pair: %s, %s", baseTokenAddress, quoteTokenAddress)
	}

	pairContract, err := pair.NewPair(pairAddress, client)
	if err != nil {
		logger.Error("bind pair contract", "error", err, "address", pairAddress)
		client.Close()
		return nil, err
	}

	token0, err := pairContract.Token0(nil)
	if err != nil {
		logger.Error("cannot resolve token 0 from liquidity pool", "error", err)
		client.Close()
		return nil, err
	}

	token1, err := pairContract.Token1(nil)
	if err != nil {
		logger.Error("cannot resolve token 1 from liquidity pool", "error", err)
		client.Close()
		return nil, err
	}

	reserves, err := pairContract.GetReserves(nil)
	if err != nil {
		logger.Error("cannot resolve reserves from liquidity pool", "error", err)
		client.Close()
		return nil, err
	}

	wPair := &WrappedPair{
		logger:           logger,
		symbol:           symbol,
		baseTokenAddress: baseTokenAddress,
		client:           client,
		doneCh:           make(chan struct{}),
		ticker:           time.NewTicker(time.Second * 1), // 1s ticker used to repair L1 connectivity if it was disconnected.

		pairContract:   pairContract,
		pairAddress:    pairAddress,
		token0:         token0,
		token1:         token1,
		token0Reserves: reserves.Reserve0,
		token1Reserves: reserves.Reserve1,
	}
	wPair.orderBooks.SetCapacity(orderBookCapacity)

	if err = wPair.EventSubscription(); err != nil {
		wPair.logger.Error("failed to subscribe to swap event", "error", err, "base", baseTokenAddress)
		client.Close()
		return nil, err
	}

	go wPair.StartWatcher()

	return wPair, nil
}

func (e *WrappedPair) EventSubscription() error {
	// subscribe on-chain swap event.
	chSwapEvent := make(chan *pair.PairSwap)
	subSwapEvent, err := e.pairContract.WatchSwap(new(bind.WatchOpts), chSwapEvent, nil, nil)
	if err != nil {
		e.logger.Error("cannot watch pair swap event", "error", err)
		return err
	}

	e.chSwapEvent = chSwapEvent
	e.subSwapEvent = subSwapEvent

	return nil
}

func (e *WrappedPair) StartWatcher() {
	for {
		select {
		case <-e.doneCh:
			e.ticker.Stop()
			e.logger.Info("uni-swap events watcher stopped")
			return
		case err := <-e.subSwapEvent.Err():
			if err != nil {
				e.logger.Info("subscription error of swap event", "error", err)
				e.handleConnectivityError()
				e.subSwapEvent.Unsubscribe()
			}
		case swapEvent := <-e.chSwapEvent:
			e.logger.Debug("receiving a swap event", "event", swapEvent)
			if err := e.handleSwapEvent(swapEvent); err != nil {
				e.logger.Error("handle swap event failed", "error", err)
			}

		case <-e.ticker.C:
			e.checkHealth()
		}
	}
}

func (e *WrappedPair) Close() {
	if e.client != nil {
		e.client.Close()
	}
	e.subSwapEvent.Unsubscribe()
	e.doneCh <- struct{}{}
}

func (e *WrappedPair) handleSwapEvent(swap *pair.PairSwap) error {
	var order Order

	if swap.Amount0Out.Cmp(common.Zero) > 0 {
		// Subtract Amount0Out from token0Reserves
		temp0 := new(big.Int).Set(e.token0Reserves)
		e.token0Reserves.Set(temp0.Sub(temp0, swap.Amount0Out))

		// Add Amount1In to token1Reserves
		temp1 := new(big.Int).Set(e.token1Reserves)
		e.token1Reserves.Set(temp1.Add(temp1, swap.Amount1In))

		// token0 is ATN/NTN, token1 is USDCx
		if e.token0 == e.baseTokenAddress {
			cryptoReserve := new(big.Int).Set(e.token0Reserves)
			usdcReserve := new(big.Int).Set(e.token1Reserves)

			price, err := ratio(cryptoReserve, usdcReserve)
			if err != nil {
				return err
			}
			order.cryptoToUsdcPrice = price
			order.volume = swap.Amount1In
		} else {
			// token0 is USDCx, token1 is ATN/NTN
			cryptoReserve := new(big.Int).Set(e.token1Reserves)
			usdcReserve := new(big.Int).Set(e.token0Reserves)

			price, err := ratio(cryptoReserve, usdcReserve)
			if err != nil {
				return err
			}
			order.cryptoToUsdcPrice = price
			order.volume = swap.Amount0Out
		}
	}

	if swap.Amount1Out.Cmp(common.Zero) > 0 {
		// update reserves.
		temp0 := new(big.Int).Set(e.token0Reserves)
		temp1 := new(big.Int).Set(e.token1Reserves)
		e.token0Reserves.Set(temp0.Add(temp0, swap.Amount0In))
		e.token1Reserves.Set(temp1.Sub(temp1, swap.Amount1Out))

		// token1 is ATN/NTN, token0 is USDCx
		if e.token1 == e.baseTokenAddress {
			cryptoReserve := new(big.Int).Set(e.token1Reserves)
			usdcReserve := new(big.Int).Set(e.token0Reserves)
			price, err := ratio(cryptoReserve, usdcReserve)
			if err != nil {
				return err
			}
			order.cryptoToUsdcPrice = price
			order.volume = swap.Amount0In
		} else {
			// token1 is USDCx, token0 is ATN/NTN
			cryptoReserve := new(big.Int).Set(e.token0Reserves)
			usdcReserve := new(big.Int).Set(e.token1Reserves)
			price, err := ratio(cryptoReserve, usdcReserve)
			if err != nil {
				return err
			}
			order.cryptoToUsdcPrice = price
			order.volume = swap.Amount1Out
		}
	}

	aggPrice, volumes, err := aggregatePrice(&e.orderBooks, order)
	if err != nil {
		e.logger.Error("aggregate atn-usdcx order book price failed", "error", err)
		return err
	}

	// update the last aggregated price.
	e.logger.Debug("newly aggregated price", "price", aggPrice)
	e.updatePrice(aggPrice.String(), volumes)
	return nil
}

func (e *WrappedPair) updatePrice(price string, volumes *big.Int) {
	e.priceMutex.Lock()
	defer e.priceMutex.Unlock()

	e.lastAggregatedPrice = &common.Price{
		Symbol: e.symbol,
		Price:  price,
		Volume: volumes.String(),
	}
}

func (e *WrappedPair) checkHealth() {
	if e.lostSync {
		err := e.EventSubscription()
		if err != nil {
			e.logger.Info("rebuilding WS connectivity with L1 node", "error", err)
			return
		}

		// re-sync reserves from pools.
		reserves, err := e.pairContract.GetReserves(nil)
		if err != nil {
			e.logger.Error("re-sync pair contract", "error", err)
			return
		}
		e.token0Reserves = reserves.Reserve0
		e.token1Reserves = reserves.Reserve1

		e.lostSync = false
		return
	}

	e.logger.Debug("checking heart beat", "alive", !e.lostSync)
}

func (e *WrappedPair) handleConnectivityError() {
	e.lostSync = true
}

func (e *WrappedPair) instantPrice() (common.Price, error) {
	var price common.Price

	var cryptoReserve *big.Int
	var usdcReserve *big.Int

	if e.token0 == e.baseTokenAddress {
		// ATN or NTN is token0, compute ATN-USDC or NTN-USDC ratio with reserves0 and reserves1.
		cryptoReserve = e.token0Reserves
		usdcReserve = e.token1Reserves
	} else {
		// ATN or NTN is token1, compute ATN-USDC or NTN-USDC ratio with reserves0 and reserves1.
		cryptoReserve = e.token1Reserves
		usdcReserve = e.token0Reserves
	}

	p, err := ratio(cryptoReserve, usdcReserve)
	if err != nil {
		e.logger.Error("cannot compute exchange ratio of ATN-USDC", "error", err)
		return price, err
	}

	price.Symbol = e.symbol
	price.Price = p.String()
	price.Volume = types.DefaultVolume.String()
	e.logger.Debug("instant price", "price", price)
	return price, nil
}

func (e *WrappedPair) aggregatedPrice() (common.Price, error) {
	e.priceMutex.RLock()
	defer e.priceMutex.RUnlock()

	if e.lastAggregatedPrice == nil {
		return e.instantPrice()
	}

	return *e.lastAggregatedPrice, nil
}

func ratio(cryptoReserve, usdcReserve *big.Int) (decimal.Decimal, error) {
	var r decimal.Decimal

	if cryptoReserve.Cmp(common.Zero) <= 0 || usdcReserve.Cmp(common.Zero) <= 0 {
		return r, fmt.Errorf("reserve value <= 0, skip price computing")
	}

	// ratio == (usdcReserve/usdcDecimals) / (cryptoReserve/cryptoDecimals)
	//       == (usdcReserve*cryptoDecimals) / (cryptoReserve*usdcDecimals)
	scaledCryptoReserve := new(big.Int).Mul(cryptoReserve, big.NewInt(int64(math.Pow(10, float64(common.USDCDecimals)))))
	scaledUsdcReserve := new(big.Int).Mul(usdcReserve, big.NewInt(int64(math.Pow(10, float64(common.AutonityCryptoDecimals)))))

	// Calculate the exchange ratio as a big.Rat
	price := new(big.Rat).SetFrac(scaledUsdcReserve, scaledCryptoReserve)
	r, err := decimal.NewFromString(price.FloatString(common.CryptoToUsdcDecimals))
	if err != nil {
		return r, err
	}

	return r, nil
}

func aggregatePrice(orderBook *ring.Ring, order Order) (decimal.Decimal, *big.Int, error) {
	orderBook.Enqueue(order)
	recentOrders := orderBook.Values()
	// nothing to aggregate
	if len(recentOrders) == 1 {
		return order.cryptoToUsdcPrice, order.volume, nil
	}

	return volumeWeightedPrice(recentOrders)
}

// volumeWeightedPrice calculates the volume-weighted exchange ratio of ATN or NTN to USDC.
func volumeWeightedPrice(orders []interface{}) (decimal.Decimal, *big.Int, error) {
	var vwap decimal.Decimal

	var totalValuesDecimal decimal.Decimal
	totalVol := new(big.Int)

	// Iterate through the orders to sum up the amounts
	for _, orderInterface := range orders {
		// Type assert to Order
		order, ok := orderInterface.(Order)
		if !ok {
			return vwap, nil, fmt.Errorf("invalid order type")
		}

		vol := decimal.NewFromBigInt(order.volume, 0)
		valuePerSwap := order.cryptoToUsdcPrice.Mul(vol)

		totalValuesDecimal = totalValuesDecimal.Add(valuePerSwap)
		totalVol.Add(totalVol, order.volume)
	}

	// Check if totalVol is zero to avoid division by zero
	if totalVol.Cmp(common.Zero) == 0 {
		return vwap, nil, fmt.Errorf("total USDC amount is zero, cannot compute ratio")
	}

	weightedRatio := totalValuesDecimal.Div(decimal.NewFromBigInt(totalVol, 0))
	vwap, err := decimal.NewFromString(weightedRatio.String())
	if err != nil {
		return vwap, nil, err
	}

	return vwap, totalVol, nil
}
