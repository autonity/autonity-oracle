package reporter

import (
	orcMock "autonity-oracle/reporter/contract/mock"
	"autonity-oracle/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"testing"
	"time"
)

func TestDataReporter(t *testing.T) {

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
		contractMock := orcMock.NewMockContractAPI(ctrl)
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
		contractMock := orcMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetVoters(nil).AnyTimes().Return(committee, nil)

		dp := &DataReporter{validatorAccount: validatorAddr,
			oracleContract: contractMock}

		isCommittee, err := dp.isCommitteeMember()
		require.Equal(t, true, isCommittee)
		require.NoError(t, err)
	})

	t.Run("test build round data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		now := time.Now()
		symbols := []string{"NTNUSD", "NTNEUR", "NTNAUD"}
		prices := make(types.PriceBySymbol)
		for _, s := range symbols {
			prices[s] = types.Price{
				Timestamp: now.UnixMilli(),
				Symbol:    s,
				Price:     decimal.RequireFromString("11.11"),
			}
		}
		contractMock := orcMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetSymbols(nil).AnyTimes().Return(symbols, nil)
		oracleMock := orcMock.NewMockOracleService(ctrl)
		oracleMock.EXPECT().GetPricesBySymbols(symbols).Return(prices)

		dp := &DataReporter{oracleContract: contractMock, oracleService: oracleMock, roundData: make(map[uint64]*types.RoundData)}
		roundData, err := dp.buildRoundData()
		require.NoError(t, err)
		require.Equal(t, symbols, roundData.Symbols)
		require.Equal(t, prices, roundData.Prices)

		var sourceBytes []byte
		for _, s := range symbols {
			sourceBytes = append(sourceBytes, common.LeftPadBytes(prices[s].Price.Mul(PricePrecision).BigInt().Bytes(), 32)...)
		}

		sourceBytes = append(sourceBytes, common.LeftPadBytes(roundData.Salt.Bytes(), 32)...)
		expectedHash := crypto.Keccak256Hash(sourceBytes)
		require.Equal(t, expectedHash, roundData.Hash)
	})

	t.Run("handle new symbol event", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		symbols := []string{"NTNUSD", "NTNEUR", "NTNAUD"}
		oracleMock := orcMock.NewMockOracleService(ctrl)
		oracleMock.EXPECT().UpdateSymbols(symbols)

		dp := &DataReporter{oracleService: oracleMock, logger: hclog.New(&hclog.LoggerOptions{
			Name:   "data reporter",
			Output: os.Stdout,
			Level:  hclog.Debug,
		})}

		dp.handleNewSymbolsEvent(symbols)
		require.Equal(t, symbols, dp.currentSymbols)
	})
}
