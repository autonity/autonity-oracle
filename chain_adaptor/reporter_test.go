package chain_adaptor

import (
	mock_oracle "autonity-oracle/chain_adaptor/contract/mock"
	"autonity-oracle/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
)

func TestDataReporter(t *testing.T) {
	/*
		priKey, err := crypto.GenerateKey()
		require.NoError(t, err)
		ws := config.DefaultAutonityWSUrl
		validator := common.Address{}
		key := &keystore.Key{
			PrivateKey: priKey,
			Address:    crypto.PubkeyToAddress(priKey.PublicKey),
		}*/
	//oracle := &oracleserver.OracleServer{}

	t.Run("gc round data", func(t *testing.T) {
		dp := &DataReporter{
			roundData: make(map[uint64]*types.RoundData),
		}
		for r := 0; r < 100; r++ {
			dp.currentRound = uint64(r)
			var roundData = &types.RoundData{}
			dp.roundData[uint64(r)] = roundData
		}
		require.Equal(t, 100, len(dp.roundData))
		dp.gcRoundData()
		require.Equal(t, MaxBufferedRounds, len(dp.roundData))
	})

	t.Run("get starting states", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		currentRound := new(big.Int).SetUint64(100)
		symbols := []string{"NTNUSD", "NTNEUR"}
		contractMock := mock_oracle.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).AnyTimes().Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).AnyTimes().Return(symbols, nil)

		retRnd, retSymbols, err := getStartingStates(contractMock)
		require.Equal(t, currentRound.Uint64(), retRnd)
		require.Equal(t, symbols, retSymbols)
		require.NoError(t, err)
	})

	t.Run("is committee member", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		validatorAddr := common.Address{}
		committee := []common.Address{validatorAddr}
		contractMock := mock_oracle.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetCommittee(nil).AnyTimes().Return(committee, nil)

		dp := &DataReporter{validatorAccount: validatorAddr,
			oracleContract: contractMock}

		isCommittee, err := dp.isCommitteeMember()
		require.Equal(t, true, isCommittee)
		require.NoError(t, err)
	})
}
