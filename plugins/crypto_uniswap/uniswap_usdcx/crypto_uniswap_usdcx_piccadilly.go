package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	client "autonity-oracle/plugins/crypto_uniswap/common"
	"autonity-oracle/types"
	"os"
)

// configs for the ATN-USDCx, NTN-USDCx, NTN-ATN market place in Piccadilly network.
var defaultConfigPicc = config.PluginConfig{
	Name:               "crypto_uniswap",
	Scheme:             "wss",                                        // both http/s ws/s works for this plugin
	Endpoint:           "rpc-internal-1.piccadilly.autonity.org/ws",  // default websocket endpoint for piccadilly network.
	Timeout:            10,                                           // 10s
	DataUpdateInterval: common.DefaultAMMDataUpdateInterval,          // 1s, shorten the default data point refresh interval for AMM market data, as they can move very fast.
	NTNTokenAddress:    types.AutonityContractAddress.Hex(),          // Same as 0xBd770416a3345F91E4B34576cb804a576fa48EB1, Autonity contract address.
	ATNTokenAddress:    "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2", // Wrapped ATN ERC20 contract address on the target blockchain.
	USDCTokenAddress:   "0xB855D5e83363A4494e09f0Bb3152A70d3f161940", // USDCx ERC20 contract address on the target blockchain.
	SwapAddress:        "0x218F76e357594C82Cc29A88B90dd67b180827c88", // UniSwap factory contract address on the target blockchain.
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfigPicc)
	c, err := client.NewUniswapClient(conf)
	if err != nil {
		return
	}

	// start the uniswapV2 event watching for price aggregation of ATN-USDCx & NTN-USDCx
	go c.StartWatcher()

	adapter := common.NewPlugin(conf, c, client.Version, types.SrcAMM, common.ChainIDPiccadilly)
	defer adapter.Close()
	common.PluginServe(adapter)
}
