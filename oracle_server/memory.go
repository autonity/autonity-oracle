package oracleserver

import (
	"autonity-oracle/types"
	"encoding/json"
	"errors"
	"fmt"
	o "os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

const (
	outlierRecordFile = "outlier_record.json"
	voteRecordFile    = "vote_record.json"
)

// VoteRecords stores the most recent MaxBufferedRounds (10) rounds vote records.
type VoteRecords map[uint64]*types.VoteRecord

// Memories stores persistent state loaded from data directory
type Memories struct {
	outlierRecord *OutlierRecord
	voteRecords   *VoteRecords
	dataDir       string
}

type OutlierRecord struct {
	LastPenalizedAtBlock uint64         `json:"last_penalized_at_block"`
	Participant          common.Address `json:"participant"`
	Symbol               string         `json:"symbol"`
	Median               uint64         `json:"median"`
	Reported             uint64         `json:"reported"`
	SlashingAmount       uint64         `json:"slashingAmount"`
	LoggedAt             string         `json:"logged_at"`
}

func (s *Memories) init(logger hclog.Logger) {
	outlierRecord, err := s.loadOutlierRecord()
	if err != nil {
		if errors.Is(err, o.ErrNotExist) {
			logger.Info("There is no outlier record in the profile data directory.")
		} else {
			// as there is no recovery mechanism for the corrupted data engine, thus we don't panic.
			logger.Warn("Loading outlier record file", "error", err)
			logger.Warn("Running server without any outlier record from persistence layer")
		}
	}
	s.outlierRecord = outlierRecord
	voteRecords, err := s.loadVoteRecords()
	if err != nil {
		if errors.Is(err, o.ErrNotExist) {
			logger.Info("There is no vote record in the profile data directory.")
		} else {
			// as there is no recovery mechanism for the corrupted data engine, thus we don't panic.
			logger.Info("loading last vote record file", "error", err)
			logger.Warn("Running server without any vote record from persistence layer")
		}
	}
	s.voteRecords = voteRecords
	if outlierRecord != nil {
		logger.Info("Loaded outlier record", "outlierRecord", outlierRecord)
	}
	if voteRecords != nil {
		for k, v := range *voteRecords {
			logger.Info("Loaded vote record", "round", k, "vote", v)
		}
	}
}

func loadRecord[T any](dir, filename string) (*T, error) {
	path := filepath.Join(dir, filename)
	if _, err := o.Stat(path); err != nil {
		return nil, err
	}

	data, err := o.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var record T
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *Memories) loadVoteRecords() (*VoteRecords, error) {
	return loadRecord[VoteRecords](s.dataDir, voteRecordFile)
}

func (s *Memories) loadOutlierRecord() (*OutlierRecord, error) {
	return loadRecord[OutlierRecord](s.dataDir, outlierRecordFile)
}

// Note! This is not a thread safe data flushing function.
func (s *Memories) flushRecord(record interface{}) error {
	// Resolve filename using direct type comparison
	var fileName string
	switch record.(type) {
	case *OutlierRecord:
		fileName = outlierRecordFile
	case VoteRecords:
		fileName = voteRecordFile
	default:
		panic("unexpected record type")
	}

	// As the existence of the directory was checked on config loading phase. Just create file under it.
	filePath := filepath.Join(s.dataDir, fileName)
	// limit the file with R&W permission only for its owner.
	file, err := o.OpenFile(filePath, o.O_RDWR|o.O_CREATE|o.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create record file: %s, %v", filePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(record); err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %v", err)
	}
	return nil
}
