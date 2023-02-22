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
	Vote(opts *bind.TransactOpts, _commit *big.Int, _prevotes []*big.Int) (*types.Transaction, error)
	GetCommittee(opts *bind.CallOpts) ([]common.Address, error)
	GetRound(opts *bind.CallOpts) (*big.Int, error)
	WatchUpdatedCommittee(opts *bind.WatchOpts, sink chan<- *OracleUpdatedCommittee) (event.Subscription, error)
	WatchUpdatedRound(opts *bind.WatchOpts, sink chan<- *OracleUpdatedRound) (event.Subscription, error)
	WatchUpdatedSymbols(opts *bind.WatchOpts, sink chan<- *OracleUpdatedSymbols) (event.Subscription, error)
}
