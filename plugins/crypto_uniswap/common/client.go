package common

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"autonity-oracle/plugins/crypto_uniswap/contracts/factory"
	"autonity-oracle/plugins/crypto_uniswap/contracts/pair"
	"autonity-oracle/types"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	ring "github.com/zfjagann/golang-ring"
	"math"
	"math/big"
	"os"
	"sync"
	"time"
)

var (
	orderBookCapacity = 64
	Version           = "v0.2.0"
	ATNUSDC           = "ATN-USDC"
	NTNUSDC           = "NTN-USDC"
	supportedSymbols  = common.DefaultCryptoSymbols
	NTNTokenAddress   = types.AutonityContractAddress // Autonity protocol contract is the NTN token contract.
	initialVolume     = new(big.Int).SetUint64(100)   // The initial volume used before a swap event happens.
)

type Order struct {
	cryptoToUsdcPrice decimal.Decimal // ATN-USDCx or NTN-USDCx ratio.
	volume            *big.Int        // trade volume in usdc of per swap event.
}

type WrappedPair struct {
	pairContract   *pair.Pair
	pairAddress    ecommon.Address
	token0         ecommon.Address
	token1         ecommon.Address
	token0Reserves *big.Int
	token1Reserves *big.Int
}

type UniswapClient struct {
	conf   *config.PluginConfig
	client *ethclient.Client
	logger hclog.Logger

	atnTokenAddress  ecommon.Address
	usdxTokenAddress ecommon.Address

	atnUSDCPairContract *WrappedPair
	ntnUSDCPairContract *WrappedPair

	chAtnSwapEvent  chan *pair.PairSwap
	subAtnSwapEvent event.Subscription

	chNtnSwapEvent  chan *pair.PairSwap
	subNtnSwapEvent event.Subscription

	doneCh   chan struct{}
	ticker   *time.Ticker
	lostSync bool

	atnOrderBooks ring.Ring
	ntnOrderBooks ring.Ring

	priceMutex           sync.RWMutex
	lastAggregatedPrices map[ecommon.Address]common.Price
}

func NewUniswapClient(conf *config.PluginConfig) (*UniswapClient, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	url := conf.Scheme + "://" + conf.Endpoint
	client, err := ethclient.Dial(url)
	if err != nil {
		logger.Error("cannot dial to L1 node", "error", err)
		return nil, err
	}

	// bind uniswap factory contract, it manages the pair contracts in the AMM.
	factoryContract, err := factory.NewFactory(ecommon.HexToAddress(conf.SwapAddress), client)
	if err != nil {
		logger.Error("cannot bind uniswap factory contract", "error", err)
		return nil, err
	}

	atnTokenAddress := ecommon.HexToAddress(conf.ATNTokenAddress)
	usdcTokenAddress := ecommon.HexToAddress(conf.USDCTokenAddress)
	atnUSDCPairContract, err := bindWithPairContract(factoryContract, client, atnTokenAddress, usdcTokenAddress, logger)
	if err != nil {
		logger.Error("bind with ATN USDC pair contract failed", "error", err)
		return nil, err
	}

	ntnUSDCPairContract, err := bindWithPairContract(factoryContract, client, NTNTokenAddress, usdcTokenAddress, logger)
	if err != nil {
		logger.Error("bind with NTN USDC pair contract failed", "error", err)
		return nil, err
	}

	uc := &UniswapClient{
		conf:                 conf,
		client:               client,
		logger:               logger,
		atnUSDCPairContract:  atnUSDCPairContract,
		ntnUSDCPairContract:  ntnUSDCPairContract,
		atnTokenAddress:      atnTokenAddress,
		usdxTokenAddress:     usdcTokenAddress,
		doneCh:               make(chan struct{}),
		ticker:               time.NewTicker(time.Second * 30),
		lastAggregatedPrices: make(map[ecommon.Address]common.Price),
	}

	uc.atnOrderBooks.SetCapacity(orderBookCapacity)
	uc.ntnOrderBooks.SetCapacity(orderBookCapacity)

	if err = uc.EventSubscription(); err != nil {
		return nil, err
	}

	return uc, nil
}

