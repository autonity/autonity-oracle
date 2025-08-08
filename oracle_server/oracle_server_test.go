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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"io/ioutil" //nolint
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var BridgerSymbols = []string{NTNUSDC, ATNUSDC, USDCUSD}
var DefaultSampledSymbols = []string{"AUD-USD", "CAD-USD", "EUR-USD", "GBP-USD", "JPY-USD", "SEK-USD", "ATN-USD", "NTN-USD", "NTN-ATN", "ATN-USDC", "NTN-USDC", "USDC-USD"}
var ChainIDPiccadilly = big.NewInt(65_100_004)

func TestOracleDecimals(t *testing.T) {
	decimals := uint8(18)
	precision := decimal.NewFromBigInt(common.Big1, int32(decimals))
	require.Equal(t, "1000000000000000000", precision.String())
}

// TestServerState tests the flush and loadState methods of ServerMemories.
func TestServerState(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	fileName := filepath.Join(tempDir, serverStateDumpFile)

	nodeKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	nodeAddr := crypto.PubkeyToAddress(nodeKey.PublicKey)

	// Create a ServerMemories instance
	originalState := &ServerMemories{
		OutlierRecord: OutlierRecord{
			LastPenalizedAtBlock: 1234556,
			Participant:          nodeAddr,
			Symbol:               NTNUSDC,
			Median:               uint64(100000000000),
			Reported:             uint64(200000000000),
			SlashingAmount:       uint64(300000000000),
		},
		LoggedAt: time.Now().Format(time.RFC3339),
	}

	// Test flush method
	err = originalState.flush(tempDir)
	if err != nil {
		t.Fatalf("failed to flush state: %v", err)
	}

	// Check if the file was created
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		t.Fatalf("expected file %s to be created, but it does not exist", fileName)
	}

	// Create a new ServerMemories instance to load data into
	loadedState := &ServerMemories{}

	// Test loadState method
	err = loadedState.loadState(tempDir)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	require.Equal(t, originalState, loadedState)
}

