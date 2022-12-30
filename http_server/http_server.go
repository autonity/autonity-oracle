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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		switch reqMsg.Method {
		case "get_version":
			rsp, err := hs.getVersion(&reqMsg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, rsp)
		case "get_prices":
			rsp, err := hs.getPrices(&reqMsg)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, rsp)
		case "update_symbols":
			rsp, err := hs.updateSymbols(&reqMsg)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, rsp)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown method"})
		}
	})
	return router
}

// handler functions
func (hs *HttpServer) getVersion(reqMsg *types.JsonRpcMessage) (types.JsonRpcMessage, error) {
	type Version struct {
		Version string
	}

	enc, err := json.Marshal(Version{Version: hs.oracle.Version()})
	if err != nil {
		return types.JsonRpcMessage{}, err
	}

	return types.JsonRpcMessage{ID: reqMsg.ID, Result: enc}, nil
}

func (hs *HttpServer) getPrices(reqMsg *types.JsonRpcMessage) (types.JsonRpcMessage, error) {
	type PriceAndSymbol struct {
		Prices  types.PriceBySymbol
		Symbols []string
	}
	enc, err := json.Marshal(PriceAndSymbol{
		Prices:  hs.oracle.GetPrices(),
		Symbols: hs.oracle.Symbols(),
	})
	if err != nil {
		return types.JsonRpcMessage{}, err
	}
	return types.JsonRpcMessage{ID: reqMsg.ID, Result: enc}, nil
}

func (hs *HttpServer) updateSymbols(reqMsg *types.JsonRpcMessage) (types.JsonRpcMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(reqMsg.Params))
	var symbols []string
	err := dec.Decode(&symbols)
	if err != nil {
		return types.JsonRpcMessage{}, err
	}
	if len(symbols) == 0 {
		return types.JsonRpcMessage{}, err
	}
	hs.oracle.UpdateSymbols(symbols)
	return types.JsonRpcMessage{ID: reqMsg.ID, Result: reqMsg.Params}, nil
}
