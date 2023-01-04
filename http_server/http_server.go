package http_server

import (
	"autonity-oracle/oracle_server"
	"autonity-oracle/types"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type HTTPServer struct {
	http.Server
	oracle *oracle_server.OracleServer
	port   int
}

func NewHttpServer(os *oracle_server.OracleServer, port int) *HTTPServer {
	hs := &HTTPServer{
		oracle: os,
		port:   port,
	}
	router := hs.createRouter()

	hs.Addr = fmt.Sprintf(":%d", port)
	hs.Handler = router
	return hs
}

// StartHTTPServer start the http server in a new go routine.
func (hs *HTTPServer) StartHTTPServer() {
	go func() {
		if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

func (hs *HTTPServer) createRouter() *gin.Engine {
	// create http api handlers.
	router := gin.Default()
	router.POST("/", func(c *gin.Context) {
		var reqMsg types.JSONRPCMessage
		if err := json.NewDecoder(c.Request.Body).Decode(&reqMsg); err != nil {
			c.JSON(http.StatusBadRequest, types.JSONRPCMessage{Error: err.Error()})
		}

		switch reqMsg.Method {
		case "get_version":
			c.JSON(hs.getVersion(&reqMsg))
		case "get_prices":
			c.JSON(hs.getPrices(&reqMsg))
		case "update_symbols":
			c.JSON(hs.updateSymbols(&reqMsg))
		default:
			c.JSON(http.StatusBadRequest, types.JSONRPCMessage{ID: reqMsg.ID, Error: "unknown method"})
		}
	})
	return router
}

// handler functions
func (hs *HTTPServer) getVersion(reqMsg *types.JSONRPCMessage) (int, types.JSONRPCMessage) {
	type Version struct {
		Version string
	}

	enc, err := json.Marshal(Version{Version: hs.oracle.Version()})
	if err != nil {
		return http.StatusInternalServerError, types.JSONRPCMessage{Error: err.Error()}
	}

	return http.StatusOK, types.JSONRPCMessage{ID: reqMsg.ID, Result: enc}
}

func (hs *HTTPServer) getPrices(reqMsg *types.JSONRPCMessage) (int, types.JSONRPCMessage) {
	type PriceAndSymbol struct {
		Prices  types.PriceBySymbol
		Symbols []string
	}
	enc, err := json.Marshal(PriceAndSymbol{
		Prices:  hs.oracle.GetPrices(),
		Symbols: hs.oracle.Symbols(),
	})
	if err != nil {
		return http.StatusInternalServerError, types.JSONRPCMessage{Error: err.Error()}
	}
	return http.StatusOK, types.JSONRPCMessage{ID: reqMsg.ID, Result: enc}
}

func (hs *HTTPServer) updateSymbols(reqMsg *types.JSONRPCMessage) (int, types.JSONRPCMessage) {
	dec := json.NewDecoder(bytes.NewReader(reqMsg.Params))
	var symbols []string
	err := dec.Decode(&symbols)
	if err != nil {
		return http.StatusBadRequest, types.JSONRPCMessage{Error: err.Error()}
	}
	if len(symbols) == 0 {
		return http.StatusBadRequest, types.JSONRPCMessage{Error: "setting with empty symbols"}
	}
	hs.oracle.UpdateSymbols(symbols)
	return http.StatusOK, types.JSONRPCMessage{ID: reqMsg.ID, Result: reqMsg.Params}
}
