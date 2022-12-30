package http_server

import (
	"autonity-oralce/oracle_server"
	"autonity-oralce/types"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type HttpServer struct {
	http.Server
	oracle *oracle_server.OracleServer
	port   int
}

func NewHttpServer(os *oracle_server.OracleServer, port int) *HttpServer {
	hs := &HttpServer{
		oracle: os,
		port:   port,
	}
	router := hs.createRouter()

	hs.Addr = fmt.Sprintf(":%d", port)
	hs.Handler = router
	return hs
}

// StartHttpServer start the http server in a new go routine.
func (hs *HttpServer) StartHttpServer() {
	go func() {
		if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

func (hs *HttpServer) createRouter() *gin.Engine {
	// create http api handlers.
	router := gin.Default()
	router.POST("/", func(c *gin.Context) {
		var reqMsg types.JsonRpcMessage
		if err := json.NewDecoder(c.Request.Body).Decode(&reqMsg); err != nil {
			c.JSON(http.StatusBadRequest, types.JsonRpcMessage{Error: err.Error()})
		}

		switch reqMsg.Method {
		case "get_version":
			rsp, code := hs.getVersion(&reqMsg)
			c.JSON(code, rsp)
		case "get_prices":
			rsp, code := hs.getPrices(&reqMsg)
			c.JSON(code, rsp)
		case "update_symbols":
			rsp, code := hs.updateSymbols(&reqMsg)
			c.JSON(code, rsp)
		default:
			c.JSON(http.StatusBadRequest, types.JsonRpcMessage{Error: "unknown method"})
		}
	})
	return router
}

// handler functions
func (hs *HttpServer) getVersion(reqMsg *types.JsonRpcMessage) (types.JsonRpcMessage, int) {
	type Version struct {
		Version string
	}

	enc, err := json.Marshal(Version{Version: hs.oracle.Version()})
	if err != nil {
		return types.JsonRpcMessage{Error: err.Error()}, http.StatusInternalServerError
	}

	return types.JsonRpcMessage{ID: reqMsg.ID, Result: enc}, http.StatusOK
}

func (hs *HttpServer) getPrices(reqMsg *types.JsonRpcMessage) (types.JsonRpcMessage, int) {
	type PriceAndSymbol struct {
		Prices  types.PriceBySymbol
		Symbols []string
	}
	enc, err := json.Marshal(PriceAndSymbol{
		Prices:  hs.oracle.GetPrices(),
		Symbols: hs.oracle.Symbols(),
	})
	if err != nil {
		return types.JsonRpcMessage{Error: err.Error()}, http.StatusInternalServerError
	}
	return types.JsonRpcMessage{ID: reqMsg.ID, Result: enc}, http.StatusOK
}

func (hs *HttpServer) updateSymbols(reqMsg *types.JsonRpcMessage) (types.JsonRpcMessage, int) {
	dec := json.NewDecoder(bytes.NewReader(reqMsg.Params))
	var symbols []string
	err := dec.Decode(&symbols)
	if err != nil {
		return types.JsonRpcMessage{Error: err.Error()}, http.StatusBadRequest
	}
	if len(symbols) == 0 {
		return types.JsonRpcMessage{Error: "setting with empty symbols"}, http.StatusBadRequest
	}
	hs.oracle.UpdateSymbols(symbols)
	return types.JsonRpcMessage{ID: reqMsg.ID, Result: reqMsg.Params}, http.StatusOK
}
