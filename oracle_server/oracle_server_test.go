package oracleserver

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	cMock "autonity-oracle/contract_binder/contract/mock"
	"autonity-oracle/helpers"
	"autonity-oracle/types"
	"autonity-oracle/types/mock"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	tp "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"io/ioutil" //nolint
	"math/big"
	"os"
	"strings"
	"testing"
	"time"
)

func TestOracleDecimals(t *testing.T) {
	decimals := uint8(18)
	precision := decimal.NewFromBigInt(common.Big1, int32(decimals))
	require.Equal(t, "1000000000000000000", precision.String())
}

func TestOracleServer(t *testing.T) {
	currentRound := new(big.Int).SetUint64(1)
	precision := config.OracleDecimals
	votePeriod := new(big.Int).SetUint64(30)
	var subRoundEvent event.Subscription
	var subSymbolsEvent event.Subscription
	var subPenalizeEvent event.Subscription
	os.Setenv("KEY.FILE", "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe") //nolint
	os.Setenv("PLUGIN.DIR", "../plugins/template_plugin/bin")                                                                    //nolint
	os.Setenv("PLUGIN.CONF", "../test_data/plugins-conf.yml")                                                                    //nolint
	defer os.Clearenv()
	conf := config.MakeConfig()

	t.Run("test init oracle server with oracle contract states", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(config.DefaultSymbols, nil)
		contractMock.EXPECT().GetDecimals(nil).Return(precision, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)

		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, config.DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))
		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.pluginSet))
		require.Equal(t, "template_plugin", srv.pluginSet["template_plugin"].Name())
		srv.pluginSet["template_plugin"].Close()
	})

	t.Run("test pre-sampling happy case", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		chainHeight := uint64(56)
		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(config.DefaultSymbols, nil)
		contractMock.EXPECT().GetDecimals(nil).Return(precision, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().BlockNumber(gomock.Any()).AnyTimes().Return(chainHeight, nil)

		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)

		ts := time.Now().Unix()
		srv.curSampleTS = uint64(ts)
		srv.curSampleHeight = uint64(30)

		for sec := ts; sec < ts+15; sec++ {
			err := srv.handlePreSampling(sec)
			require.NoError(t, err)
			time.Sleep(time.Second)
		}

		roundData, err := srv.buildRoundData(1)
		require.NoError(t, err)
		require.Equal(t, uint64(1), roundData.RoundID)
		require.Equal(t, config.DefaultSymbols, roundData.Symbols)
		require.Equal(t, len(config.DefaultSymbols), len(roundData.Prices))
		require.Equal(t, true, helpers.ResolveSimulatedPrice(NTNUSD).Equal(roundData.Prices[NTNUSD].Price))
		require.Equal(t, true, helpers.ResolveSimulatedPrice(ATNUSD).Equal(roundData.Prices[ATNUSD].Price))
		t.Log(roundData)
		srv.gcDataSamples()
		srv.pluginSet["template_plugin"].Close()
	})

	t.Run("test round vote happy case, with commitment and round data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		chainHeight := uint64(55)

		var voters []common.Address
		voters = append(voters, conf.Key.Address)
		price := contract.IOracleRoundData{
			Round:     currentRound,
			Price:     new(big.Int).SetUint64(0),
			Timestamp: new(big.Int).SetUint64(10000),
			Success:   true,
		}

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).AnyTimes().Return(config.DefaultSymbols, nil)
		contractMock.EXPECT().GetDecimals(nil).Return(precision, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().GetVoters(nil).Return(voters, nil)
		contractMock.EXPECT().GetRoundData(nil, new(big.Int).SetUint64(1), gomock.Any()).AnyTimes().Return(price, nil)
		contractMock.EXPECT().LatestRoundData(nil, gomock.Any()).AnyTimes().Return(price, nil)

		txdata := &tp.DynamicFeeTx{ChainID: new(big.Int).SetUint64(1000), Nonce: 1}
		tx := tp.NewTx(txdata)
		contractMock.EXPECT().Vote(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tx, nil)

		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().BlockNumber(gomock.Any()).AnyTimes().Return(chainHeight, nil)
		l1Mock.EXPECT().SyncProgress(gomock.Any()).Return(nil, nil)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(new(big.Int).SetUint64(1000), nil)
		l1Mock.EXPECT().BalanceAt(gomock.Any(), gomock.Any(), gomock.Any()).Return(AlertBalance, nil)
		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)

		// prepare last round data.
		prices := make(types.PriceBySymbol)
		for _, s := range config.DefaultSymbols {
			prices[s] = types.Price{
				Timestamp: 0,
				Symbol:    s,
				Price:     helpers.ResolveSimulatedPrice(s),
			}
		}

		roundData, err := srv.assembleReportData(srv.curRound, config.DefaultSymbols, prices)
		require.NoError(t, err)
		srv.roundData[srv.curRound] = roundData

		// pre-sampling with data.
		ts := time.Now().Unix()
		srv.curSampleTS = uint64(ts)
		srv.curSampleHeight = uint64(30)
		for sec := ts; sec < ts+15; sec++ {
			err = srv.handlePreSampling(time.Now().Unix())
			require.NoError(t, err)
			time.Sleep(time.Second)
		}

		// handle vote event that change to next round with
		srv.curRound = srv.curRound + 1
		srv.curSampleHeight = 60
		srv.curSampleTS = uint64(time.Now().Unix())

		err = srv.handleRoundVote()
		require.NoError(t, err)

		require.Equal(t, 2, len(srv.roundData))
		require.Equal(t, srv.curRound, srv.roundData[srv.curRound].RoundID)
		require.Equal(t, tx.Hash(), srv.roundData[srv.curRound].Tx.Hash())
		require.Equal(t, config.DefaultSymbols, srv.roundData[srv.curRound].Symbols)
		hash, err := srv.commitmentHashComputer.CommitmentHash(srv.roundData[srv.curRound].Reports, srv.roundData[srv.curRound].Salt, srv.key.Address)
		require.NoError(t, err)
		require.Equal(t, hash, srv.roundData[srv.curRound].CommitmentHash)

		srv.pluginSet["template_plugin"].Close()
	})

	t.Run("test handle new symbol event", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(config.DefaultSymbols, nil)
		contractMock.EXPECT().GetDecimals(nil).Return(precision, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)

		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, config.DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))
		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.pluginSet))

		nSymbols := append(config.DefaultSymbols, "NTNETH", "NTNBTC", "NTNCNY")
		srv.handleNewSymbolsEvent(nSymbols)
		require.Equal(t, len(nSymbols)+len(config.BridgerSymbols), len(srv.samplingSymbols))
		srv.pluginSet["template_plugin"].Close()
	})

	t.Run("test plugin runtime discovery, add new plugin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(config.DefaultSymbols, nil)
		contractMock.EXPECT().GetDecimals(nil).Return(precision, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)

		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, config.DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))
		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.pluginSet))

		// add a new plugin into the plugin directory.
		clones, err := clonePlugins(srv.pluginDIR, "cloned", srv.pluginDIR)
		require.NoError(t, err)
		defer func() {
			for _, f := range clones {
				err := os.Remove(f)
				require.NoError(t, err)
			}
		}()

		srv.PluginRuntimeDiscovery()
		require.Equal(t, 2, len(srv.pluginSet))
		require.Equal(t, "template_plugin", srv.pluginSet["template_plugin"].Name())
		require.Equal(t, "clonedtemplate_plugin", srv.pluginSet["clonedtemplate_plugin"].Name())
		srv.pluginSet["template_plugin"].Close()
		srv.pluginSet["clonedtemplate_plugin"].Close()
	})

	t.Run("test plugin runtime discovery, upgrade plugin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(config.DefaultSymbols, nil)
		contractMock.EXPECT().GetDecimals(nil).Return(precision, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)

		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, config.DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))
		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.pluginSet))
		firstStart := srv.pluginSet["template_plugin"].StartTime()

		// cpy and replace the legacy plugins
		err := replacePlugins(srv.pluginDIR)
		require.NoError(t, err)

		srv.PluginRuntimeDiscovery()

		require.Equal(t, 1, len(srv.pluginSet))
		require.Equal(t, "template_plugin", srv.pluginSet["template_plugin"].Name())
		require.Greater(t, srv.pluginSet["template_plugin"].StartTime(), firstStart)
		srv.pluginSet["template_plugin"].Close()
	})

	t.Run("gcRounddata", func(t *testing.T) {
		os := &OracleServer{
			roundData: make(map[uint64]*types.RoundData),
			curRound:  100,
		}

		for rd := uint64(1); rd <= 100; rd++ {
			os.roundData[rd] = &types.RoundData{
				RoundID: rd,
			}
		}

		os.gcRoundData()
		require.Equal(t, types.MaxBufferedRounds, len(os.roundData))

	})
}

