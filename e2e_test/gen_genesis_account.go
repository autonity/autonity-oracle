package test

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
)

var _ = (*genesisAccountMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (g GenesisAccount) MarshalJSON() ([]byte, error) {
	type GenesisAccount struct {
		Code          hexutil.Bytes                            `json:"code,omitempty"`
		Storage       map[storageJSON]storageJSON              `json:"storage,omitempty"`
		Balance       *math.HexOrDecimal256                    `json:"balance" gencodec:"required"`
		NewtonBalance *math.HexOrDecimal256                    `json:"newtonBalance"`
		Bonds         map[common.Address]*math.HexOrDecimal256 `json:"bonds"`
		Nonce         math.HexOrDecimal64                      `json:"nonce,omitempty"`
		PrivateKey    hexutil.Bytes                            `json:"secretKey,omitempty"`
	}
	var enc GenesisAccount
	enc.Code = g.Code
	if g.Storage != nil {
		enc.Storage = make(map[storageJSON]storageJSON, len(g.Storage))
		for k, v := range g.Storage {
			enc.Storage[storageJSON(k)] = storageJSON(v)
		}
	}
	enc.Balance = (*math.HexOrDecimal256)(g.Balance)
	enc.NewtonBalance = (*math.HexOrDecimal256)(g.NewtonBalance)
	if g.Bonds != nil {
		enc.Bonds = make(map[common.Address]*math.HexOrDecimal256, len(g.Bonds))
		for k, v := range g.Bonds {
			enc.Bonds[k] = (*math.HexOrDecimal256)(v)
		}
	}
	enc.Nonce = math.HexOrDecimal64(g.Nonce)
	enc.PrivateKey = g.PrivateKey
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (g *GenesisAccount) UnmarshalJSON(input []byte) error {
	type GenesisAccount struct {
		Code          *hexutil.Bytes                           `json:"code,omitempty"`
		Storage       map[storageJSON]storageJSON              `json:"storage,omitempty"`
		Balance       *math.HexOrDecimal256                    `json:"balance" gencodec:"required"`
		NewtonBalance *math.HexOrDecimal256                    `json:"newtonBalance"`
		Bonds         map[common.Address]*math.HexOrDecimal256 `json:"bonds"`
		Nonce         *math.HexOrDecimal64                     `json:"nonce,omitempty"`
		PrivateKey    *hexutil.Bytes                           `json:"secretKey,omitempty"`
	}
	var dec GenesisAccount
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Code != nil {
		g.Code = *dec.Code
	}
	if dec.Storage != nil {
		g.Storage = make(map[common.Hash]common.Hash, len(dec.Storage))
		for k, v := range dec.Storage {
			g.Storage[common.Hash(k)] = common.Hash(v)
		}
	}
	if dec.Balance == nil {
		return errors.New("missing required field 'balance' for GenesisAccount")
	}
	g.Balance = (*big.Int)(dec.Balance)
	if dec.NewtonBalance != nil {
		g.NewtonBalance = (*big.Int)(dec.NewtonBalance)
	}
	if dec.Bonds != nil {
		g.Bonds = make(map[common.Address]*big.Int, len(dec.Bonds))
		for k, v := range dec.Bonds {
			g.Bonds[k] = (*big.Int)(v)
		}
	}
	if dec.Nonce != nil {
		g.Nonce = uint64(*dec.Nonce)
	}
	if dec.PrivateKey != nil {
		g.PrivateKey = *dec.PrivateKey
	}
	return nil
}
