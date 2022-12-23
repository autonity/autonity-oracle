package autonity_oralce

import (
	"autonity-oralce/types"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// todo: register a signal handler to exit the server for a clean shutdown.

	var config types.OracleServiceConfig
	// set default configs
	config.HttpPort = 8080
	config.Symbols = append(config.Symbols, "BNBBTC", "BTCUSDT")

	// create oracle service.
	oracle := NewOracleService(&config)
	go oracle.Start()

	// create http endpoint for data service.
	router := gin.Default()

	router.GET("/version", func(c *gin.Context) {
		version := oracle.Version()
		c.JSON(http.StatusOK, version)
	})

	router.GET("/prices", func(c *gin.Context) {
		prices := oracle.GetPrices()
		c.JSON(http.StatusOK, prices)
	})

	router.POST("/update_symbol", func(c *gin.Context) {
		//todo: get symbols from body.
		var symbols []string
		oracle.UpdateSymbols(symbols)
		c.JSON(http.StatusOK, symbols)
	})

	router.Run(fmt.Sprintf(":%d", config.HttpPort))
}
