package server

import (
	"autonity-oracle/config"
	"autonity-oracle/helpers"
	"autonity-oracle/monitor"
	pWrapper "autonity-oracle/plugin_wrapper"
	"autonity-oracle/types"
	"encoding/json"
	"errors"
	"io/fs"
	"math/big"
	o "os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/metrics"
	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/go-hclog"
	"github.com/shopspring/decimal"
)

var oneMinInterval = 1 * time.Minute

type IPluginManager interface {
	Start()
	Stop()
	SelectSamples(symbol string, target int64) ([]decimal.Decimal, []*big.Int, error)
}

type PluginManager struct {
	logger hclog.Logger

	regularTicker *time.Ticker // the clock source to trigger the 10s interval job.
	address       common.Address
	configFile    string      // be watching for
	pluginDir     string      // be watching for
	loggingLvl    hclog.Level // logging lvl.

	doneCh         chan struct{}
	subSampleEvent types.SampleEventSubscriber // subscriber of the sampling event.

	rwMutex            sync.RWMutex
	runningPlugins     map[string]*pWrapper.PluginWrapper // the plugin clients that connect with different adapters.
	keyRequiredPlugins map[string]struct{}                // saving those plugins which require a key granted by data provider

	configWatcher  *fsnotify.Watcher // config file watcher which watches the config changes.
	pluginsWatcher *fsnotify.Watcher // plugins watcher which watches the changes of plugins and the plugins' configs.
	chainID        int64             // ChainID saves the L1 chain ID, it is used for plugin compatibility check.
}

func (m *PluginManager) Start() {
	for {
		select {
		case <-m.doneCh:
			m.regularTicker.Stop()
			if m.pluginsWatcher != nil {
				m.pluginsWatcher.Close() //nolint
			}

			if m.configWatcher != nil {
				m.configWatcher.Close() //nolint
			}
			m.logger.Info("plugin manager is stopping...")
			return
		case err := <-m.configWatcher.Errors:
			if err != nil {
				m.logger.Error("oracle config file watcher err", "err", err.Error())
			}
		case err := <-m.pluginsWatcher.Errors:
			if err != nil {
				m.logger.Error("plugin watcher errors", "err", err.Error())
			}
		case fsEvent, ok := <-m.configWatcher.Events:
			if !ok {
				m.logger.Error("config watcher channel has been closed")
				return
			}
			m.handleConfigEvent(fsEvent)

		case fsEvent, ok := <-m.pluginsWatcher.Events:
			if !ok {
				m.logger.Error("plugin watcher channel has been closed")
				return
			}

			m.logger.Info("watched plugins fs event", "file", fsEvent.Name, "event", fsEvent.Op.String())
			// updates on the watched plugin directory will trigger plugin management.
			m.PluginRuntimeManagement()
		case <-m.regularTicker.C:
			if metrics.Enabled {
				metrics.GetOrRegisterGauge(monitor.PluginMetric, nil).Update(int64(m.numOfPlugins()))
			}
		}
	}
}

func (m *PluginManager) Stop() {
	m.doneCh <- struct{}{}
	m.stopAllPlugins()
	m.logger.Info("plugin manager is stopped")
}

func (m *PluginManager) stopAllPlugins() {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	for _, c := range m.runningPlugins {
		p := c
		p.Close()
	}
}

// SelectSamples is called on round event to select the most optimal samples from different sources of a symbol for voting.
func (m *PluginManager) SelectSamples(symbol string, target int64) ([]decimal.Decimal, []*big.Int) {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()

	var prices []decimal.Decimal
	var volumes []*big.Int
	for _, plugin := range m.runningPlugins {
		p, err := plugin.SelectSample(symbol, target)
		if err != nil {
			m.logger.Warn("select sample err", "err", err.Error(), "plugin", plugin.Name(), "symbol", symbol, "target", target)
			continue
		}
		prices = append(prices, p.Price)
		volumes = append(volumes, p.Volume)
	}

	return prices, volumes
}

