//go:build usdx

package main

import (
	"autonity-oracle/plugins/common"
	"autonity-oracle/types"
)

func init() {
	// overwrite below configs for uniswap ATN-USDX, NTN-USDX plugin.
	version = "v0.2.0"
	ATNPairToStableCoin = "ATN-USDX"
	NTNPairToStableCoin = "NTN-USDX"
	supportedSymbols = []string{ATNPairToStableCoin, NTNPairToStableCoin, common.NTNATNSymbol}
	usdStableCoinDecimals = common.USDXDecimals // todo: double check with Edward or Jay?
	defaultConfig = types.PluginConfig{
		Name:               "crypto_uniswap_usdx",             // The default built binary is named for ATN-USDC, NTN-USDC market plugin.
		Scheme:             "wss",                             // both http/s ws/s works for this plugin, todo: update this on redeployment of infra
		Endpoint:           "rpc1.piccadilly.autonity.org/ws", // todo: update this on redeployment of infra
		Timeout:            10,                                // 10s
		DataUpdateInterval: 30,                                // 30s
		// todo: update below protocol contract addresses on redeployment of protocols.
		NTNTokenAddress: NTNTokenAddress.Hex(),                        // NTN ERC20 token address on the target blockchain.
		ATNTokenAddress: "0xcE17e51cE4F0417A1aB31a3c5d6831ff3BbFa1d2", // Wrapped ATN ERC20 contract address on the target blockchain.
		USDTokenAddress: "0x3a60C03a86eEAe30501ce1af04a6C04Cf0188700", // USDX ERC20 contract address on the target blockchain.
		SwapAddress:     "0x218F76e357594C82Cc29A88B90dd67b180827c88", // UniSwap factory contract address on the target blockchain.
	}
}
