package main

import (
	"autonity-oracle/config"
	"autonity-oracle/oracle_server"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() { //nolint
	conf := config.MakeConfig()
	log.Printf("\n\n\n \tRunning autonity oracle node with symbols: %s and plugin diretory: %s by connnecting to L1 node: %s \n\n\n",
		strings.Join(conf.Symbols, ","), conf.PluginDIR, conf.AutonityWSUrl)

	oracle := oracleserver.NewOracleServer(conf.Symbols, conf.PluginDIR, conf.AutonityWSUrl, conf.Key)
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
	log.Println("Server exiting")
}
