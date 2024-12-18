package common

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/plugins/crypto_uniswap/contracts/factory"
	"autonity-oracle/plugins/crypto_uniswap/contracts/pair"
	"autonity-oracle/types"
	"fmt"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"math"
	"math/big"
)

var (
	Version          = "v0.2.0"
	ATNUSDC          = "ATN-USDC"
	NTNUSDC          = "NTN-USDC"
	supportedSymbols = common.DefaultCryptoSymbols
	NTNTokenAddress  = types.AutonityContractAddress // Autonity protocol contract is the NTN token contract.
)

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

	if len(prices) == 2 {
		ntnATNPrice, err := common.ComputeDerivedPrice(ntnUSDCPrice.Price, atnUSDCPrice.Price)
		if err != nil {
			e.logger.Error("failed to compute NTN-ATN price", "error", err)
			return prices, nil
		}
		prices = append(prices, ntnATNPrice)
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

	// ratio == (cryptoReserve/cryptoDecimals) / (usdcReserve/usdcDecimals)
	//       == (cryptoReserve*usdcDecimals) / (usdcReserve*cryptoDecimals)
	scaledCryptoReserve := new(big.Int).Mul(cryptoReserve, big.NewInt(int64(math.Pow(10, float64(common.USDCDecimals)))))
	scaledUsdcReserve := new(big.Int).Mul(usdcReserve, big.NewInt(int64(math.Pow(10, float64(common.AutonityCryptoDecimals)))))

	// Calculate the exchange ratio as a big.Rat
	ratio := new(big.Rat).SetFrac(scaledCryptoReserve, scaledUsdcReserve)

	// Return the string representation of the ratio
	return ratio.FloatString(common.CryptoToUsdcDecimals), nil
}
