package server

import (
	"autonity-oracle/config"
	"autonity-oracle/helpers"
	pWrapper "autonity-oracle/plugin_wrapper"
	"autonity-oracle/types"
	"encoding/json"
	"errors"
	"io/fs"
	o "os"

	"github.com/ethereum/go-ethereum/metrics"
)

func (os *Server) PluginRuntimeManagement() {
	// load plugin configs before start them.
	newConfs, err := config.LoadPluginsConfig(os.conf.ConfigFile)
	if err != nil {
		os.logger.Error("cannot load plugin configuration", "error", err.Error())
		return
	}

	// load plugin binaries
	binaries, err := helpers.ListPlugins(os.conf.PluginDIR)
	if err != nil {
		os.logger.Error("list plugin", "error", err.Error())
		return
	}

	// shutdown the plugins which are removed, disabled or with config update.
	for name, plugin := range os.runningPlugins {
		// shutdown the plugins that were removed.
		if _, ok := binaries[name]; !ok {
			os.logger.Info("removing plugin", "name", name)
			plugin.Close()
			delete(os.runningPlugins, name)
			continue
		}

		// shutdown the plugins that are runtime disabled.
		newConf := newConfs[name]
		if newConf.Disabled {
			os.logger.Info("disabling plugin", "name", name)
			plugin.Close()
			delete(os.runningPlugins, name)
			continue
		}

		// shutdown the plugins that with config updates, they will be reloaded after the shutdown.
		if plugin.Config().Diff(&newConf) {
			os.logger.Info("updating plugin", "name", name)
			plugin.Close()
			delete(os.runningPlugins, name)
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
		if _, ok := os.keyRequiredPlugins[f.Name()]; ok && pConf.Key == "" {
			continue
		}

		os.tryToLaunchPlugin(f, pConf)
	}

	if metrics.Enabled {
		metrics.GetOrRegisterGauge("oracle/plugins", nil).Update(int64(len(os.runningPlugins)))
	}
}

func (os *Server) tryToLaunchPlugin(f fs.FileInfo, plugConf config.PluginConfig) {
	plugin, ok := os.runningPlugins[f.Name()]
	if !ok {
		os.logger.Info("new plugin discovered, going to setup it: ", f.Name(), f.Mode().String())
		pluginWrapper, err := os.setupNewPlugin(f.Name(), &plugConf)
		if err != nil {
			return
		}
		os.runningPlugins[f.Name()] = pluginWrapper
		return
	}

	if f.ModTime().After(plugin.StartTime()) || plugin.Exited() {
		os.logger.Info("replacing legacy plugin with new one: ", f.Name(), f.Mode().String())
		// stop the legacy plugin
		plugin.Close()
		delete(os.runningPlugins, f.Name())

		pluginWrapper, err := os.setupNewPlugin(f.Name(), &plugConf)
		if err != nil {
			return
		}
		os.runningPlugins[f.Name()] = pluginWrapper
	}
}

func (os *Server) setupNewPlugin(name string, conf *config.PluginConfig) (*pWrapper.PluginWrapper, error) {
	if err := os.ApplyPluginConf(name, conf); err != nil {
		os.logger.Error("apply plugin config", "error", err.Error())
		return nil, err
	}

	pluginWrapper := pWrapper.NewPluginWrapper(os.conf.LoggingLevel, name, os.conf.PluginDIR, os, conf)
	if err := pluginWrapper.Initialize(os.chainID); err != nil {
		// if the plugin states that a service key is missing, then we mark it down, thus the runtime discovery can
		// skip those plugins without a key configured.
		if errors.Is(err, types.ErrMissingServiceKey) {
			os.keyRequiredPlugins[name] = struct{}{}
		}
		os.logger.Error("cannot run plugin", "name", name, "error", err.Error())
		pluginWrapper.CleanPluginProcess()
		return nil, err
	}

	return pluginWrapper, nil
}

func (os *Server) ApplyPluginConf(name string, plugConf *config.PluginConfig) error {
	// set the plugin configuration via system env, thus the plugin can load it on startup.
	conf, err := json.Marshal(plugConf)
	if err != nil {
		os.logger.Error("cannot marshal plugin's configuration", "error", err.Error())
		return err
	}
	if err = o.Setenv(name, string(conf)); err != nil {
		os.logger.Error("cannot set plugin configuration via system ENV")
		return err
	}
	return nil
}
