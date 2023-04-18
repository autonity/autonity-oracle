package oracleserver

import (
	orcMock "autonity-oracle/contract_binder/contract/mock"
	"autonity-oracle/types"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestOracleServer(t *testing.T) {
	var symbols = []string{"NTNUSD", "NTNAUD", "NTNCAD", "NTNEUR", "NTNGBP", "NTNJPY", "NTNSEK"}
	t.Run("oracle service getters", func(t *testing.T) {
		os := NewOracleServer(symbols, ".")
		version := os.Version()
		require.Equal(t, Version, version)
		actualSymbols := os.Symbols()
		require.Equal(t, symbols, actualSymbols)
		prices := os.GetPrices()
		require.Equal(t, 0, len(prices))
	})

	t.Run("oracle service setters", func(t *testing.T) {
		newSymbols := []string{"NTNUSD", "NTNAUD", "NTNCAD", "NTNEUR", "NTNGBP", "NTNJPY", "NTNSEK", "NTNRMB"}
		os := NewOracleServer(symbols, ".")
		os.UpdateSymbols(newSymbols)
		require.Equal(t, newSymbols, os.Symbols())

		NTNEURRate := types.Price{
			Timestamp: 0,
			Symbol:    "NTNEUR",
			Price:     decimal.RequireFromString("999.99"),
		}
		NTNUSDRate := types.Price{
			Timestamp: 0,
			Symbol:    "NTNUSD",
			Price:     decimal.RequireFromString("127.32"),
		}
		NTNGBPRate := types.Price{
			Timestamp: 0,
			Symbol:    "NTNGBP",
			Price:     decimal.RequireFromString("111.11"),
		}
		NTNRMBRate := types.Price{
			Timestamp: 0,
			Symbol:    "NTNRMB",
			Price:     decimal.RequireFromString("12.12"),
		}
		os.UpdatePrice(NTNEURRate)
		os.UpdatePrice(NTNUSDRate)
		os.UpdatePrice(NTNGBPRate)
		os.UpdatePrice(NTNRMBRate)

		require.Equal(t, 4, len(os.GetPrices()))
		actualPrices := os.GetPrices()
		require.Equal(t, true, NTNUSDRate.Price.Equals(actualPrices["NTNUSD"].Price))
		require.Equal(t, NTNUSDRate.Symbol, actualPrices["NTNUSD"].Symbol)

		require.Equal(t, true, NTNEURRate.Price.Equals(actualPrices["NTNEUR"].Price))
		require.Equal(t, NTNEURRate.Symbol, actualPrices["NTNEUR"].Symbol)

		require.Equal(t, true, NTNGBPRate.Price.Equals(actualPrices["NTNGBP"].Price))
		require.Equal(t, NTNGBPRate.Symbol, actualPrices["NTNGBP"].Symbol)

		require.Equal(t, true, NTNRMBRate.Price.Equals(actualPrices["NTNRMB"].Price))
		require.Equal(t, NTNRMBRate.Symbol, actualPrices["NTNRMB"].Symbol)
	})
}

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

		dp := &DataReporter{key: &keystore.Key{Address: common.Address{}},
			oracleContract: contractMock}

		isCommittee, err := dp.isVoter()
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

		dp := &DataReporter{oracleContract: contractMock, oracleService: oracleMock,
			roundData: make(map[uint64]*types.RoundData), logger: hclog.New(&hclog.LoggerOptions{
				Name:   "data contract_binder",
				Output: os.Stdout,
				Level:  hclog.Debug,
			})}
		roundData, err := dp.buildRoundData(uint64(100))
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
			Name:   "data contract_binder",
			Output: os.Stdout,
			Level:  hclog.Debug,
		})}

		dp.handleNewSymbolsEvent(symbols)
	})
}
