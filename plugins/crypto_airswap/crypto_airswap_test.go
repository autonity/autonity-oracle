package main

import (
	swaperc20 "autonity-oracle/plugins/crypto_airswap/swap_erc20"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"testing"
)

// todo: Jason, the best way to test this could be dump the swap event logs into a json file, then we can remove the
// network dependency of piccadilly to test this component by taking event logs from this test-data.json.
func TestAirswapClientWithPiccadilly(t *testing.T) {
	config := defaultConfig
	config.Endpoint = "rpc2.piccadilly.autonity.org/ws"
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   config.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})
	client, err := NewAirswapClient(&config, logger)
	require.NoError(t, err)
	defer client.client.Close()
	defer client.subSwapEvent.Unsubscribe()

	swapEvents := []struct {
		txn       common.Hash
		swapEvent *swaperc20.Swaperc20SwapERC20
		expectErr bool
		symbol    string
		ratio     string
	}{
		{
			txn: common.HexToHash("0xc75b9ec0dd99a224bab9d0ee894cb5bb26ae44a915a60c202371598997fddb18"),
			swapEvent: &swaperc20.Swaperc20SwapERC20{
				Nonce:        new(big.Int).SetUint64(1719327692657),
				SignerWallet: common.HexToAddress("0x5eb2f7511405b0cb0f24abdc77412fa1dfe68f3e"),
			},
			expectErr: true, // it is a swap between NTN and ATN, should rise an error to skip it.
		},
		{
			txn: common.HexToHash("0x29ff16bbddf531b86c627932cff2f0e8f15e81b2a76e161d878f8f98ab7a2148"),
			swapEvent: &swaperc20.Swaperc20SwapERC20{
				Nonce:        new(big.Int).SetUint64(1719328735436),
				SignerWallet: common.HexToAddress("0xf47fdd88c8f6f80239e177386cc5ae3d6bcdeeea"),
			},
			expectErr: true, // it is a swap between NTN and ATN, should rise an error to skip it.
		},
		{
			txn: common.HexToHash("0x32979c314f61cf95c045feb51bbf5eef72657cc03bfcf15f95df54fb262dca27"),
			swapEvent: &swaperc20.Swaperc20SwapERC20{
				Nonce:        new(big.Int).SetUint64(1719617957284),
				SignerWallet: common.HexToAddress("0x4a2f43996d1fc03b054d89963f395c6ebff02cad"),
			},
			expectErr: false,
			symbol:    NTNUSDC,
			ratio:     "0.1002004",
		},

		{
			txn: common.HexToHash("0x6a6477aed203b0f95d0f5f72ba780a528bbaac2db905d53dc317e0c0c9004723"),
			swapEvent: &swaperc20.Swaperc20SwapERC20{
				Nonce:        new(big.Int).SetUint64(1719803346312),
				SignerWallet: common.HexToAddress("0x4a2f43996d1fc03b054d89963f395c6ebff02cad"),
			},
			expectErr: false,
			symbol:    NTNUSDC,
			ratio:     "0.1110933",
		},
	}

	for _, swaps := range swapEvents {
		err = client.handleSwapEvent(swaps.txn, swaps.swapEvent)
		if swaps.expectErr {
			require.Error(t, err)
			continue
		} else {
			require.NoError(t, err)
		}
		prices, err := client.FetchPrice(nil)
		require.NoError(t, err)
		require.Equal(t, 1, len(prices))
		require.Equal(t, swaps.symbol, prices[0].Symbol)
		require.Equal(t, swaps.ratio, prices[0].Price)
	}
}
