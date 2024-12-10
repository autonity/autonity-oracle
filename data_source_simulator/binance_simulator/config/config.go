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
	DefSimulatorConf = "ATN-USDC:1.0:0.0003|NTN-USDC:10.0:0.002|NTN-ATN:10.0:0.001|ATN-USDX:1.0:0.0003|NTN-USDX:10.0:0.002"
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
