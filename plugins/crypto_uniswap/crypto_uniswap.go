package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/plugins/crypto_uniswap/uniswap/factory"
	"autonity-oracle/plugins/crypto_uniswap/uniswap/pair"
	"autonity-oracle/types"
	"fmt"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"math"
	"math/big"
	"os"
)

var (
	version          = "v0.0.1"
	ATNUSDC          = "ATN-USDC"
	NTNUSDC          = "NTN-USDC"
	supportedSymbols = []string{ATNUSDC, NTNUSDC}
	NTNTokenAddress  = types.AutonityContractAddress // Autonity protocol contract is the NTN token contract.
)

// Todo: require a config PR to finalize below configuration for the target autonity network.
var defaultConfig = types.PluginConfig{
	Name:               "crypto_uniswap",
	Scheme:             "wss",                             // todo: update this on redeployment of infra
	Endpoint:           "rpc1.piccadilly.autonity.org/ws", // todo: update this on redeployment of infra
	Timeout:            10,                                // 10s
	DataUpdateInterval: 30,                                // 30s
	// todo: update below protocol contract addresses on redeployment of protocols.
	ATNTokenAddress:  "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2", // Wrapped ATN ERC20 contract address on Autonity blockchain.
	USDCTokenAddress: "0x3a60C03a86eEAe30501ce1af04a6C04Cf0188700", // USDC ERC20 contract address on Autonity blockchain.
	SwapAddress:      "0x218F76e357594C82Cc29A88B90dd67b180827c88",
}

type WrappedPair struct {
	pairContract *pair.Pair
	token0       ecommon.Address
	token1       ecommon.Address
}

type UniswapClient struct {
	conf                *types.PluginConfig
	client              *ethclient.Client
	logger              hclog.Logger
	atnUSDCPairContract *WrappedPair
	ntnUSDCPairContract *WrappedPair
}

func NewUniswapClient(conf *types.PluginConfig, logger hclog.Logger) (*UniswapClient, error) {
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

	return &UniswapClient{conf: conf, client: client, logger: logger, atnUSDCPairContract: atnUSDCPairContract, ntnUSDCPairContract: ntnUSDCPairContract}, nil
}

func bindWithPairContract(factoryContract *factory.Factory, client *ethclient.Client, tokenAddress1, tokenAddress2 ecommon.Address, logger hclog.Logger) (*WrappedPair, error) {
	pairAddress, err := factoryContract.GetPair(nil, tokenAddress1, tokenAddress2)
	if err != nil {
		logger.Error("cannot find pair contract from uniswap factory contract", "error", err, "token1", tokenAddress1, "token2", tokenAddress2)
		return nil, err
	}

	if pairAddress == (ecommon.Address{}) {
		logger.Error("cannot find pair contract from uniswap factory contract", "error", err, "token1", tokenAddress1, "token2", tokenAddress2)
		return nil, fmt.Errorf("ATN-USDC liquidity pair address is empty")
	}

	pairContract, err := pair.NewPair(pairAddress, client)
	if err != nil {
		logger.Error("cannot bind pair contract", "error", err, "address", pairAddress)
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

	return &WrappedPair{
		pairContract: pairContract,
		token0:       token0,
		token1:       token1,
	}, nil
}

func (e *UniswapClient) KeyRequired() bool {
	return false
}

func (e *UniswapClient) FetchPrice(_ []string) (common.Prices, error) {
	var prices common.Prices

	atnUSDCPrice, err := e.fetchPrice(e.atnUSDCPairContract, ATNUSDC)
	if err == nil {
		prices = append(prices, atnUSDCPrice)
	} else {
		e.logger.Error("failed to fetch ATN-USDC price", "error", err)
	}

	ntnUSDCPrice, err := e.fetchPrice(e.ntnUSDCPairContract, NTNUSDC)
	if err == nil {
		prices = append(prices, ntnUSDCPrice)
	} else {
		e.logger.Error("failed to fetch NTN-USDC price", "error", err)
	}

	return prices, nil
}

func (e *UniswapClient) fetchPrice(pair *WrappedPair, symbol string) (common.Price, error) {
	var price common.Price
	reserves, err := pair.pairContract.GetReserves(nil)
	if err != nil {
		e.logger.Error("cannot get reserves from uni-swap liquidity pool", "error", err)
		return price, err
	}

	if reserves.Reserve0 == nil || reserves.Reserve1 == nil {
		e.logger.Error("get nil reserves from uniswap liquidity pool")
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

	p, err := ComputeExchangeRatio(cryptoReserve, usdcReserve)
	if err != nil {
		e.logger.Error("cannot compute exchange ratio of ATN-USDC", "error", err)
		return price, err
	}

	price.Symbol = symbol
	price.Price = p
	return price, nil
}

func (e *UniswapClient) AvailableSymbols() ([]string, error) {
	return supportedSymbols, nil
}

func (e *UniswapClient) Close() {
	if e.client != nil {
		e.client.Close()
	}
}

// ComputeExchangeRatio calculates the exchange ratio from ATN or NTN to USDC
func ComputeExchangeRatio(cryptoReserve, usdcReserve *big.Int) (string, error) {
	if usdcReserve.Cmp(common.Zero) == 0 {
		return "", fmt.Errorf("usdcReserve is zero, cannot compute exchange ratio")
	}

	// Scale the reserves to the same precision
	scaledCryptoReserve := new(big.Int).Div(cryptoReserve, big.NewInt(int64(math.Pow(10, float64(common.AutonityCryptoDecimals)))))
	scaledUsdcReserve := new(big.Int).Div(usdcReserve, big.NewInt(int64(math.Pow(10, float64(common.USDCDecimals)))))

	if scaledUsdcReserve.Cmp(common.Zero) == 0 {
		return "", fmt.Errorf("scaledUsdcReserve is zero, cannot compute exchange ratio")
	}

	// Calculate the exchange ratio as a big.Rat
	ratio := new(big.Rat).SetFrac(scaledCryptoReserve, scaledUsdcReserve)

	// Return the string representation of the ratio
	return ratio.FloatString(common.CryptoToUsdcDecimals), nil
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   conf.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})

	client, err := NewUniswapClient(conf, logger)
	if err != nil {
		return
	}

	adapter := common.NewPlugin(conf, client, version)
	defer adapter.Close()
	common.PluginServe(adapter)
}
