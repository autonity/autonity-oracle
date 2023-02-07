package types

// Price is the basic data structure returned by Binance.
type Price struct {
	Symbol string `json:"symbol,omitempty"`
	Price  string `json:"price,omitempty"`
}

type Prices []Price

type BadRequest struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

type GeneratorParameter struct {
	Symbol string  `json:"symbol,omitempty"`
	Value  float64 `json:"value,omitempty"`
}

type GeneratorParams []GeneratorParameter
