package config

import (
	"github.com/namsral/flag"
	"github.com/shopspring/decimal"
	"strings"
)

var (
	// DefSimulatorConf contains a list of data items with the pattern of SYMBOL:StartingDataPoint:DataDistributionRateRange
	// with each separated by a "|". To tune the starting data point for symbols, one can set the values on demand with CLI
	// flags or system environment variables with such data pattern, or tune it via HTTP rpc call during the runtime, please
	// refer to the readme.
	DefSimulatorConf = "NTNUSD:7.0:0.01|NTNAUD:9.856:0.01|NTNCAD:9.333:0.01|NTNEUR:6.8369:0.01|NTNGBP:5.691:0.01|NTNJPY:128.205:0.01|NTNSEK:72.163:0.01|" +
		"AUD/USD:0.67:0.01|CAD/USD:0.74:0.01|EUR/USD:1.086:0.01|GBP/USD:1.25:0.01|JPY/USD:0.0073:0.01|SEK/USD:0.096:0.01|ATN/USD:1.0:0.001|NTN/USD:7.0:0.01|NTN/ATN:7.0:0.01"
	DefSimulatorPort = 50991 // default port bind with the http service in the simulator.
	DefPlaybook      = ""    // the default playbook file used to replay data points in the generator.
	DefTimeout       = 0     // the default timeout simulated when processing a http request.
)

type RandGeneratorConfig struct {
	ReferenceDataPoint decimal.Decimal
	DistributionRate   decimal.Decimal
}

type SimulatorConfig struct {
	Port            int
	Playbook        string
	SimulatorConf   map[string]*RandGeneratorConfig
	SimulateTimeOut int
}

func MakeSimulatorConfig() *SimulatorConfig {
	var simulateTimeOut int
	var port int
	var simulatorConf string
	var playbook string

	flag.IntVar(&simulateTimeOut, "sim_timeout", DefTimeout, "The timeout in seconds to be simulated in processing http request")
	flag.IntVar(&port, "sim_http_port", DefSimulatorPort, "The HTTP rpc port to be bind for binance_simulator simulator")
	flag.StringVar(&playbook, "sim_playbook_file", DefPlaybook, "The .csv file which contains datapoint for symbols.")
	flag.StringVar(&simulatorConf, "sim_symbol_config", DefSimulatorConf,
		"The list of data items with the pattern of SYMBOL:StartingDataPoint:DataDistributionRateRange with each separated by a \"|\"")

	flag.Parse()

	conf := ParseSimulatorConf(simulatorConf)

	println("\n\n\n\tRunning simulator with conf: ", simulatorConf)
	println("\tRunning simulator only with playbook if playbook is configured: ", playbook)

	return &SimulatorConfig{
		Port:            port,
		Playbook:        playbook,
		SimulatorConf:   conf,
		SimulateTimeOut: simulateTimeOut,
	}
}

func ParseSimulatorConf(conf string) map[string]*RandGeneratorConfig {
	result := make(map[string]*RandGeneratorConfig)
	items := strings.Split(conf, "|")
	for _, it := range items {
		i := strings.TrimSpace(it)
		if len(i) == 0 {
			continue
		}
		fields := strings.Split(i, ":")
		if len(fields) != 3 {
			continue
		}

		symbol := fields[0]
		startPoint := fields[1]
		rateRange := fields[2]
		result[symbol] = &RandGeneratorConfig{
			ReferenceDataPoint: decimal.RequireFromString(startPoint),
			DistributionRate:   decimal.RequireFromString(rateRange),
		}
	}
	return result
}
