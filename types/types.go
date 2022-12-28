package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

var (
	ErrWrongParameters = errors.New("wrong parameters")
)

type Aggregator interface {
	Aggregate(prices []decimal.Decimal) decimal.Decimal
}

type PricePool interface {
	AddPrices(prices []Price)
}

type Adapter interface {
	Name() string
	Version() string
	FetchPrices(symbols []string) error
	Alive() bool
}

type Price struct {
	Timestamp int64
	Symbol    string
	Price     decimal.Decimal
}

type PriceBySymbol map[string]Price

type OracleServiceConfig struct {
	Providers []string
	Symbols   []string
	HttpPort  int
}

type JsonRpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *JsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type JsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (err *JsonError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

func (err *JsonError) ErrorCode() int {
	return err.Code
}

func (err *JsonError) ErrorData() interface{} {
	return err.Data
}