func (m *PluginManager) numOfPlugins() int {
	m.rwMutex.RLock()
	defer m.rwMutex.RUnlock()
	return len(m.runningPlugins)
}

func NewPluginManager(configFile string, pluginDir string, loggingLvl hclog.Level, sub types.SampleEventSubscriber,
	chainID int64, address common.Address, pluginConfigs map[string]config.PluginConfig) *PluginManager {

	m := &PluginManager{
		configFile:         configFile,
		pluginDir:          pluginDir,
		loggingLvl:         loggingLvl,
		regularTicker:      time.NewTicker(oneMinInterval),
		doneCh:             make(chan struct{}),
		subSampleEvent:     sub,
		runningPlugins:     make(map[string]*pWrapper.PluginWrapper),
		keyRequiredPlugins: make(map[string]struct{}),
		chainID:            chainID,
		address:            address,
	}

	m.logger = hclog.New(&hclog.LoggerOptions{
		Name:   "plugin_manager" + address.String(),
		Output: o.Stdout,
		Level:  m.loggingLvl,
	})

	// discover plugins from plugin dir at startup.
	binaries, err := helpers.ListPlugins(m.pluginDir)
	if len(binaries) == 0 || err != nil {
		// to stop the service on the start once there is no plugin in the db.
		m.logger.Error("no plugin discovered", "plugin-dir", m.pluginDir)
		o.Exit(1)
	}

	// load plugins with the initial plugin configs.
	for _, file := range binaries {
		f := file
		pConf := pluginConfigs[f.Name()]
		if pConf.Disabled {
			continue
		}
		m.tryToLaunchPlugin(f, pConf)
	}

	// subscribe FS notifications of the watched plugins.
	pluginsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		m.logger.Error("cannot create fsnotify watcher", "error", err)
		o.Exit(1)
	}

	err = pluginsWatcher.Add(m.pluginDir)
	if err != nil {
		m.logger.Error("cannot watch plugin dir", "error", err)
		o.Exit(1)
	}
	m.pluginsWatcher = pluginsWatcher

	// subscribe FS notification of the watched config file.
	configWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		m.logger.Error("cannot create fsnotify watcher", "error", err)
		o.Exit(1)
	}

	dir := filepath.Dir(m.configFile)
	err = configWatcher.Add(dir) // Watch parent directory
	if err != nil {
		m.logger.Error("cannot watch oracle config directory", "error", err)
		o.Exit(1)
	}

	m.configWatcher = configWatcher
	return m
}

func (m *PluginManager) PluginRuntimeManagement() {
	// load plugin configs before start them.
	newConfs, err := config.LoadPluginsConfig(m.configFile)
	if err != nil {
		m.logger.Error("cannot load plugin configuration", "error", err.Error())
		return
	}

	// load plugin binaries
	binaries, err := helpers.ListPlugins(m.pluginDir)
	if err != nil {
		m.logger.Error("list plugin", "error", err.Error())
		return
	}

	m.rwMutex.Lock()
	defer m.rwMutex.Unlock()
	// shutdown the plugins which are removed, disabled or with config update.
	for name, plugin := range m.runningPlugins {
		// shutdown the plugins that were removed.
		if _, ok := binaries[name]; !ok {
			m.logger.Info("removing plugin", "name", name)
			plugin.Close()
			delete(m.runningPlugins, name)
			continue
		}

		// shutdown the plugins that are runtime disabled.
		newConf := newConfs[name]
		if newConf.Disabled {
			m.logger.Info("disabling plugin", "name", name)
			plugin.Close()
			delete(m.runningPlugins, name)
			continue
		}

		// shutdown the plugins that with config updates, they will be reloaded after the shutdown.
		if plugin.Config().Diff(&newConf) {
			m.logger.Info("resetting plugin config", "name", name)
			plugin.Close()
			delete(m.runningPlugins, name)
		}
	}

	// try to load new plugins.
	for _, file := range binaries {
		f := file
		pConf := newConfs[f.Name()]

		if pConf.Disabled {
			continue
		}

		// skip to set up plugins until there is a service key is presented at plugin-confs.yml
		if _, ok := m.keyRequiredPlugins[f.Name()]; ok && pConf.Key == "" {
			continue
		}

		m.tryToLaunchPlugin(f, pConf)
	}

	if metrics.Enabled {
		metrics.GetOrRegisterGauge("oracle/plugins", nil).Update(int64(len(m.runningPlugins)))
	}
}

