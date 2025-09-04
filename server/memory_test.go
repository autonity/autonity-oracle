package server

import (
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/types"
	"encoding/json"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================
// Test Setup Utilities
// =====================

func setupTestDir(t *testing.T) string {
	dir := t.TempDir()
	require.DirExists(t, dir)
	return dir
}

func writeTestFile(t *testing.T, dir, filename string, content interface{}) string {
	path := filepath.Join(dir, filename)

	file, err := os.OpenFile(
		path,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	require.NoError(t, err, "Failed to create file")
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	require.NoError(t, encoder.Encode(content), "Failed to encode JSON")
	return path
}

// ================
// Core Test Cases
// ================

func TestLoadRecord_ValidCases(t *testing.T) {
	t.Run("VoteRecords", func(t *testing.T) {
		dir := setupTestDir(t)

		// Create a VoteRecords map instead of single record
		expected := VoteRecords{
			1: &types.VoteRecord{
				RoundID: 1,
				Salt:    big.NewInt(12345),
				Reports: []contract.IOracleReport{
					{Price: big.NewInt(100), Confidence: 90},
				},
			},
			2: &types.VoteRecord{
				RoundID: 2,
				Salt:    big.NewInt(67890),
				Reports: []contract.IOracleReport{
					{Price: big.NewInt(200), Confidence: 95},
				},
			},
		}

		writeTestFile(t, dir, voteRecordFile, expected)

		// Execute
		actual, err := loadRecord[VoteRecords](dir, voteRecordFile)

		// Validate
		require.NoError(t, err)
		require.Len(t, *actual, 2)
		assert.Equal(t, uint64(1), (*actual)[1].RoundID)
		assert.Equal(t, int64(12345), (*actual)[1].Salt.Int64())
		assert.Equal(t, uint64(2), (*actual)[2].RoundID)
		assert.Equal(t, int64(67890), (*actual)[2].Salt.Int64())
	})

	t.Run("OutlierRecord", func(t *testing.T) {
		dir := setupTestDir(t)
		expected := &OutlierRecord{
			Participant:          common.HexToAddress("0x1"),
			LastPenalizedAtBlock: 150000,
			SlashingAmount:       500,
		}
		writeTestFile(t, dir, outlierRecordFile, expected)

		// Execute
		actual, err := loadRecord[OutlierRecord](dir, outlierRecordFile)

		// Validate
		require.NoError(t, err)
		assert.Equal(t, expected.Participant, actual.Participant)
		assert.Equal(t, expected.SlashingAmount, actual.SlashingAmount)
	})
}

func TestLoadRecord_ErrorScenarios(t *testing.T) {
	t.Run("NonexistentFile", func(t *testing.T) {
		dir := setupTestDir(t)
		_, err := loadRecord[VoteRecords](dir, "missing.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no such file")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		dir := setupTestDir(t)
		path := filepath.Join(dir, voteRecordFile)
		require.NoError(t, os.WriteFile(path, []byte("{invalid-json}"), 0600))

		_, err := loadRecord[VoteRecords](dir, voteRecordFile)
		require.Error(t, err)
	})
}

func TestFlushRecord_TypeHandling(t *testing.T) {
	t.Run("VoteRecords", func(t *testing.T) {
		dir := setupTestDir(t)
		mem := &Memories{dataDir: dir}

		// Create a populated VoteRecords map
		records := VoteRecords{
			1: &types.VoteRecord{RoundID: 1},
			2: &types.VoteRecord{RoundID: 2},
		}

		// Execute
		err := mem.flushRecord(records)

		// Validate
		require.NoError(t, err)
		path := filepath.Join(dir, voteRecordFile)
		assert.FileExists(t, path)

		var loaded VoteRecords
		data, _ := os.ReadFile(path)
		require.NoError(t, json.Unmarshal(data, &loaded))
		require.Len(t, loaded, 2)
		assert.Equal(t, uint64(1), loaded[1].RoundID)
		assert.Equal(t, uint64(2), loaded[2].RoundID)
	})

	t.Run("OutlierRecord", func(t *testing.T) {
		dir := setupTestDir(t)
		mem := &Memories{dataDir: dir}
		record := &OutlierRecord{SlashingAmount: 1000}

		// Execute
		err := mem.flushRecord(record)

		// Validate
		require.NoError(t, err)
		path := filepath.Join(dir, outlierRecordFile)
		assert.FileExists(t, path)
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		mem := &Memories{dataDir: t.TempDir()}
		require.Panics(t, func() {
			_ = mem.flushRecord("invalid-type")
		})
	})
}

func TestMemoriesInit(t *testing.T) {
	t.Run("PartialInitialization", func(t *testing.T) {
		dir := setupTestDir(t)
		writeTestFile(t, dir, outlierRecordFile, &OutlierRecord{
			Participant: common.HexToAddress("0x2"),
		})

		// Create vote records file
		records := VoteRecords{
			1: &types.VoteRecord{RoundID: 1},
		}
		writeTestFile(t, dir, voteRecordFile, records)

		mem := &Memories{dataDir: dir}

		// Execute
		logger := hclog.New(&hclog.LoggerOptions{
			Output: io.Discard, // Discard log output
		})
		mem.init(logger)

		// Validate
		assert.NotNil(t, mem.outlierRecord)
		assert.NotNil(t, mem.voteRecords)
		assert.Len(t, *mem.voteRecords, 1)
		assert.Equal(t, uint64(1), (*mem.voteRecords)[1].RoundID)
	})

	t.Run("VoteRecordsLoading", func(t *testing.T) {
		dir := setupTestDir(t)

		// Create multi-record vote file
		records := VoteRecords{
			10: &types.VoteRecord{RoundID: 10, Salt: big.NewInt(100)},
			20: &types.VoteRecord{RoundID: 20, Salt: big.NewInt(200)},
		}
		writeTestFile(t, dir, voteRecordFile, records)

		mem := &Memories{dataDir: dir}
		logger := hclog.New(&hclog.LoggerOptions{Output: io.Discard})
		mem.init(logger)

		require.NotNil(t, mem.voteRecords)
		assert.Len(t, *mem.voteRecords, 2)
		assert.Equal(t, int64(100), (*mem.voteRecords)[10].Salt.Int64())
		assert.Equal(t, int64(200), (*mem.voteRecords)[20].Salt.Int64())
	})

	t.Run("EmptyVoteRecords", func(t *testing.T) {
		dir := setupTestDir(t)
		// Create empty vote records file
		writeTestFile(t, dir, voteRecordFile, VoteRecords{})

		mem := &Memories{dataDir: dir}
		logger := hclog.New(&hclog.LoggerOptions{Output: io.Discard})
		mem.init(logger)

		require.NotNil(t, mem.voteRecords)
		assert.Empty(t, *mem.voteRecords)
	})

	t.Run("MissingFiles", func(t *testing.T) {
		dir := setupTestDir(t) // Empty directory
		mem := &Memories{dataDir: dir}
		logger := hclog.New(&hclog.LoggerOptions{Output: io.Discard})

		// Should not panic with missing files
		mem.init(logger)

		assert.Nil(t, mem.outlierRecord)
		assert.Nil(t, mem.voteRecords)
	})
}
