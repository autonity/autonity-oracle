package main

import (
	"autonity-oracle/config"
	"autonity-oracle/plugins/common"
	client "autonity-oracle/plugins/simulator_plugin/common"
	"autonity-oracle/types"
	"os"
)

var defaultConfig = config.PluginConfig{
	Name:               "simulator_plugin",
	Key:                "",
	Scheme:             "https",
	Endpoint:           "simfeed.bakerloo.autonity.org",
	Confidence:         types.BaseConfidence,  // range from [1, 100], the higher, the better data quality is.
	Timeout:            10, //10s
	DataUpdateInterval: 10, //10s
}

func main() {
	conf := common.ResolveConf(os.Args[0], &defaultConfig)
	adapter := common.NewPlugin(conf, client.NewSIMClient(conf), client.Version, types.SrcCEX, common.ChainIDBakerloo)
	defer adapter.Close()
	common.PluginServe(adapter)
}