// clone plugins from a src directory to new directory by adding prefix in the name of each binary, and return the cloned
// new file names and an error.
func clonePlugins(pluginDIR string, clonePrefix string, destDir string) ([]string, error) {

	var clonedPlugins []string
	files, err := helpers.ListPlugins(pluginDIR)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// read srcFile
		srcContent, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", pluginDIR, file.Name()))
		if err != nil {
			return clonedPlugins, err
		}

		// create dstFile and copy the content
		newPlugin := fmt.Sprintf("%s/%s%s", destDir, clonePrefix, file.Name())
		err = ioutil.WriteFile(newPlugin, srcContent, file.Mode())
		if err != nil {
			return clonedPlugins, err
		}
		clonedPlugins = append(clonedPlugins, newPlugin)
	}
	return clonedPlugins, nil
}

func replacePlugins(pluginDir string) error {
	rawPlugins, err := helpers.ListPlugins(pluginDir)
	if err != nil {
		return err
	}

	clonePrefix := "clone"
	clonedPlugins, err := clonePlugins(pluginDir, clonePrefix, fmt.Sprintf("%s/..", pluginDir))
	if err != nil {
		return err
	}

	for _, file := range clonedPlugins {
		for _, info := range rawPlugins {
			if strings.Contains(file, info.Name()) {
				err := os.Rename(file, fmt.Sprintf("%s/%s", pluginDir, info.Name()))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
