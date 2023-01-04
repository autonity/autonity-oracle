package autonity_oracle

import (
	"autonity-oracle/config"
	"autonity-oracle/http_server"
	"autonity-oracle/oracle_server"
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() { //nolint
	// create config from system environment variables.
	conf := config.MakeConfig()
	log.Printf("Start to run autonity oracle service at port: %d\n, with symbols: %s\n",
		conf.HTTPPort, strings.Join(conf.Symbols, ","))

	// create oracle service and start the ticker job.
	oracle := oracle_server.NewOracleServer(conf.Symbols)
	go oracle.Start()
	defer oracle.Stop()

	// create http service.
	srv := http_server.NewHttpServer(oracle, conf.HTTPPort)
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
