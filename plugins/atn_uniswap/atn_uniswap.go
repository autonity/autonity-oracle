package main

import (
	"autonity-oracle/plugins/atn_uniswap/uniswap/factory"
	"autonity-oracle/plugins/atn_uniswap/uniswap/pair"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"fmt"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"math/big"
	"os"
)

var (
	version          = "v0.0.1"
	ATNUSDC          = "ATN-USDC"
	supportedSymbols = []string{ATNUSDC}
)

// todo: Jason, to keep a high availability of L1 node, who will have the ownership of the operation of this node? Or,
//
//	shall we config multiple L1 node endpoints for accessibility?
var defaultConfig = types.PluginConfig{
	Name:               "atn_uniswap",
	Key:                "",
	Scheme:             "https", // todo: set the protocol to connect to L1 blockchain node. ws or https
	Endpoint:           "",      // todo: set the host name or IP address and port for the service endpoint.
	Timeout:            10,      // 10s
	DataUpdateInterval: 30,      // todo: resolve the interval by according to the rate limit policy of the service end point.
	BaseTokenAddress:   "0x",    // todo: set the wrapped ATN erc20 token address
	QuoteTokenAddress:  "0x",    // todo: set the USDC erc20 token address
	SwapAddress:        "0x",    // todo: set the uniswap factory contract address
}

type EvmClient struct {
	conf         *types.PluginConfig
	client       *ethclient.Client
	logger       hclog.Logger
	pairContract *pair.Pair
	token0       ecommon.Address
	token1       ecommon.Address
}

func NewEVMClient(conf *types.PluginConfig, logger hclog.Logger) (*EvmClient, error) {
	url := conf.Scheme + "://" + conf.Endpoint
	client, err := ethclient.Dial(url)
	if err != nil {
		logger.Error("cannot dial to L1 node", "error", err)
		return nil, err
	}

	factoryContract, err := factory.NewFactory(ecommon.HexToAddress(conf.SwapAddress), client)
	if err != nil {
		logger.Error("cannot bind uniswap factory contract", "error", err)
		return nil, err
	}

	pairAddress, err := factoryContract.GetPair(nil, ecommon.HexToAddress(conf.BaseTokenAddress), ecommon.HexToAddress(conf.QuoteTokenAddress))
	if err != nil {
		logger.Error("cannot find ATN-USDC liquidity pool from uniswap factory contract", "error", err)
		return nil, err
	}

	if pairAddress == (ecommon.Address{}) {
		logger.Error("cannot find ATN-USDC liquidity pool from uniswap factory contract", "address", pairAddress)
		return nil, fmt.Errorf("ATN-USDC liquidity pair address is empty")
	}

	pairContract, err := pair.NewPair(pairAddress, client)
	if err != nil {
		logger.Error("cannot bind ATN-USDC pair contract", "error", err)
		return nil, err
	}

	token0, err := pairContract.Token0(nil)
	if err != nil {
		logger.Error("cannot resolve token 0 from liquidity uniswap", "error", err)
		return nil, err
	}

	token1, err := pairContract.Token1(nil)
	if err != nil {
		logger.Error("cannot resolve token 1 from liquidity uniswap", "error", err)
		return nil, err
	}

	return &EvmClient{conf: conf, client: client, logger: logger, pairContract: pairContract, token0: token0, token1: token1}, nil
}

func (e *EvmClient) KeyRequired() bool {
	return false
}

func (e *EvmClient) FetchPrice(_ []string) (common.Prices, error) {
	var prices common.Prices
	reserves, err := e.pairContract.GetReserves(nil)
	if err != nil {
		e.logger.Error("cannot get reserves from uni-swap liquidity pool", "error", err)
		return nil, err
	}

	if reserves.Reserve0 == nil || reserves.Reserve1 == nil {
		e.logger.Error("get nil reserves from uniswap liquidity pool")
		return nil, fmt.Errorf("nil reserves get from liquidity pool")
	}

	var atnReserve *big.Int
	var usdcReserve *big.Int
	if e.token0 == ecommon.HexToAddress(e.conf.BaseTokenAddress) {
		// ATN is token0, compute ATN-USDC ratio with reserves0 and reserves1.
		atnReserve = reserves.Reserve0
		usdcReserve = reserves.Reserve1
	} else {
		// ATN is token1, compute ATN-USDC ratio with reserves0 and reserves1.
		atnReserve = reserves.Reserve1
		usdcReserve = reserves.Reserve0
	}

	p, err := ComputeExchangeRatio(atnReserve, usdcReserve)
	if err != nil {
		e.logger.Error("cannot compute exchange ratio of ATN-USDC", "error", err)
		return nil, err
	}

	price := common.Price{
		Symbol: ATNUSDC,
		Price:  p.String(),
	}
	prices = append(prices, price)
	return prices, nil
}

func (e *EvmClient) AvailableSymbols() ([]string, error) {
	return supportedSymbols, nil
}

func (e *EvmClient) Close() {
	if e.client != nil {
		e.client.Close()
	}
}

// ComputeExchangeRatio calculates the exchange ratio based on current reserves
func ComputeExchangeRatio(reserve0, reserve1 *big.Int) (*big.Float, error) {
	if reserve1.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("reserve1 is zero, cannot compute exchange ratio")
	}

	ratio := new(big.Float).Quo(new(big.Float).SetInt(reserve0), new(big.Float).SetInt(reserve1))
	return ratio, nil
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	client, err := NewEVMClient(conf, logger)
	if err != nil {
		return
	}

	adapter := common.NewPlugin(conf, client, version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
