package main

import (
	"autonity-oracle/config"
	contract "autonity-oracle/contract_binder/contract"
	"autonity-oracle/helpers"
	"autonity-oracle/oracle_server"
	"autonity-oracle/types"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() { //nolint
	conf := config.MakeConfig()
	log.Printf("\n\n\n \tRunning autonity oracle server %s\n\twith plugin directory: %s\n "+
		"\tby connecting to L1 node: %s\n \ton oracle contract address: %s \n\n\n",
		config.Version, conf.PluginDIR, conf.AutonityWSUrl, types.OracleContractAddress)

	dialer := &types.L1Dialer{}
	client, err := dialer.Dial(conf.AutonityWSUrl)
	if err != nil {
		log.Printf("Cannot connect to Autonity network via web socket: %s", err.Error())
		helpers.PrintUsage()
		os.Exit(1)
	}

	oc, err := contract.NewOracle(types.OracleContractAddress, client)
	if err != nil {
		log.Printf("Cannot bind to oracle contract in Autonity network via web socket: %s", err.Error())
		helpers.PrintUsage()
		os.Exit(1)
	}

	oracle := oracleserver.NewOracleServer(conf, dialer, client, oc)
	go oracle.Start()
	defer oracle.Stop()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down oracle server...")
}
