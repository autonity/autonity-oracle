package oracle

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"math/big"
)

type ContractAPI interface {
	SetSymbols(opts *bind.TransactOpts, _symbols []string) (*types.Transaction, error)
	GetSymbols(opts *bind.CallOpts) ([]string, error)
	Vote(opts *bind.TransactOpts, _commit *big.Int, _reports []IOracleReport, _salt *big.Int, _extra uint8) (*types.Transaction, error)
	GetVotePeriod(opts *bind.CallOpts) (*big.Int, error)
	GetVoters(opts *bind.CallOpts) ([]common.Address, error)
	GetRound(opts *bind.CallOpts) (*big.Int, error)
	WatchNewRound(opts *bind.WatchOpts, sink chan<- *OracleNewRound) (event.Subscription, error)
	WatchNewSymbols(opts *bind.WatchOpts, sink chan<- *OracleNewSymbols) (event.Subscription, error)
	WatchPenalized(opts *bind.WatchOpts, sink chan<- *OraclePenalized, _participant []common.Address) (event.Subscription, error)
	GetRoundData(opts *bind.CallOpts, _round *big.Int, _symbol string) (IOracleRoundData, error)
	LatestRoundData(opts *bind.CallOpts, _symbol string) (IOracleRoundData, error)
	GetDecimals(opts *bind.CallOpts) (uint8, error)
	WatchVoted(opts *bind.WatchOpts, sink chan<- *OracleVoted, _voter []common.Address) (event.Subscription, error)
}
