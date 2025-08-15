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
		expected := &VoteRecord{
			RoundID: 100,
			Salt:    big.NewInt(12345),
			Reports: []Report{
				{Price: big.NewInt(100), Confidence: 90},
			},
		}
		writeTestFile(t, dir, voteRecordFile, expected)

		// Execute
		actual, err := loadRecord[VoteRecord](dir, voteRecordFile)

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
		_, err := loadRecord[VoteRecord](dir, "missing.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no such file")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		dir := setupTestDir(t)
		path := filepath.Join(dir, voteRecordFile)
		require.NoError(t, os.WriteFile(path, []byte("{invalid-json}"), 0644))

		_, err := loadRecord[VoteRecord](dir, voteRecordFile)
		require.Error(t, err)
	})
}

func TestFlushRecord_TypeHandling(t *testing.T) {
	t.Run("VoteRecord", func(t *testing.T) {
		dir := setupTestDir(t)
		mem := &Memories{dataDir: dir}
		record := &VoteRecord{RoundID: 200}

		// Execute
		err := mem.flushRecord(record)

		// Validate
		require.NoError(t, err)
		path := filepath.Join(dir, voteRecordFile)
		assert.FileExists(t, path)

		var loaded VoteRecord
		data, _ := os.ReadFile(path)
		require.NoError(t, json.Unmarshal(data, &loaded))
		assert.Equal(t, uint64(200), loaded.RoundID)
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

func TestTypeConversions(t *testing.T) {
	t.Run("RoundDataRecord_ToRoundData", func(t *testing.T) {
		rr := &VoteRecord{
			RoundID: 300,
			Salt:    big.NewInt(999),
			Reports: []Report{
				{Price: big.NewInt(250), Confidence: 95},
			},
		}

		// Execute
		rd := rr.ToRoundData()

		// Validate
		assert.Equal(t, rr.RoundID, rd.RoundID)
		assert.Equal(t, rr.Salt.Int64(), rd.Salt.Int64())
		require.Len(t, rd.Reports, 1)
		assert.Equal(t, rr.Reports[0].Price.Int64(), rd.Reports[0].Price.Int64())
		assert.Equal(t, rr.Reports[0].Confidence, rd.Reports[0].Confidence)
	})

	t.Run("ToRoundRecord_RoundTrip", func(t *testing.T) {
		original := &types.RoundData{
			RoundID: 400,
			Salt:    big.NewInt(111),
			Reports: []contract.IOracleReport{
				{Price: big.NewInt(300), Confidence: 85},
			},
		}

		// Execute
		record := toVoteRecord(original)
		converted := record.ToRoundData()

		// Validate
		assert.Equal(t, original.RoundID, converted.RoundID)
		assert.Equal(t, original.Salt.Int64(), converted.Salt.Int64())
		require.Len(t, converted.Reports, 1)
		assert.Equal(t, original.Reports[0].Price.Int64(), converted.Reports[0].Price.Int64())
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
		assert.Nil(t, mem.voteRecord) // Should not error on missing round data
	})
}
