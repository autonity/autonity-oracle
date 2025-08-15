package oracle

import (
	"encoding/json"
	"math/big"
)

// Custom JSON marshaler/unmarshaler for contract.IOracleReport
func (r IOracleReport) MarshalJSON() ([]byte, error) {
	type Alias struct {
		Price      string `json:"price"`
		Confidence uint8  `json:"confidence"`
	}
	return json.Marshal(Alias{
		Price:      r.Price.String(),
		Confidence: r.Confidence,
	})
}

func (r *IOracleReport) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Price      string `json:"price"`
		Confidence uint8  `json:"confidence"`
	}
	var tmp Alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	r.Price = new(big.Int)
	r.Price.SetString(tmp.Price, 10)
	r.Confidence = tmp.Confidence
	return nil
}
