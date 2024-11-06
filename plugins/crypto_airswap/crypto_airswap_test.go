package main

import (
	swaperc20 "autonity-oracle/plugins/crypto_airswap/swap_erc20"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"testing"
)

func TestExtractOrder(t *testing.T) {
	testDataFile := "./test-log.json"
	logs, err := ReadLogsFromFile(testDataFile)
	require.NoError(t, err)
	require.NotEmpty(t, logs)

	config := defaultConfig
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   config.Name,
		Level:  hclog.Info,
		Output: os.Stdout,
	})
	client, err := NewAirswapClient(&config, logger)
	require.NoError(t, err)

	swapEvent := &swaperc20.Swaperc20SwapERC20{
		Nonce:        new(big.Int).SetUint64(1719803346312),
		SignerWallet: common.HexToAddress("0x51c72848c68a965f66fa7a88855f9f7784502a7f"),
	}

	_, err = client.extractOrder(logs, swapEvent)
	require.NoError(t, err)
}

// Log represents the Ethereum transaction log structure.
type Log struct {
	Address     common.Address `json:"address" gencodec:"required"`
	Topics      []common.Hash  `json:"topics" gencodec:"required"`
	Data        hexutil.Bytes  `json:"data" gencodec:"required"` // Use hexutil.Bytes for hex data
	BlockNumber uint64         `json:"blockNumber"`              // Change to uint64
	TxHash      common.Hash    `json:"transactionHash" gencodec:"required"`
	TxIndex     uint           `json:"transactionIndex"`
	BlockHash   common.Hash    `json:"blockHash"`
	Index       uint           `json:"logIndex"`
	Removed     bool           `json:"removed"`
}

// ReadLogsFromFile reads a JSON file and unmarshals it into a slice of Log objects.
func ReadLogsFromFile(filename string) ([]*types.Log, error) {
	// Open the JSON file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode the JSON data
	var logs []*Log
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&logs); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Convert []*Log to []*RawLog
	rawLogs := make([]*types.Log, len(logs))
	for i, log := range logs {
		rawLogs[i] = &types.Log{
			Address:     log.Address,
			Topics:      log.Topics,
			Data:        log.Data,
			BlockNumber: log.BlockNumber,
			TxHash:      log.TxHash,
			TxIndex:     log.TxIndex,
			BlockHash:   log.BlockHash,
			Index:       log.Index,
			Removed:     log.Removed,
		}
	}

	return rawLogs, nil
}
