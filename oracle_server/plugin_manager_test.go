package oracleserver

import (
	"autonity-oracle/config"
	cMock "autonity-oracle/contract_binder/contract/mock"
	"autonity-oracle/helpers"
	"autonity-oracle/types/mock"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestPluginManagement(t *testing.T) {
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
