package oracleserver

import (
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/types"
	"encoding/json"
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
	file, err := os.Create(path)
	require.NoError(t, err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	require.NoError(t, encoder.Encode(content))
	return path
}

// ================
// Core Test Cases
// ================

func TestLoadRecord_ValidCases(t *testing.T) {
	t.Run("VoteRecord", func(t *testing.T) {
		dir := setupTestDir(t)
		expected := &types.VoteRecord{
			RoundID: 100,
			Salt:    big.NewInt(12345),
			Reports: []contract.IOracleReport{
				{Price: big.NewInt(100), Confidence: 90},
			},
		}
		writeTestFile(t, dir, voteRecordFile, expected)

		// Execute
		actual, err := loadRecord[types.VoteRecord](dir, voteRecordFile)

		// Validate
		require.NoError(t, err)
		assert.Equal(t, expected.RoundID, actual.RoundID)
		assert.Equal(t, expected.Salt.Int64(), actual.Salt.Int64())
		assert.Len(t, actual.Reports, 1)
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
		_, err := loadRecord[types.VoteRecord](dir, "missing.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no such file")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		dir := setupTestDir(t)
		path := filepath.Join(dir, voteRecordFile)
		require.NoError(t, os.WriteFile(path, []byte("{invalid-json}"), 0600))

		_, err := loadRecord[types.VoteRecord](dir, voteRecordFile)
		require.Error(t, err)
	})
}

func TestFlushRecord_TypeHandling(t *testing.T) {
	t.Run("VoteRecord", func(t *testing.T) {
		dir := setupTestDir(t)
		mem := &Memories{dataDir: dir}
		records := VoteRecords{}

		// Execute
		err := mem.flushRecord(records)

		// Validate
		require.NoError(t, err)
		path := filepath.Join(dir, voteRecordFile)
		assert.FileExists(t, path)

		var loaded VoteRecords
		data, _ := os.ReadFile(path)
		require.NoError(t, json.Unmarshal(data, &loaded))
		assert.Equal(t, records, loaded)
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

		mem := &Memories{dataDir: dir}

		// Execute
		mem.init(hclog.Default()) // Use default logger

		// Validate
		assert.NotNil(t, mem.outlierRecord)
		assert.Nil(t, mem.voteRecords) // Should not error on missing round data
	})
}