func TestOracleServer(t *testing.T) {
	currentRound := new(big.Int).SetUint64(1)
	precision := OracleDecimals
	votePeriod := new(big.Int).SetUint64(30)
	var subRoundEvent event.Subscription
	var subSymbolsEvent event.Subscription
	var subPenalizeEvent event.Subscription
	var subVoteEvent event.Subscription
	var subInvalidVoteEvent event.Subscription
	var subReportedEvent event.Subscription

	keyFile := "../test_data/keystore/UTC--2023-02-27T09-10-19.592765887Z--b749d3d83376276ab4ddef2d9300fb5ce70ebafe"
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
		contractMock.EXPECT().GetSymbols(nil).Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)

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
		contractMock.EXPECT().GetSymbols(nil).Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().BlockNumber(gomock.Any()).AnyTimes().Return(chainHeight, nil)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(ChainIDPiccadilly, nil)
		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)

		ts := time.Now().Unix()
		srv.curSampleTS = ts
		srv.curSampleHeight = uint64(30)

		for sec := ts; sec < ts+15; sec++ {
			err := srv.handlePreSampling(sec)
			require.NoError(t, err)
			time.Sleep(time.Second)
		}

		roundData, err := srv.buildRoundData(roundID)
		require.NoError(t, err)
		require.Equal(t, roundID, roundData.RoundID)
		require.Equal(t, helpers.DefaultSymbols, roundData.Symbols)
		require.Equal(t, len(helpers.DefaultSymbols), len(roundData.Prices))
		require.Equal(t, true, helpers.ResolveSimulatedPrice(NTNUSD).Equal(roundData.Prices[NTNUSD].Price))
		require.Equal(t, true, helpers.ResolveSimulatedPrice(ATNUSD).Equal(roundData.Prices[ATNUSD].Price))
		t.Log(roundData)
		srv.gcExpiredSamples()
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

		roundData, err := srv.assembleReportData(srv.curRound, helpers.DefaultSymbols, prices)
		require.NoError(t, err)
		srv.roundData[srv.curRound] = roundData

		// pre-sampling with data.
		ts := time.Now().Unix()
		srv.curSampleTS = ts
		srv.curSampleHeight = uint64(30)
		for sec := ts; sec < ts+15; sec++ {
			err = srv.handlePreSampling(time.Now().Unix())
			require.NoError(t, err)
			time.Sleep(time.Second)
		}

		// handle vote event that change to next round with
		srv.curRound = srv.curRound + 1
		srv.curSampleHeight = 60
		srv.curSampleTS = time.Now().Unix()

		err = srv.handleRoundVote()
		require.NoError(t, err)

		require.Equal(t, 2, len(srv.roundData))
		require.Equal(t, srv.curRound, srv.roundData[srv.curRound].RoundID)
		require.Equal(t, tx.Hash(), srv.roundData[srv.curRound].Tx.Hash())
		require.Equal(t, helpers.DefaultSymbols, srv.roundData[srv.curRound].Symbols)
		hash, err := srv.commitmentHashComputer.CommitmentHash(srv.roundData[srv.curRound].Reports, srv.roundData[srv.curRound].Salt, srv.conf.Key.Address)
		require.NoError(t, err)
		require.Equal(t, hash, srv.roundData[srv.curRound].CommitmentHash)

		srv.runningPlugins["template_plugin"].Close()
	})

	t.Run("test handle new symbol event", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)

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

	t.Run("test plugin runtime management, add new plugin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(ChainIDPiccadilly, nil)
		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))

		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.runningPlugins))

		// add a new plugin into the plugin directory.
		clones, err := clonePlugins(srv.conf.PluginDIR, "cloned", srv.conf.PluginDIR)
		require.NoError(t, err)
		defer func() {
			for _, f := range clones {
				err := os.Remove(f)
				require.NoError(t, err)
			}
		}()

		srv.PluginRuntimeManagement()
		require.Equal(t, 2, len(srv.runningPlugins))
		require.Equal(t, "template_plugin", srv.runningPlugins["template_plugin"].Name())
		require.Equal(t, "clonedtemplate_plugin", srv.runningPlugins["clonedtemplate_plugin"].Name())
		srv.runningPlugins["template_plugin"].Close()
		srv.runningPlugins["clonedtemplate_plugin"].Close()
	})

	t.Run("test plugin runtime management, upgrade plugin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)
		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(ChainIDPiccadilly, nil)
		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))

		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.runningPlugins))
		firstStart := srv.runningPlugins["template_plugin"].StartTime()

		// cpy and replace the legacy plugins
		err := replacePlugins(srv.conf.PluginDIR)
		require.NoError(t, err)

		srv.PluginRuntimeManagement()

		require.Equal(t, 1, len(srv.runningPlugins))
		require.Equal(t, "template_plugin", srv.runningPlugins["template_plugin"].Name())
		require.Greater(t, srv.runningPlugins["template_plugin"].StartTime(), firstStart)
		srv.runningPlugins["template_plugin"].Close()
	})

	t.Run("test plugin runtime management, remove plugin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		dialerMock := mock.NewMockDialer(ctrl)
		contractMock := cMock.NewMockContractAPI(ctrl)
		contractMock.EXPECT().GetRound(nil).Return(currentRound, nil)
		contractMock.EXPECT().GetSymbols(nil).Return(helpers.DefaultSymbols, nil)
		contractMock.EXPECT().GetVotePeriod(nil).Return(votePeriod, nil)
		contractMock.EXPECT().WatchNewRound(gomock.Any(), gomock.Any()).Return(subRoundEvent, nil)
		contractMock.EXPECT().WatchNewSymbols(gomock.Any(), gomock.Any()).Return(subSymbolsEvent, nil)
		contractMock.EXPECT().WatchPenalized(gomock.Any(), gomock.Any(), gomock.Any()).Return(subPenalizeEvent, nil)
		contractMock.EXPECT().WatchSuccessfulVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subVoteEvent, nil)
		contractMock.EXPECT().WatchTotalOracleRewards(gomock.Any(), gomock.Any()).Return(subReportedEvent, nil)
		contractMock.EXPECT().WatchInvalidVote(gomock.Any(), gomock.Any(), gomock.Any()).Return(subInvalidVoteEvent, nil)

		l1Mock := mock.NewMockBlockchain(ctrl)
		l1Mock.EXPECT().ChainID(gomock.Any()).Return(ChainIDPiccadilly, nil)
		srv := NewOracleServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))

		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.runningPlugins))

		// Backup the existing plugins
		backupDir := "/tmp/plugin_backup"   // Temporary directory for backup
		err := os.MkdirAll(backupDir, 0755) // Create backup directory if it doesn't exist
		require.NoError(t, err)

		// Copy existing plugins to the backup directory
		files, err := ioutil.ReadDir(srv.conf.PluginDIR)
		require.NoError(t, err)
		for _, file := range files {
			if !file.IsDir() && helpers.IsExecOwnerGroup(file.Mode()) {
				src := filepath.Join(srv.conf.PluginDIR, file.Name())
				dst := filepath.Join(backupDir, file.Name())
				err = os.Link(src, dst) // Use os.Link for hard link; use os.Copy for actual copy if needed
				require.NoError(t, err)
			}
		}

		// Defer the recovery action to restore the removed plugins
		defer func() {
			// Restore the plugins from the backup directory
			files, err = ioutil.ReadDir(backupDir)
			require.NoError(t, err)
			for _, file := range files {
				src := filepath.Join(backupDir, file.Name())
				dst := filepath.Join(srv.conf.PluginDIR, file.Name())
				err = os.Link(src, dst) // Use os.Link for hard link; use os.Copy for actual copy if needed
				require.NoError(t, err)
			}
			os.RemoveAll(backupDir) // Clean up the backup directory
		}()

		// Remove the plugins
		err = removePlugins(srv.conf.PluginDIR)
		require.NoError(t, err)

		srv.PluginRuntimeManagement()

		require.Equal(t, 0, len(srv.runningPlugins))
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
		require.Equal(t, MaxBufferedRounds, len(os.roundData))

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

