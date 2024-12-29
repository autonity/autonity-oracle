package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/helpers"
	"autonity-oracle/monitor"
	"autonity-oracle/oracle_server"
	"autonity-oracle/types"
)

func main() { //nolint
	configs := config.MakeConfig()
	log.Printf("\n\n\n \tRunning autonity oracle server %s\n\twith plugin directory: %s\n "+
		"\tby connecting to L1 node: %s\n \ton oracle contract address: %s \n\n\n",
		config.VersionString(config.Version), configs.PluginDIR, configs.AutonityWSUrl, types.OracleContractAddress)

	dialer := &types.L1Dialer{}
	client, err := dialer.Dial(configs.AutonityWSUrl)
	if err != nil {
		log.Printf("cannot connect to Autonity network via web socket: %s", err.Error())
		helpers.PrintUsage()
		os.Exit(1)
	}

	oc, err := contract.NewOracle(types.OracleContractAddress, client)
	if err != nil {
		log.Printf("cannot bind to oracle contract in Autonity network via web socket: %s", err.Error())
		helpers.PrintUsage()
		os.Exit(1)
	}

	oracle := oracleserver.NewOracleServer(configs, dialer, client, oc)
	go oracle.Start()
	defer oracle.Stop()

	monitorConfig := monitor.DefaultMonitorConfig
	ms := monitor.New(&monitorConfig, configs.ProfileDir)
	ms.Start()

	// todo: start metrics collection routines if metrics are enabled.

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
