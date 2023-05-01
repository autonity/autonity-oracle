package config

import (
	"github.com/namsral/flag"
	"github.com/shopspring/decimal"
	"strings"
)

var (
	// DefSimulatorConf contains a list of data items with the pattern of SYMBOL:StartingDataPoint:DataDistributionRateRange
	// with each separated by a "|". To make things be simple, we assume that the default opening rate of NTNUSD is
	// 1.0, and the other pair's opening rate are pre-computed by fiat money exchange rate. Thus, the starting data point
	// for each symbol list here also reflect the exchange rate between the corresponding fiat money, to tune the
	// starting data point for symbols, one can set the values on demand with CLI flags or system environment variables
	// with such data pattern, or just use the magnification factor parameter to increase or decrease the starting data point.
	DefSimulatorConf                            = "NTNUSD:1.0:0.01|NTNAUD:1.408:0.01|NTNCAD:1.3333:0.01|NTNEUR:0.9767:0.01|NTNGBP:0.813:0.01|NTNJPY:128.205:0.01|NTNSEK:10.309:0.01"
	DefSimulatorPort                            = 50991 // default port bind with the http service in the simulator.
	DefDataPointMagnificationFactor             = 7.0   // the default starting points are multiplied by this factor for increasing or decreasing.
	DefDataDistributionRangeMagnificationFactor = 2.0   // the default data distribution rate range is multiplied by this factor for increasing or decreasing.
	DefPlaybook                                 = ""    // the default playbook file used to replay data points in the generator.
	DefTimeout                                  = 0     // the default timeout simulated when processing a http request.
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
	dataPointMagnificationFactor := flag.Float64("sim_data_magnification_factor", DefDataPointMagnificationFactor,
		"The magnification factor to increase or decrease symbols' starting data point")
	distributionRangeMagnificationFactor := flag.Float64("sim_data_dist_range_magnification_factor",
		DefDataDistributionRangeMagnificationFactor, "The magnification factor to increase or decrease the range of the rate for random data distribution")

	flag.Parse()

	dataPointFactor := decimal.NewFromFloat(*dataPointMagnificationFactor)
	distributionRateFactor := decimal.NewFromFloat(*distributionRangeMagnificationFactor)
	conf := ParseSimulatorConf(simulatorConf, dataPointFactor, distributionRateFactor)

	println("\n\n\n\tRunning simulator with conf: ", simulatorConf)
	println("\twith data point factor: ", dataPointFactor.String())
	println("\twith data distribution rate factor: ", distributionRateFactor.String())
	println("\tRunning simulator only with playbook if playbook is configured: ", playbook)

	return &SimulatorConfig{
		Port:            port,
		Playbook:        playbook,
		SimulatorConf:   conf,
		SimulateTimeOut: simulateTimeOut,
	}
}

func ParseSimulatorConf(conf string, dataPointFactor decimal.Decimal, distributionRateFactor decimal.Decimal) map[string]*RandGeneratorConfig {
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
			ReferenceDataPoint: decimal.RequireFromString(startPoint).Mul(dataPointFactor),
			DistributionRate:   decimal.RequireFromString(rateRange).Mul(distributionRateFactor),
		}
	}
	return result
}
