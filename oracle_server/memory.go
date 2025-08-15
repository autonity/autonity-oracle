package oracleserver

import (
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/types"
	"encoding/json"
	"fmt"
	"math/big"
	o "os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/go-hclog"
)

const (
	outlierRecordFile = "outlier_record.json"
	voteRecordFile    = "vote_record.json"
)

// Memories stores persistent state loaded from data directory
type Memories struct {
	outlierRecord *OutlierRecord
	voteRecord    *VoteRecord
	dataDir       string
}

type Report struct {
	Price      *big.Int `json:"price"`
	Confidence uint8    `json:"confidence"`
}

type VoteRecord struct {
	RoundID  uint64   `json:"round_id"`
	Salt     *big.Int `json:"salt"`
	Reports  []Report `json:"reports"`
	LoggedAt string   `json:"logged_at"`
}

func (rr *VoteRecord) ToRoundData() *types.RoundData {
	to := &types.RoundData{
		RoundID: rr.RoundID,
		Salt:    rr.Salt,
		Reports: make([]contract.IOracleReport, len(rr.Reports)),
	}

	for i, report := range rr.Reports {
		to.Reports[i] = contract.IOracleReport{Price: report.Price, Confidence: report.Confidence}
	}
	return to
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
		logger.Info("Loading outlier record file", "error", err)
	}
	s.outlierRecord = outlierRecord
	roundDataRecord, err := s.loadVoteRecord()
	if err != nil {
		logger.Info("loading last round data record file", "error", err)
	}
	s.voteRecord = roundDataRecord
}

func toVoteRecord(data *types.RoundData) *VoteRecord {
	record := VoteRecord{
		RoundID:  data.RoundID,
		Salt:     data.Salt,
		LoggedAt: time.Now().Format(time.RFC3339),
		Reports:  make([]Report, len(data.Reports)),
	}
	for i, report := range data.Reports {
		record.Reports[i] = Report{
			Price:      report.Price,
			Confidence: report.Confidence,
		}
	}
	return &record
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

func (s *Memories) loadVoteRecord() (*VoteRecord, error) {
	return loadRecord[VoteRecord](s.dataDir, voteRecordFile)
}

func (s *Memories) loadOutlierRecord() (*OutlierRecord, error) {
	return loadRecord[OutlierRecord](s.dataDir, outlierRecordFile)
}

func (s *Memories) flushRecord(record interface{}) error {
	// Resolve filename using direct type comparison
	var fileName string
	switch record.(type) {
	case *OutlierRecord:
		fileName = outlierRecordFile
	case *VoteRecord:
		fileName = voteRecordFile
	default:
		panic("unexpected record type")
	}

	if _, err := o.Stat(s.dataDir); o.IsNotExist(err) {
		return fmt.Errorf("data dir does not exist: %s", s.dataDir)
	}
	filePath := filepath.Join(s.dataDir, fileName)
	file, err := o.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create record file: %s, %v", filePath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(record); err != nil {
		return fmt.Errorf("failed to marshual data in json: %v", err)
	}
	return nil
}
