package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	client "autonity-oracle/plugins/crypto_uniswap/common"
	"autonity-oracle/types"
	"os"
)

// configs for the ATN-USDCx marketplace in Bakerloo network.
var defaultConfig = config.PluginConfig{
	Name:               "crypto_uniswap",
	Scheme:             "wss",                                        // both ws and wss works for this plugin.
	Endpoint:           "replace with your host:port/path",           // default websocket endpoint for bakerloo network.
	Timeout:            10,                                           // 10s.
	DataUpdateInterval: common.DefaultAMMDataUpdateInterval,          // 1s, shorten the default data point refresh interval for AMM market data, as they can move very fast.
	NTNTokenAddress:    types.AutonityContractAddress.Hex(),          // Same as 0xBd770416a3345F91E4B34576cb804a576fa48EB1, Autonity contract address.
	ATNTokenAddress:    "0x7152e69E173D631ee7B8df89b98fd25decb7263D", // Wrapped ATN ERC20 contract address on the target blockchain.
	USDCTokenAddress:   "0x90488152F52e1aDc63CaA2CDb6Ad84F3AEC1df3E", // USDCx ERC20 contract address on the target blockchain.
	SwapAddress:        "0x9709D1709bDE7C59716FE74D3EEad0b1f12D3944", // UniSwap factory contract address on the target blockchain.
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	c, err := client.NewUniswapClient(conf)
	if err != nil {
		return
	}

	adapter := common.NewPlugin(conf, c, client.Version, types.SrcAMM, common.ChainIDBakerloo)
	defer adapter.Close()
	common.PluginServe(adapter)
}
