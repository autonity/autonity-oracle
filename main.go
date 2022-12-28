package autonity_oralce

import (
	"autonity-oralce/types"
	"bytes"
	"encoding/json"
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

	router.Run(fmt.Sprintf(":%d", config.HttpPort))
}
