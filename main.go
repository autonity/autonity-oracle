package main

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/monitor"
	"autonity-oracle/server"
	"autonity-oracle/types"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/metrics"
	"github.com/ethereum/go-ethereum/metrics/exp"
	"github.com/ethereum/go-ethereum/metrics/influxdb"
)

func main() { //nolint
	conf := config.MakeConfig()
	log.Printf("\n\n\n \tRunning autonity oracle server %s\n\twith account: %s\n\twith plugin directory: %s\n "+
		"\twith profile data directory: %s\n "+"\tby connecting to L1 node: %s\n \ton oracle contract address: %s \n\n\n",
		config.VersionString(config.Version), conf.Key.Address.String(), conf.PluginDir, conf.ProfileDir,
		conf.AutonityWSUrl, types.OracleContractAddress)

	// start prometheus metrics exposer if it is enabled.
	if conf.MetricConfigs.EnablePrometheusExp {
		metrics.Enabled = true
		log.Println("Prometheus metrics exposer enabled")
		address := fmt.Sprintf("%s:%d", conf.MetricConfigs.HTTP, conf.MetricConfigs.Port)
		log.Printf("Enabling stand-alone metrics HTTP endpoint: %s", address)
		exp.Setup(address)
	}

	// start influxDB metrics reporter if it is enabled.
	tagsMap := config.SplitTagsFlag(conf.MetricConfigs.InfluxDBTags)
	if conf.MetricConfigs.EnableInfluxDB {
		metrics.Enabled = true
		log.Printf("InfluxDB metrics enabled")
		go influxdb.InfluxDBWithTags(metrics.DefaultRegistry,
			config.MetricsInterval,
			conf.MetricConfigs.InfluxDBEndpoint,
			conf.MetricConfigs.InfluxDBDatabase,
			conf.MetricConfigs.InfluxDBUsername,
			conf.MetricConfigs.InfluxDBPassword,
			config.MetricsNameSpace, tagsMap)

		// Start system runtime metrics collection
		go metrics.CollectProcessMetrics(config.MetricsInterval)
	} else if conf.MetricConfigs.EnableInfluxDBV2 {
		metrics.Enabled = true
		log.Printf("InfluxDBV2 metrics enabled")
		go influxdb.InfluxDBV2WithTags(metrics.DefaultRegistry,
			config.MetricsInterval,
			conf.MetricConfigs.InfluxDBEndpoint,
			conf.MetricConfigs.InfluxDBToken,
			conf.MetricConfigs.InfluxDBBucket,
			conf.MetricConfigs.InfluxDBOrganization,
			config.MetricsNameSpace, tagsMap)

		// Start system runtime metrics collection
		go metrics.CollectProcessMetrics(config.MetricsInterval)
	}

	// create metrics before the context of the usage.
	monitor.InitOracleMetrics()

	// dail to L1 network, and start oracle server.
	dialer := &types.L1Dialer{}
	client, err := dialer.Dial(conf.AutonityWSUrl)
	if err != nil {
		log.Printf("cannot connect to Autonity network via web socket: %s", err.Error())
		os.Exit(1)
	}

	oc, err := contract.NewOracle(types.OracleContractAddress, client)
	if err != nil {
		log.Printf("cannot bind to oracle contract in Autonity network via web socket: %s", err.Error())
		os.Exit(1)
	}

	oracle := server.NewServer(conf, dialer, client, oc)
	go oracle.Start()
	defer oracle.Stop()

	monitorConfig := monitor.DefaultMonitorConfig
	ms := monitor.New(&monitorConfig, conf.ProfileDir)
	ms.Start()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ms.Stop()
	log.Println("shutting down oracle server...")
}