// removePlugins removes all executable plugin binaries from the specified plugin directory.
func removePlugins(pluginDir string) error {
	// Read the directory
	files, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	// Iterate over the files
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Check if the file is an executable
		if helpers.IsExecOwnerGroup(file.Mode()) {
			// Construct the full file path
			filePath := pluginDir + "/" + file.Name()
			// Remove the file
			if err := os.Remove(filePath); err != nil {
				return err // Return error if unable to remove the file
			}
		}
	}
	return nil
}

// TestComputeConfidence tests the ComputeConfidence function.
func TestComputeConfidence(t *testing.T) {
	tests := []struct {
		symbol       string
		numOfSamples int
		strategy     int
		expected     uint8
	}{
		// Forex symbols with ConfidenceStrategyFixed, max confidence are expected.
		{"AUD-USD", 1, config.ConfidenceStrategyFixed, MaxConfidence},
		{"CAD-USD", 2, config.ConfidenceStrategyFixed, MaxConfidence},
		{"EUR-USD", 3, config.ConfidenceStrategyFixed, MaxConfidence},
		{"GBP-USD", 4, config.ConfidenceStrategyFixed, MaxConfidence},
		{"JPY-USD", 5, config.ConfidenceStrategyFixed, MaxConfidence},
		{"SEK-USD", 10, config.ConfidenceStrategyFixed, MaxConfidence},

		// Forex symbols with ConfidenceStrategyLinear
		{"AUD-USD", 1, config.ConfidenceStrategyLinear, uint8(BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, 1)))}, //nolint
		{"CAD-USD", 2, config.ConfidenceStrategyLinear, uint8(BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, 2)))}, //nolint
		{"EUR-USD", 3, config.ConfidenceStrategyLinear, uint8(BaseConfidence + SourceScalingFactor*uint64(math.Pow(1.75, 3)))}, //nolint
		{"GBP-USD", 4, config.ConfidenceStrategyLinear, MaxConfidence},
		{"JPY-USD", 5, config.ConfidenceStrategyLinear, MaxConfidence},
		{"SEK-USD", 10, config.ConfidenceStrategyLinear, MaxConfidence},

		// Non-forex symbols with ConfidenceStrategyLinear, max confidence are expected.
		{"ATN-USD", 1, config.ConfidenceStrategyLinear, MaxConfidence},
		{"NTN-USD", 1, config.ConfidenceStrategyLinear, MaxConfidence},
		{"NTN-ATN", 1, config.ConfidenceStrategyLinear, MaxConfidence},

		// Non-forex symbols with ConfidenceStrategyFixed, max confidence are expected.
		{"ATN-USD", 1, config.ConfidenceStrategyFixed, MaxConfidence},
		{"NTN-USD", 1, config.ConfidenceStrategyFixed, MaxConfidence},
		{"NTN-ATN", 1, config.ConfidenceStrategyFixed, MaxConfidence},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			got := ComputeConfidence(tt.symbol, tt.numOfSamples, tt.strategy)
			if got != tt.expected {
				t.Errorf("ComputeConfidence(%q, %d, %d) = %d; want %d", tt.symbol, tt.numOfSamples, tt.strategy, got, tt.expected)
			}
		})
	}
}
