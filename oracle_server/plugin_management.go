package oracleserver

import (
	pWrapper "autonity-oracle/plugin_wrapper"
	"io/fs"
	"io/ioutil"
)

func (os *OracleServer) PluginRuntimeDiscovery() {
	binaries := os.listPluginDIR()

	for _, file := range binaries {
		os.createPlugin(file)
	}
}

func (os *OracleServer) createPlugin(f fs.FileInfo) {
	os.pluginLock.Lock()
	defer os.pluginLock.Unlock()
	plugin, ok := os.pluginSet[f.Name()]
	if !ok {
		os.logger.Info("** New pWrapper discovered, going to setup it: ", f.Name(), f.Mode().String())
		pluginWrapper := pWrapper.NewPluginWrapper(f.Name(), os.pluginDIR)
		pluginWrapper.Initialize()
		os.pluginSet[f.Name()] = pluginWrapper
		os.logger.Info("** New pWrapper on ready: ", f.Name())
		return
	}

	if f.ModTime().After(plugin.StartTime()) {
		os.logger.Info("*** Replacing legacy pWrapper with new one: ", f.Name(), f.Mode().String())
		// stop the legacy plugins process, disconnect rpc connection and release memory.
		plugin.Close()
		delete(os.pluginSet, f.Name())
		pluginWrapper := pWrapper.NewPluginWrapper(f.Name(), os.pluginDIR)
		pluginWrapper.Initialize()
		os.pluginSet[f.Name()] = pluginWrapper
		os.logger.Info("*** Finnish the replacement of pWrapper: ", f.Name())
	}
}

func (os *OracleServer) listPluginDIR() []fs.FileInfo {
	var plugins []fs.FileInfo

	files, err := ioutil.ReadDir(os.pluginDIR)
	if err != nil {
		os.logger.Error("cannot read from pWrapper store, please double check plugins are saved in the directory: ",
			os.pluginDIR, err.Error())
		return nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		plugins = append(plugins, file)
	}
	return plugins
}