func (e *UniswapClient) EventSubscription() error {
	// subscribe on-chain swap event of atn-usdc.
	chAtnSwapEvent := make(chan *pair.PairSwap)
	subAtnSwapEvent, err := e.atnUSDCPairContract.pairContract.WatchSwap(new(bind.WatchOpts), chAtnSwapEvent, nil, nil)
	if err != nil {
		e.logger.Error("cannot watch ATN USDC pair swap event", "error", err)
		return err
	}

	chNtnSwapEvent := make(chan *pair.PairSwap)
	subNtnSwapEvent, err := e.ntnUSDCPairContract.pairContract.WatchSwap(new(bind.WatchOpts), chNtnSwapEvent, nil, nil)
	if err != nil {
		e.logger.Error("cannot watch NTN USDC pair swap event", "error", err)
		return err
	}
	e.chAtnSwapEvent = chAtnSwapEvent
	e.subAtnSwapEvent = subAtnSwapEvent

	e.chNtnSwapEvent = chNtnSwapEvent
	e.subNtnSwapEvent = subNtnSwapEvent

	return nil
}

func (e *UniswapClient) StartWatcher() {
	for {
		select {
		case <-e.doneCh:
			e.ticker.Stop()
			e.logger.Info("uni-swap events watcher stopped")
			return
		case err := <-e.subAtnSwapEvent.Err():
			if err != nil {
				e.logger.Info("subscription error of ATN-USDCx swap event", "error", err)
				e.handleConnectivityError()
				e.subAtnSwapEvent.Unsubscribe()
			}
		case err := <-e.subNtnSwapEvent.Err():
			if err != nil {
				e.logger.Info("subscription error of NTN-USDCx swap event", "error", err)
				e.handleConnectivityError()
				e.subNtnSwapEvent.Unsubscribe()
			}
		case atnSwapEvent := <-e.chAtnSwapEvent:
			e.logger.Debug("receiving an ATN-USDC swap event", "event", atnSwapEvent)

			if err := e.handleSwapEvent(e.atnTokenAddress, e.atnUSDCPairContract, atnSwapEvent, &e.atnOrderBooks); err != nil {
				e.logger.Error("handle swap event failed", "error", err)
			}
		case ntnSwapEvent := <-e.chNtnSwapEvent:
			e.logger.Debug("receiving a NTN-USDC swap event", "event", ntnSwapEvent)
			if err := e.handleSwapEvent(NTNTokenAddress, e.ntnUSDCPairContract, ntnSwapEvent, &e.ntnOrderBooks); err != nil {
				e.logger.Error("handle swap event failed", "error", err)
			}

		case <-e.ticker.C:
			e.checkHealth()
		}
	}
}

func bindWithPairContract(factoryContract *factory.Factory, client *ethclient.Client, tokenAddress1, tokenAddress2 ecommon.Address, logger hclog.Logger) (*WrappedPair, error) {
	pairAddress, err := factoryContract.GetPair(nil, tokenAddress1, tokenAddress2)
	if err != nil {
		logger.Error("cannot find pair contract from uniswap factory contract", "error", err, "token1", tokenAddress1, "token2", tokenAddress2)
		return nil, err
	}

	if pairAddress == (ecommon.Address{}) {
		logger.Error("cannot find pair contract from uniswap factory contract", "error", err, "token1", tokenAddress1, "token2", tokenAddress2)
		return nil, fmt.Errorf("pair contract from uniswap factory not found, pair: %s, %s", tokenAddress1, tokenAddress2)
	}

	pairContract, err := pair.NewPair(pairAddress, client)
	if err != nil {
		logger.Error("cannot bind pair contract", "error", err, "address", pairAddress)
		return nil, err
	}

	token0, err := pairContract.Token0(nil)
	if err != nil {
		logger.Error("cannot resolve token 0 from liquidity pool", "error", err)
		return nil, err
	}

	token1, err := pairContract.Token1(nil)
	if err != nil {
		logger.Error("cannot resolve token 1 from liquidity pool", "error", err)
		return nil, err
	}

	reserves, err := pairContract.GetReserves(nil)
	if err != nil {
		logger.Error("cannot resolve reserves from liquidity pool", "error", err)
		return nil, err
	}

	return &WrappedPair{
		pairContract:   pairContract,
		pairAddress:    pairAddress,
		token0:         token0,
		token1:         token1,
		token0Reserves: reserves.Reserve0,
		token1Reserves: reserves.Reserve1,
	}, nil
}

