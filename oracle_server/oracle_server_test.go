package oracleserver

import (
	"autonity-oracle/config"
	cMock "autonity-oracle/contract_binder/contract/mock"
	"autonity-oracle/types/mock"
	"github.com/ethereum/go-ethereum/event"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"math/big"
	"os"
	"testing"
)

func TestOracleServer(t *testing.T) {
	currentRound := new(big.Int).SetUint64(1)
	precision := new(big.Int).SetUint64(10000000)
	votePeriod := new(big.Int).SetUint64(30)
	var subRoundEvent event.Subscription
	var subSymbolsEvent event.Subscription
	err := os.Setenv("ORACLE_KEY_FILE", "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe")
	require.NoError(t, err)
	err = os.Setenv("ORACLE_PLUGIN_DIR", "..//build/bin/plugins")
	require.NoError(t, err)
	defer os.Clearenv()

	t.Run("test init oracle server with oracle contract states", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		conf := config.MakeConfig()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(conf.Symbols, nil)
		contractMock.EXPECT().GetPrecision(nil).Return(precision, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)

		srv := NewOracleServer(conf.Symbols, conf.PluginDIR, conf.AutonityWSUrl, conf.Key, dialerMock, l1Mock, contractMock)

		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, conf.Symbols, srv.symbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromInt(precision.Int64())))
		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
	})
}

func TestDataReporter(t *testing.T) {
	/*
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
	*/
}
