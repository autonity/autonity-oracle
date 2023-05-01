package httpsrv

import (
	"autonity-oracle/data_source_simulator"
	types2 "autonity-oracle/data_source_simulator/binance_simulator/types"
	"autonity-oracle/types"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-hclog"
	"github.com/modern-go/reflect2"
	"net/http"
	o "os"
	"time"
)

var HttpRequestCounter = 0

type BinanceSimulatorHTTPServer struct {
	logger hclog.Logger
	http.Server
	generators data_source_simulator.GeneratorManager
	port       int
	timeout    int
}

func NewHttpServer(gen data_source_simulator.GeneratorManager, port int, timeout int) *BinanceSimulatorHTTPServer {
	hs := &BinanceSimulatorHTTPServer{
		generators: gen,
		port:       port,
		timeout:    timeout,
	}
	router := hs.createRouter()
	hs.logger = hclog.New(&hclog.LoggerOptions{
		Name:   reflect2.TypeOfPtr(hs).String(),
		Output: o.Stdout,
		Level:  hclog.Debug,
	})
	hs.Addr = fmt.Sprintf(":%d", port)
	hs.Handler = router
	return hs
}

// StartHTTPServer start the http server in a new go routine.
func (bs *BinanceSimulatorHTTPServer) StartHTTPServer() {
	go func() {
		if err := bs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			bs.logger.Error("Binance simulator HTTP service listen on port: ", bs.port, err)
			panic(err.Error())
		}
	}()
}

func (bs *BinanceSimulatorHTTPServer) createRouter() *gin.Engine {
	// create http api handlers.
	gin.SetMode("release")
	router := gin.Default()

	// data reader handler
	router.GET("/api/v3/ticker/price", func(c *gin.Context) {

		// if the simulator is configured to simulate timeout.
		if bs.timeout != 0 {
			HttpRequestCounter++
			if HttpRequestCounter%5 == 0 {
				time.Sleep(time.Second * time.Duration(bs.timeout))
			}
		}

		s := c.Query("symbols")
		var symbols []string

		if s != "" {
			if err := json.Unmarshal([]byte(s), &symbols); err != nil {
				c.JSON(http.StatusBadRequest, types2.BadRequest{
					Code: 400,
					Msg:  "Invalid parameters",
				})
			}
		}

		prices, err := bs.generators.GetSymbolPrice(symbols)
		if err != nil {
			c.JSON(http.StatusBadRequest, types2.BadRequest{
				Code: 400,
				Msg:  err.Error(),
			})
		}

		c.JSON(http.StatusOK, prices)
	})

	router.POST("/", func(c *gin.Context) {
		var reqMsg types.JSONRPCMessage
		if err := json.NewDecoder(c.Request.Body).Decode(&reqMsg); err != nil {
			c.JSON(http.StatusBadRequest, types.JSONRPCMessage{Error: err.Error()})
		}
		bs.logger.Debug("handling method:", reqMsg.Method)
		c.JSON(bs.adjustSimulatorParam(&reqMsg))
	})

	return router
}

func (bs *BinanceSimulatorHTTPServer) adjustSimulatorParam(reqMsg *types.JSONRPCMessage) (int, types.JSONRPCMessage) {
	dec := json.NewDecoder(bytes.NewReader(reqMsg.Params))
	var params types2.GeneratorParams
	err := dec.Decode(&params)
	if err != nil {
		return http.StatusBadRequest, types.JSONRPCMessage{Error: err.Error()}
	}
	err = bs.generators.AdjustParams(params, reqMsg.Method)
	if err != nil {
		return http.StatusBadRequest, types.JSONRPCMessage{Error: err.Error()}
	}

	return http.StatusOK, types.JSONRPCMessage{ID: reqMsg.ID, Result: reqMsg.Params}
}