func (e *UniswapClient) handleSwapEvent(cryptoToken ecommon.Address, pair *WrappedPair, swap *pair.PairSwap, orderBook *ring.Ring) error {
	var order Order

	if swap.Amount0Out.Cmp(common.Zero) > 0 {
		// Subtract Amount0Out from token0Reserves
		temp0 := new(big.Int).Set(pair.token0Reserves)
		pair.token0Reserves.Set(temp0.Sub(temp0, swap.Amount0Out))

		// Add Amount1In to token1Reserves
		temp1 := new(big.Int).Set(pair.token1Reserves)
		pair.token1Reserves.Set(temp1.Add(temp1, swap.Amount1In))

		// token0 is ATN/NTN, token1 is USDCx
		if pair.token0 == cryptoToken {
			cryptoReserve := new(big.Int).Set(pair.token0Reserves)
			usdcReserve := new(big.Int).Set(pair.token1Reserves)

			price, err := ratio(cryptoReserve, usdcReserve)
			if err != nil {
				return err
			}
			order.cryptoToUsdcPrice = price
			order.volume = swap.Amount1In
		} else {
			// token0 is USDCx, token1 is ATN/NTN
			cryptoReserve := new(big.Int).Set(pair.token1Reserves)
			usdcReserve := new(big.Int).Set(pair.token0Reserves)

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
		temp0 := new(big.Int).Set(pair.token0Reserves)
		temp1 := new(big.Int).Set(pair.token1Reserves)
		pair.token0Reserves.Set(temp0.Add(temp0, swap.Amount0In))
		pair.token1Reserves.Set(temp1.Sub(temp1, swap.Amount1Out))

		// token1 is ATN/NTN, token0 is USDCx
		if pair.token1 == cryptoToken {
			cryptoReserve := new(big.Int).Set(pair.token1Reserves)
			usdcReserve := new(big.Int).Set(pair.token0Reserves)
			price, err := ratio(cryptoReserve, usdcReserve)
			if err != nil {
				return err
			}
			order.cryptoToUsdcPrice = price
			order.volume = swap.Amount0In
		} else {
			// token1 is USDCx, token0 is ATN/NTN
			cryptoReserve := new(big.Int).Set(pair.token0Reserves)
			usdcReserve := new(big.Int).Set(pair.token1Reserves)
			price, err := ratio(cryptoReserve, usdcReserve)
			if err != nil {
				return err
			}
			order.cryptoToUsdcPrice = price
			order.volume = swap.Amount1Out
		}
	}

	aggPrice, volumes, err := aggregatePrice(orderBook, order)
	if err != nil {
		e.logger.Error("aggregate atn-usdcx order book price failed", "error", err)
		return err
	}

	// update the last aggregated price.
	e.updatePrice(cryptoToken, aggPrice.String(), volumes)
	return nil
}

func (e *UniswapClient) updatePrice(tokenAddress ecommon.Address, price string, volumes *big.Int) {
	e.priceMutex.Lock()
	defer e.priceMutex.Unlock()

	symbol := ATNUSDC
	if tokenAddress == NTNTokenAddress {
		symbol = NTNUSDC
	}

	e.lastAggregatedPrices[tokenAddress] = common.Price{
		Symbol: symbol,
		Price:  price,
		Volume: volumes.String(),
	}
}

func aggregatePrice(orderBook *ring.Ring, order Order) (decimal.Decimal, *big.Int, error) {
	orderBook.Enqueue(order)
	recentOrders := orderBook.Values()
	return volumeWeightedPrice(recentOrders)
}

// volumeWeightedPrice calculates the volume-weighted exchange ratio of ATN or NTN to USDC.
func volumeWeightedPrice(orders []interface{}) (decimal.Decimal, *big.Int, error) {
	var vwap decimal.Decimal

	totalValues := new(big.Int)
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

		totalValues.Add(totalValues, valuePerSwap.BigInt())
		totalVol.Add(totalVol, order.volume)
	}

	// Check if totalVol is zero to avoid division by zero
	if totalVol.Cmp(common.Zero) == 0 {
		return vwap, nil, fmt.Errorf("total USDC amount is zero, cannot compute ratio")
	}

	weightedRatio := new(big.Rat).SetFrac(totalValues, totalVol)
	vwap, err := decimal.NewFromString(weightedRatio.FloatString(common.CryptoToUsdcDecimals))
	if err != nil {
		return vwap, nil, err
	}

	return vwap, totalVol, nil
}

