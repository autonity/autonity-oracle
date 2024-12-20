package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	client "autonity-oracle/plugins/simulator_plugin/common"
	"os"
)

var defaultConfig = config.PluginConfig{
	Name:               "simulator_plugin",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "simfeed.piccadilly.autonity.org",
	Timeout:            10, //10s
	DataUpdateInterval: 10, //10s
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, client.NewSIMClient(conf), client.Version, common.ChainIDPiccadilly)
	defer adapter.Close()
	common.PluginServe(adapter)
}
