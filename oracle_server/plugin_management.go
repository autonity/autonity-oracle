package oracleserver

import (
	cryptoprovider "autonity-oracle/plugin_wrapper"
	"io/fs"
	"io/ioutil"
)

func (os *OracleServer) PluginRuntimeDiscovery() {
	binaries := os.listPluginDIR()

	for _, file := range binaries {
		plugin, ok := os.pluginSet[file.Name()]
		if !ok {
			os.logger.Info("** New plugin discovered, going to setup it: ", file.Name(), file.Mode().String())
			os.createPlugin(file.Name())
			os.logger.Info("** New plugin on ready: ", file.Name())
			continue
		}

		if file.ModTime().After(plugin.StartTime()) {
			os.logger.Info("*** Replacing legacy plugin with new one: ", file.Name(), file.Mode().String())
			// stop the legacy plugins process, disconnect rpc connection and release memory.
			plugin.Close()
			delete(os.pluginSet, file.Name())
			os.createPlugin(file.Name())
			os.logger.Info("*** Finnish the replacement of plugin: ", file.Name())
		}
	}
}

func (os *OracleServer) createPlugin(name string) {
	pool := os.dataSet.GetDataCache(name)
	if pool == nil {
		pool = os.dataSet.AddDataCache(name)
	}

	pluginWrapper := cryptoprovider.NewPluginWrapper(name, os.pluginDIR, pool)
	pluginWrapper.Initialize()
	os.pluginSet[name] = pluginWrapper
}

func (os *OracleServer) listPluginDIR() []fs.FileInfo {
	var plugins []fs.FileInfo

	files, err := ioutil.ReadDir(os.pluginDIR)
	if err != nil {
		os.logger.Error("cannot read from plugin store, please double check plugins are saved in the directory: ",
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
