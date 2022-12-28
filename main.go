package autonity_oralce

import (
	"autonity-oralce/types"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// todo: resolve configs from system environment variables.
	var config types.OracleServiceConfig
	// set default configs
	config.HttpPort = 8080
	config.Symbols = append(config.Symbols, "BNBBTC", "BTCUSDT")

	// create oracle service.
	oracle := NewOracleService(&config)
	go oracle.Start()
	defer oracle.Stop()

	// create http endpoint for data service.
	router := gin.Default()

	router.POST("/", func(c *gin.Context) {
		var reqMsg types.JsonRpcMessage
		if err := json.NewDecoder(c.Request.Body).Decode(&reqMsg); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		switch reqMsg.Method {
		case "get_version":
			type Version struct {
				Version string
			}

			enc, err := json.Marshal(Version{Version: oracle.Version()})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, types.JsonRpcMessage{ID: reqMsg.ID, Result: enc})
			return
		case "get_prices":
			type PriceAndSymbol struct {
				Prices  types.PriceBySymbol
				Symbols []string
			}
			enc, err := json.Marshal(PriceAndSymbol{
				Prices:  oracle.GetPrices(),
				Symbols: oracle.Symbols(),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, types.JsonRpcMessage{ID: reqMsg.ID, Result: enc})
			return
		case "update_symbols":
			dec := json.NewDecoder(bytes.NewReader(reqMsg.Params))
			var symbols []string
			if len(symbols) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid symbols"})
				return
			}
			err := dec.Decode(&symbols)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			oracle.UpdateSymbols(symbols)
			c.JSON(http.StatusOK, types.JsonRpcMessage{
				ID: reqMsg.ID, Result: reqMsg.Params,
			})
			return
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown method"})
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.HttpPort),
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
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
