package autonity_oralce

import (
	"autonity-oralce/types"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// todo: register a signal handler to exit the server for a clean shutdown.

	//todo: resolve configurations.
	var config types.OracleServiceConfig

	// create oracle service.
	oracle := NewOracleService(&config)
	go oracle.Start()

	// create http endpoint for data service.
	router := gin.Default()

	router.GET("/version", func(c *gin.Context) {
		version := oracle.Version()
		c.SecureJSON(http.StatusOK, version)
	})

	router.GET("/prices", func(c *gin.Context) {
		prices := oracle.GetPrices()
		c.SecureJSON(http.StatusOK, prices)
	})

	// todo: resolve port from config.
	router.Run(":8080")
}