// After the setup, this function assumes the caller already hold the RWMutex.
func (m *PluginManager) tryToLaunchPlugin(f fs.FileInfo, plugConf config.PluginConfig) {
	plugin, ok := m.runningPlugins[f.Name()]
	if !ok {
		m.logger.Info("new plugin discovered, going to setup it: ", f.Name(), f.Mode().String())
		pluginWrapper, err := m.setupNewPlugin(f.Name(), &plugConf)
		if err != nil {
			return
		}
		m.runningPlugins[f.Name()] = pluginWrapper
		return
	}

	if f.ModTime().After(plugin.StartTime()) || plugin.Exited() {
		m.logger.Info("replacing legacy plugin with new one: ", f.Name(), f.Mode().String())
		// stop the legacy plugin
		plugin.Close()
		delete(m.runningPlugins, f.Name())

		pluginWrapper, err := m.setupNewPlugin(f.Name(), &plugConf)
		if err != nil {
			return
		}
		m.runningPlugins[f.Name()] = pluginWrapper
	}
}

func (m *PluginManager) setupNewPlugin(name string, conf *config.PluginConfig) (*pWrapper.PluginWrapper, error) {
	if err := m.applyPluginConf(name, conf); err != nil {
		m.logger.Error("apply plugin config", "error", err.Error())
		return nil, err
	}

	pluginWrapper := pWrapper.NewPluginWrapper(m.loggingLvl, name, m.pluginDir, m.subSampleEvent, conf)
	if err := pluginWrapper.Initialize(m.chainID); err != nil {
		// if the plugin states that a service key is missing, then we mark it down, thus the runtime discovery can
		// skip those plugins without a key configured.
		if errors.Is(err, types.ErrMissingServiceKey) {
			m.keyRequiredPlugins[name] = struct{}{}
		}
		m.logger.Error("cannot run plugin", "name", name, "error", err.Error())
		pluginWrapper.CleanPluginProcess()
		return nil, err
	}

	return pluginWrapper, nil
}

func (m *PluginManager) applyPluginConf(name string, plugConf *config.PluginConfig) error {
	// set the plugin configuration via system env, thus the plugin can load it on startup.
	conf, err := json.Marshal(plugConf)
	if err != nil {
		m.logger.Error("cannot marshal plugin's configuration", "error", err.Error())
		return err
	}
	if err = o.Setenv(name, string(conf)); err != nil {
		m.logger.Error("cannot set plugin configuration via system ENV")
		return err
	}
	return nil
}

func (m *PluginManager) handleConfigEvent(ev fsnotify.Event) {
	// filter unwatched files in the dir.
	if filepath.Base(ev.Name) != filepath.Base(m.configFile) {
		return
	}

	switch {
	// tools like sed issues write event for the updates.
	case ev.Op&fsnotify.Write > 0:
		// apply plugin config changes.
		m.logger.Info("config file content changed", "file", ev.Name)
		m.PluginRuntimeManagement()

	// tools like vim or vscode issues rename, chmod and remove events for the update for an atomic change mode.
	case ev.Op&fsnotify.Rename > 0:
		m.logger.Info("config file changed", "file", ev.Name)
		m.PluginRuntimeManagement()
	}
}
