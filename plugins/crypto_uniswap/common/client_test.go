package common

import (
	config2 "autonity-oracle/config"
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewUniswapClientWithFullMarkets(t *testing.T) {

	// using current piccadilly protocol configs.
	config := config2.PluginConfig{
		Name:               "crypto_uniswap",
		Scheme:             "wss",
		Endpoint:           "rpc-internal-1.piccadilly.autonity.org/ws",
		Timeout:            10,
		DataUpdateInterval: 30,
		// set the ATN token address with an un exist value, to let the market cannot be discovered.
		ATNTokenAddress:  "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2",
		NTNTokenAddress:  types.AutonityContractAddress.Hex(),
		USDCTokenAddress: "0xB855D5e83363A4494e09f0Bb3152A70d3f161940",
		SwapAddress:      "0x218F76e357594C82Cc29A88B90dd67b180827c88",
	}

	client, err := NewUniswapClient(&config)
	require.NoError(t, err)

	defer client.Close()

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		prices, err := client.FetchPrice(supportedSymbols)
		require.NoError(t, err)
		require.Equal(t, 3, len(prices))
		for _, price := range prices {
			_, err := decimal.NewFromString(price.Price)
			require.NoError(t, err)
		}
		t.Log(prices)
	}
}

func TestNewUniswapClientWithATNMarket(t *testing.T) {

	// using current piccadilly protocol configs.
	config := config2.PluginConfig{
		Name:               "crypto_uniswap",
		Scheme:             "wss",
		Endpoint:           "rpc-internal-1.piccadilly.autonity.org/ws",
		Timeout:            10,
		DataUpdateInterval: 30,
		ATNTokenAddress:    "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2",
		// set the NTN token address with an un exist value, to let the market cannot be discovered.
		NTNTokenAddress:  "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1df",
		USDCTokenAddress: "0xB855D5e83363A4494e09f0Bb3152A70d3f161940",
		SwapAddress:      "0x218F76e357594C82Cc29A88B90dd67b180827c88",
	}

	client, err := NewUniswapClient(&config)
	require.NoError(t, err)

	defer client.Close()

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		prices, err := client.FetchPrice([]string{common.NTNUSDCSymbol, common.ATNUSDCSymbol})
		require.NoError(t, err)
		require.Equal(t, 1, len(prices))
		for _, price := range prices {
			_, err := decimal.NewFromString(price.Price)
			require.NoError(t, err)
			require.Equal(t, common.ATNUSDCSymbol, price.Symbol)
		}
		t.Log(prices)
	}
}

func TestNewUniswapClientWithNTNMarket(t *testing.T) {

	// using current piccadilly protocol configs.
	config := config2.PluginConfig{
		Name:               "crypto_uniswap",
		Scheme:             "wss",
		Endpoint:           "rpc-internal-1.piccadilly.autonity.org/ws",
		Timeout:            10,
		DataUpdateInterval: 30,
		// set the ATN token address with an un exist value, to let the market cannot be discovered.
		ATNTokenAddress:  "0xaE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d3",
		NTNTokenAddress:  "0xbE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d3",
		USDCTokenAddress: "0xc855D5e83363A4494e09f0Bb3152A70d3f161941",
		SwapAddress:      "0x218F76e357594C82Cc29A88B90dd67b180827c88",
	}

	client, err := NewUniswapClient(&config)
	require.NoError(t, err)

	defer client.Close()

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		prices, err := client.FetchPrice([]string{common.NTNUSDCSymbol, common.ATNUSDCSymbol})
		require.NoError(t, err)
		require.Equal(t, 0, len(prices))
		t.Log(prices)
	}
}

func TestNewUniswapClientWithNoMarkets(t *testing.T) {

	// using current piccadilly protocol configs.
	config := config2.PluginConfig{
		Name:               "crypto_uniswap",
		Scheme:             "wss",
		Endpoint:           "rpc-internal-1.piccadilly.autonity.org/ws",
		Timeout:            10,
		DataUpdateInterval: 30,
		// set the ATN token address with an un exist value, to let the market cannot be discovered.
		ATNTokenAddress:  "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d3",
		NTNTokenAddress:  types.AutonityContractAddress.Hex(),
		USDCTokenAddress: "0xB855D5e83363A4494e09f0Bb3152A70d3f161940",
		SwapAddress:      "0x218F76e357594C82Cc29A88B90dd67b180827c88",
	}

	client, err := NewUniswapClient(&config)
	require.NoError(t, err)

	defer client.Close()

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		prices, err := client.FetchPrice([]string{common.NTNUSDCSymbol, common.ATNUSDCSymbol})
		require.NoError(t, err)
		require.Equal(t, 1, len(prices))
		for _, price := range prices {
			_, err := decimal.NewFromString(price.Price)
			require.NoError(t, err)
			require.Equal(t, common.NTNUSDCSymbol, price.Symbol)
		}
		t.Log(prices)
	}
}

func TestNewUniswapClientWithWrongSwapAddress(t *testing.T) {

	// using current piccadilly protocol configs.
	config := config2.PluginConfig{
		Name:               "crypto_uniswap",
		Scheme:             "wss",
		Endpoint:           "rpc-internal-1.piccadilly.autonity.org/ws",
		Timeout:            10,
		DataUpdateInterval: 30,
		ATNTokenAddress:    "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2",
		NTNTokenAddress:    types.AutonityContractAddress.Hex(),
		USDCTokenAddress:   "0xB855D5e83363A4494e09f0Bb3152A70d3f161940",
		// set to wrong swap address.
		SwapAddress: "0x218F76e357594C82Cc29A88B90dd67b180827c89",
	}

	client, err := NewUniswapClient(&config)
	require.NoError(t, err)

	defer client.Close()

	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		prices, err := client.FetchPrice([]string{common.NTNUSDCSymbol, common.ATNUSDCSymbol})
		require.NoError(t, err)
		require.Equal(t, 0, len(prices))
		t.Log(prices)
	}
}

func TestNewUniswapClientWithWrongRPCEndpoint(t *testing.T) {
	// using current piccadilly protocol configs.
	config := config2.PluginConfig{
		Name:               "crypto_uniswap",
		Scheme:             "wss",
		Endpoint:           "replace with your host:port/path",
		Timeout:            10,
		DataUpdateInterval: 30,
		// set the ATN token address with an un exist value, to let the market cannot be discovered.
		ATNTokenAddress:  "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2",
		NTNTokenAddress:  types.AutonityContractAddress.Hex(),
		USDCTokenAddress: "0xB855D5e83363A4494e09f0Bb3152A70d3f161940",
		SwapAddress:      "0x218F76e357594C82Cc29A88B90dd67b180827c88",
	}

	client, err := NewUniswapClient(&config)
	require.NoError(t, err)

	defer client.Close()

	// The client would not panic with the wrong rpc endpoint.
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Second)
		prices, err := client.FetchPrice(supportedSymbols)
		require.NoError(t, err)
		require.Equal(t, 0, len(prices))
	}
}
