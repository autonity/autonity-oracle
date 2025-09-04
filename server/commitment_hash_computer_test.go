package server

import (
	contract "autonity-oracle/contract_binder/contract"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCommitmentHash(t *testing.T) {
	computer, err := NewCommitmentHashComputer()
	require.NoError(t, err)

	report := []contract.IOracleReport{
		{
			Price:      common.Big1,
			Confidence: 1,
		},
	}
	salt := common.Big1

	msgSender := common.HexToAddress("0x71562b71999873DB5b286dF957af199Ec94617F7")
	hash, err := computer.CommitmentHash(report, salt, msgSender)
	require.NoError(t, err)
	require.Equal(t, "0x08968f6f64cc0f74029fcd9b21203ba53a59600456f4ccf58aee3476dddd39f1", hash.String())
}
