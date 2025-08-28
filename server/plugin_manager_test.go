package server

import (
	"autonity-oracle/config"
	cMock "autonity-oracle/contract_binder/contract/mock"
	"autonity-oracle/helpers"
	"autonity-oracle/types/mock"
	"fmt"
	"io"
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
		PluginDir:          "../plugins/template_plugin/bin",
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
		srv := NewServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))

		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.runningPlugins))

		// add a new plugin into the plugin directory.
		clones, err := clonePlugins(srv.conf.PluginDir, "cloned", srv.conf.PluginDir)
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
		srv := NewServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))

		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.runningPlugins))
		firstStart := srv.runningPlugins["template_plugin"].StartTime()

		// cpy and replace the legacy plugins
		err := replacePlugins(srv.conf.PluginDir)
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
		srv := NewServer(conf, dialerMock, l1Mock, contractMock)
		require.Equal(t, currentRound.Uint64(), srv.curRound)
		require.Equal(t, DefaultSampledSymbols, srv.samplingSymbols)
		require.Equal(t, true, srv.pricePrecision.Equal(decimal.NewFromBigInt(common.Big1, int32(precision))))

		require.Equal(t, votePeriod.Uint64(), srv.votePeriod)
		require.Equal(t, 1, len(srv.runningPlugins))

		// Backup the existing plugins
		backupDir := t.TempDir()
		err := os.MkdirAll(backupDir, 0755) // Create backup directory if it doesn't exist
		require.NoError(t, err)

		// Copy existing plugins to the backup directory
		files, err := os.ReadDir(srv.conf.PluginDir)
		require.NoError(t, err)

		for _, file := range files {
			info, err := file.Info() //nolint
			require.NoError(t, err)

			if !file.IsDir() && helpers.IsExecOwnerGroup(info.Mode()) {
				src := filepath.Join(srv.conf.PluginDir, file.Name())
				dst := filepath.Join(backupDir, file.Name())

				// FIXED: Use file copy instead of hard link
				err = copyFile(src, dst, info.Mode())
				require.NoError(t, err)
			}
		}

		// Defer the recovery action to restore the removed plugins
		defer func() {
			// Restore the plugins from the backup directory
			files, err = os.ReadDir(backupDir)
			require.NoError(t, err)

			for _, file := range files {
				info, err := file.Info() //nolint
				require.NoError(t, err)

				src := filepath.Join(backupDir, file.Name())
				dst := filepath.Join(srv.conf.PluginDir, file.Name())

				// FIXED: Use file copy instead of hard link
				err = copyFile(src, dst, info.Mode())
				require.NoError(t, err)
			}
		}()

		// Remove the plugins
		err = removePlugins(srv.conf.PluginDir)
		require.NoError(t, err)

		srv.PluginRuntimeManagement()
		require.Equal(t, 0, len(srv.runningPlugins))
	})
}

// Helper function to copy files with proper permissions
func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
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
		srcContent, err := os.ReadFile(fmt.Sprintf("%s/%s", pluginDIR, file.Name()))
		if err != nil {
			return clonedPlugins, err
		}

		// create dstFile and copy the content
		newPlugin := fmt.Sprintf("%s/%s%s", destDir, clonePrefix, file.Name())
		err = os.WriteFile(newPlugin, srcContent, file.Mode())
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
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	// Iterate over the files
	for _, file := range files {
		// Skip directories
		if file.IsDir() {
			continue
		}

		// Construct the full file path
		filePath := pluginDir + "/" + file.Name()
		// Remove the file
		if err := os.Remove(filePath); err != nil {
			return err // Return error if unable to remove the file
		}
	}
	return nil
}
