package main

import (
	"autonity-oracle/config"
	"autonity-oracle/http_server"
	"autonity-oracle/oracle_server"
	rp "autonity-oracle/reporter"
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() { //nolint
	// create config from system environment variables or from console flags.
	conf := config.MakeConfig()
	log.Printf("\n\n\n \tRunning autonity oracle service at port: %d, with symbols: %s and plugin diretory: %s\n\n\n",
		conf.HTTPPort, strings.Join(conf.Symbols, ","), conf.PluginDIR)

	// create oracle service and start the ticker job.
	oracle := oracleserver.NewOracleServer(conf.Symbols, conf.PluginDIR)
	go oracle.Start()
	defer oracle.Stop()

	reporter := rp.NewDataReporter(conf.AutonityWSUrl, conf.Key, oracle)
	go reporter.Start()
	defer reporter.Stop()

	// create http service.
	srv := httpserver.NewHttpServer(oracle, conf.HTTPPort)
	srv.StartHTTPServer()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown: ", err)
	}

	log.Println("Server exiting")
}
