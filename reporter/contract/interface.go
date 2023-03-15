package oracle

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
)

type ContractAPI interface {
	GetSymbols(opts *bind.CallOpts) ([]string, error)
	Vote(opts *bind.TransactOpts, _commit *big.Int, _prevotes []*big.Int, _salt *big.Int) (*types.Transaction, error)
	GetVoters(opts *bind.CallOpts) ([]common.Address, error)
	GetRound(opts *bind.CallOpts) (*big.Int, error)
	WatchNewRound(opts *bind.WatchOpts, sink chan<- *OracleNewRound) (event.Subscription, error)
	WatchNewSymbols(opts *bind.WatchOpts, sink chan<- *OracleNewSymbols) (event.Subscription, error)
	GetRoundData(opts *bind.CallOpts, _round *big.Int, _symbol string) (struct {
		Price     *big.Int
		Timestamp *big.Int
		Error     *big.Int
	}, error)
	LatestRoundData(opts *bind.CallOpts, _symbol string) (struct {
		Round     *big.Int
		Price     *big.Int
		Timestamp *big.Int
		Error     *big.Int
	}, error)
}