func (e *UniswapClient) checkHealth() {
	if e.lostSync {
		err := e.EventSubscription()
		if err != nil {
			e.logger.Info("rebuilding WS connectivity with L1 node", "error", err)
			return
		}

		// re-sync reserves from pools.
		atnUsdcReserves, err := e.atnUSDCPairContract.pairContract.GetReserves(nil)
		if err != nil {
			e.logger.Error("re-sync atn-usdcx pair contract", "error", err)
			return
		}
		e.atnUSDCPairContract.token0Reserves = atnUsdcReserves.Reserve0
		e.atnUSDCPairContract.token1Reserves = atnUsdcReserves.Reserve1

		ntnUsdcReservers, err := e.ntnUSDCPairContract.pairContract.GetReserves(nil)
		if err != nil {
			e.logger.Error("re-sync ntn-usdcx pair contract", "error", err)
			return
		}
		e.ntnUSDCPairContract.token0Reserves = ntnUsdcReservers.Reserve0
		e.ntnUSDCPairContract.token1Reserves = ntnUsdcReservers.Reserve1

		e.lostSync = false
		return
	}

	e.logger.Debug("checking heart beat", "alive", !e.lostSync)
}

func (e *UniswapClient) handleConnectivityError() {
	e.lostSync = true
}

func (e *UniswapClient) KeyRequired() bool {
	return false
}

func (e *UniswapClient) fetchPrice(pair *WrappedPair, symbol string) (common.Price, error) {
	var price common.Price
	reserves, err := pair.pairContract.GetReserves(nil)
	if err != nil {
		e.logger.Error("cannot get reserves from uni-swap liquidity pool", "error", err)
		return price, err
	}

	if reserves.Reserve0 == nil || reserves.Reserve1 == nil {
		e.logger.Error("get nil reserves from contracts liquidity pool")
		return price, fmt.Errorf("nil reserves get from liquidity pool")
	}

	var cryptoReserve *big.Int
	var usdcReserve *big.Int

	if pair.token0 == ecommon.HexToAddress(e.conf.ATNTokenAddress) || pair.token0 == NTNTokenAddress {
		// ATN or NTN is token0, compute ATN-USDC or NTN-USDC ratio with reserves0 and reserves1.
		cryptoReserve = reserves.Reserve0
		usdcReserve = reserves.Reserve1
	} else {
		// ATN or NTN is token1, compute ATN-USDC or NTN-USDC ratio with reserves0 and reserves1.
		cryptoReserve = reserves.Reserve1
		usdcReserve = reserves.Reserve0
	}

	p, err := ratio(cryptoReserve, usdcReserve)
	if err != nil {
		e.logger.Error("cannot compute exchange ratio of ATN-USDC", "error", err)
		return price, err
	}

	price.Symbol = symbol
	price.Price = p.String()
	price.Volume = initialVolume.String()
	return price, nil
}

