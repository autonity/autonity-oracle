package oracleserver

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	cMock "autonity-oracle/contract_binder/contract/mock"
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"autonity-oracle/types/mock"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	tp "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

var BridgerSymbols = []string{NTNUSDC, ATNUSDC, USDCUSD}
var DefaultSampledSymbols = []string{"AUD-USD", "CAD-USD", "EUR-USD", "GBP-USD", "JPY-USD", "SEK-USD", "ATN-USD", "NTN-USD", "NTN-ATN", "ATN-USDC", "NTN-USDC", "USDC-USD"}
var ChainIDPiccadilly = big.NewInt(65_100_004)
var testKeyFile = "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"

func TestOracleDecimals(t *testing.T) {
	decimals := uint8(18)
	precision := decimal.NewFromBigInt(common.Big1, int32(decimals))
	require.Equal(t, "1000000000000000000", precision.String())
}

func TestOracleServer(t *testing.T) {
	currentRoundHeight := new(big.Int).SetUint64(0)
	currentRound := new(big.Int).SetUint64(1)
	precision := OracleDecimals
	votePeriod := new(big.Int).SetUint64(30)
	var subRoundEvent event.Subscription
	var subSymbolsEvent event.Subscription
	var subPenalizeEvent event.Subscription
	var subVoteEvent event.Subscription
	var subInvalidVoteEvent event.Subscription
	var subReportedEvent event.Subscription
	var subNoRevealEvent event.Subscription

	keyFile := testKeyFile
	passWord := config.DefaultConfig.KeyPassword
	key, err := config.LoadKey(keyFile, passWord)
	require.NoError(t, err)

	conf := &config.Config{
		ConfigFile:         "../test_data/oracle_config.yml",
		LoggingLevel:       hclog.Level(config.DefaultConfig.LoggingLevel), //nolint
		GasTipCap:          config.DefaultConfig.GasTipCap,
		VoteBuffer:         config.DefaultConfig.VoteBuffer,
		Key:                key,
		AutonityWSUrl:      config.DefaultConfig.AutonityWSUrl,
		PluginDIR:          "../plugins/template_plugin/bin",
		ProfileDir:         ".",
		ConfidenceStrategy: 0,
		PluginConfigs:      nil,
		MetricConfigs:      config.MetricConfig{},
	}

	t.Run("test init oracle server with oracle contract states", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetLastRoundBlock(nil).Return(currentRoundHeight, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)
		contractMock.EXPECT().WatchNoRevealPenalty(gomock.Any(), gomock.Any(), gomock.Any()).Return(subNoRevealEvent, nil)

		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(ChainIDPiccadilly, nil)
		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))

		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.runningPlugins))
		require.Equal(t, "template_plugin", srv.runningPlugins["template_plugin"].Name())
		srv.runningPlugins["template_plugin"].Close()
	})

	t.Run("test pre-sampling happy case", func(t *testing.T) {
		roundID := currentRound.Uint64() + 1
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		chainHeight := uint64(56)
		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(new(big.Int).SetUint64(roundID), nil)
		contractMock.EXPECT().GetLastRoundBlock(nil).Return(currentRoundHeight, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)
		contractMock.EXPECT().WatchNoRevealPenalty(gomock.Any(), gomock.Any(), gomock.Any()).Return(subNoRevealEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().BlockNumber(gomock.Any()).AnyTimes().Return(chainHeight, nil)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(ChainIDPiccadilly, nil)
		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)

		ts := time.Now().Unix()
		srv.curSampleTS = ts
		srv.curRoundHeight = uint64(30)

		for sec := ts; sec < ts+15; sec++ {
			err := srv.handlePreSampling(sec)
			require.NoError(t, err)
			time.Sleep(time.Second)
		}

		voteRecord, err := srv.buildVoteRecord(roundID)
		require.NoError(t, err)
		require.Equal(t, roundID, voteRecord.RoundID)
		require.Equal(t, helpers.DefaultSymbols, voteRecord.Symbols)
		require.Equal(t, len(helpers.DefaultSymbols), len(voteRecord.Prices))
		require.Equal(t, true, helpers.ResolveSimulatedPrice(NTNUSD).Equal(voteRecord.Prices[NTNUSD].Price))
		require.Equal(t, true, helpers.ResolveSimulatedPrice(ATNUSD).Equal(voteRecord.Prices[ATNUSD].Price))
		t.Log(voteRecord)
		srv.gcStaleSamples()
		srv.runningPlugins["template_plugin"].Close()
	})

	t.Run("test round vote happy case, with commitment and round data", func(t *testing.T) {
		round := new(big.Int).SetUint64(2)
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		chainHeight := uint64(55)
		header := &tp.Header{BaseFee: common.Big256}

		var voters []common.Address
		voters = append(voters, conf.Key.Address)
		price := contract.IOracleRoundData{
			Round:     round,
			Price:     new(big.Int).SetUint64(0),
			Timestamp: new(big.Int).SetUint64(10000),
			Success:   true,
		}

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(round, nil)
		contractMock.EXPECT().GetLastRoundBlock(nil).Return(currentRoundHeight, nil)
		contractMock.EXPECT().GetSymbols(nil).AnyTimes().Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().GetVoters(nil).Return(voters, nil)
		contractMock.EXPECT().GetRoundData(nil, new(big.Int).SetUint64(2), gomock.Any()).AnyTimes().Return(price, nil)
		contractMock.EXPECT().LatestRoundData(nil, gomock.Any()).AnyTimes().Return(price, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)
		contractMock.EXPECT().WatchNoRevealPenalty(gomock.Any(), gomock.Any(), gomock.Any()).Return(subNoRevealEvent, nil)
		txdata := &tp.DynamicFeeTx{ChainID: new(big.Int).SetUint64(1000), Nonce: 1}
		tx := tp.NewTx(txdata)
		contractMock.EXPECT().Vote(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tx, nil)

		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(ChainIDPiccadilly, nil)
		l1Mock.EXPECT().BlockNumber(gomock.Any()).AnyTimes().Return(chainHeight, nil)
		l1Mock.EXPECT().SyncProgress(gomock.Any()).Return(nil, nil)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(new(big.Int).SetUint64(1000), nil)
		l1Mock.EXPECT().SuggestGasTipCap(gomock.Any()).Return(new(big.Int).SetUint64(1000), nil)
		l1Mock.EXPECT().HeaderByNumber(gomock.Any(), nil).Return(header, nil)
		l1Mock.EXPECT().BalanceAt(gomock.Any(), gomock.Any(), gomock.Any()).Return(alertBalance, nil)
		l1Mock.EXPECT().FilterLogs(gomock.Any(), gomock.Any()).Return(nil, nil)
		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)

		// prepare last round data.
		prices := make(types.PriceBySymbol)
		for _, s := range helpers.DefaultSymbols {
			prices[s] = types.Price{
				Timestamp: 0,
				Symbol:    s,
				Price:     helpers.ResolveSimulatedPrice(s),
			}
		}

		voteRecord, err := srv.assembleVote(srv.curRound, helpers.DefaultSymbols, prices)
		require.NoError(t, err)
		srv.voteRecords[srv.curRound] = voteRecord

		// pre-sampling with data.
		ts := time.Now().Unix()
		srv.curSampleTS = ts
		srv.curRoundHeight = uint64(30)
		for sec := ts; sec < ts+15; sec++ {
			err = srv.handlePreSampling(time.Now().Unix())
			require.NoError(t, err)
			time.Sleep(time.Second)
		}

		// handle vote event that change to next round with
		srv.curRound = srv.curRound + 1
		srv.curRoundHeight = 60
		srv.curSampleTS = time.Now().Unix()

		err = srv.vote()
		require.NoError(t, err)

		require.Equal(t, 2, len(srv.voteRecords))
		require.Equal(t, srv.curRound, srv.voteRecords[srv.curRound].RoundID)
		require.Equal(t, tx.Hash(), srv.voteRecords[srv.curRound].TxHash)
		require.Equal(t, helpers.DefaultSymbols, srv.voteRecords[srv.curRound].Symbols)
		hash, err := srv.commitmentHashComputer.CommitmentHash(srv.voteRecords[srv.curRound].Reports, srv.voteRecords[srv.curRound].Salt, srv.conf.Key.Address)
		require.NoError(t, err)
		require.Equal(t, hash, srv.voteRecords[srv.curRound].CommitmentHash)

		srv.runningPlugins["template_plugin"].Close()
	})

	t.Run("test handle new symbol event", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetLastRoundBlock(nil).Return(currentRoundHeight, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)
		contractMock.EXPECT().WatchNoRevealPenalty(gomock.Any(), gomock.Any(), gomock.Any()).Return(subNoRevealEvent, nil)

		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(ChainIDPiccadilly, nil)

		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))

		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.runningPlugins))

		nSymbols := append(helpers.DefaultSymbols, "NTNETH", "NTNBTC", "NTNCNY")
		srv.handleNewSymbolsEvent(nSymbols)
		require.Equal(t, len(nSymbols)+len(BridgerSymbols), len(srv.samplingSymbols))
		srv.runningPlugins["template_plugin"].Close()
	})

	t.Run("gcRounddata", func(t *testing.T) {
		os := &OracleServer{
			voteRecords: make(map[uint64]*types.VoteRecord),
			curRound:    100,
		}

		for rd := uint64(1); rd <= 100; rd++ {
			os.voteRecords[rd] = &types.VoteRecord{
				RoundID: rd,
			}
		}

		os.gcVoteRecords()
		require.Equal(t, MaxBufferedRounds, len(os.voteRecords))

	})
}
