package main

import (
	"autonity-oracle/data_source_simulator"
	"autonity-oracle/data_source_simulator/binance_simulator/config"
	"autonity-oracle/data_source_simulator/binance_simulator/generator_manager"
	"autonity-oracle/data_source_simulator/binance_simulator/httpsrv"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() { //nolint
	// create config from flags or from system environment variables.
	conf := config.MakeSimulatorConfig()

	log.Printf("\n\n\n \t Running Binance data simulator at port: %d", conf.Port)

	// create simulators and start the ticker job to generate data points.
	var genManager data_source_simulator.GeneratorManager
	if len(conf.Playbook) == 0 {
		genManager = generator_manager.NewRandGeneratorManager(conf.SimulatorConf)
	} else {
		genManager = generator_manager.NewPlaybookGeneratorManager(conf.Playbook)
	}

	go genManager.Start()
	defer genManager.Stop()
	// create http service.
	srv := httpsrv.NewHttpServer(genManager, conf.Port, conf.SimulateTimeOut)
	srv.StartHTTPServer()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down binance simulator server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Binance simulator server forced to shutdown: ", err)
	}

	log.Println("Binance simulator server exiting")
}