func (e *UniswapClient) FetchPrice(_ []string) (common.Prices, error) {
	var prices common.Prices

	atnUSDCPrice, err := e.lastAggregatedPrice(e.atnTokenAddress)
	if err == nil {
		prices = append(prices, atnUSDCPrice)
	} else {
		e.logger.Debug("no aggregated ATN-USDCx price yet, going to fetch from pool", "error", err)
		// no swap event accumulated, compute price from current pool reserves.
		price, er := e.fetchPrice(e.atnUSDCPairContract, ATNUSDC)
		if er == nil {
			prices = append(prices, price)
		} else {
			e.logger.Error("failed to fetch ATN-USDC price", "error", er)
		}
	}

	ntnUSDCPrice, err := e.lastAggregatedPrice(NTNTokenAddress)
	if err == nil {
		prices = append(prices, ntnUSDCPrice)
	} else {
		e.logger.Debug("no aggregated NTN-USDCx price yet, going to fetch from pool", "error", err)
		// no swap event accumulated, compute price from current pool reserves.
		price, er := e.fetchPrice(e.ntnUSDCPairContract, NTNUSDC)
		if er == nil {
			prices = append(prices, price)
		} else {
			e.logger.Error("failed to fetch NTN-USDC price", "error", er)
		}
	}

	if len(prices) == 2 {
		ntnATNPrice, err := common.ComputeDerivedPrice(prices[1].Price, prices[0].Price)
		if err != nil {
			e.logger.Error("failed to compute NTN-ATN price", "error", err)
			return prices, nil
		}
		ntnATNPrice.Volume = prices[0].Volume
		prices = append(prices, ntnATNPrice)
	}

	return prices, nil
}

func (e *UniswapClient) lastAggregatedPrice(tokenAddress ecommon.Address) (common.Price, error) {
	e.priceMutex.RLock()
	defer e.priceMutex.RUnlock()

	var price common.Price
	latestPrice, ok := e.lastAggregatedPrices[tokenAddress]
	if !ok {
		return price, fmt.Errorf("no available price yet for token %s", tokenAddress.Hex())
	}

	return latestPrice, nil
}

func (e *UniswapClient) AvailableSymbols() ([]string, error) {
	return supportedSymbols, nil
}

func (e *UniswapClient) Close() {
	if e.client != nil {
		e.client.Close()
	}
	e.subAtnSwapEvent.Unsubscribe()
	e.subNtnSwapEvent.Unsubscribe()
	e.doneCh <- struct{}{}
}

func ratio(cryptoReserve, usdcReserve *big.Int) (decimal.Decimal, error) {
	var r decimal.Decimal

	if usdcReserve.Cmp(common.Zero) == 0 {
		return r, fmt.Errorf("usdcReserve is zero, cannot compute exchange ratio")
	}

	if cryptoReserve.Cmp(common.Zero) < 0 || usdcReserve.Cmp(common.Zero) < 0 {
		return r, fmt.Errorf("negative reserve value")
	}

	// ratio == (cryptoReserve/cryptoDecimals) / (usdcReserve/usdcDecimals)
	//       == (cryptoReserve*usdcDecimals) / (usdcReserve*cryptoDecimals)
	scaledCryptoReserve := new(big.Int).Mul(cryptoReserve, big.NewInt(int64(math.Pow(10, float64(common.USDCDecimals)))))
	scaledUsdcReserve := new(big.Int).Mul(usdcReserve, big.NewInt(int64(math.Pow(10, float64(common.AutonityCryptoDecimals)))))

	// Calculate the exchange ratio as a big.Rat
	price := new(big.Rat).SetFrac(scaledCryptoReserve, scaledUsdcReserve)

	r, err := decimal.NewFromString(price.FloatString(common.CryptoToUsdcDecimals))
	if err != nil {
		return r, err
	}

	return r, nil
}
