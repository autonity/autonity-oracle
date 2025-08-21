package oracle

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

type ContractAPI interface {
	SetSymbols(opts *bind.TransactOpts, _symbols []string) (*types.Transaction, error)
	GetSymbols(opts *bind.CallOpts) ([]string, error)
	Vote(opts *bind.TransactOpts, _commit *big.Int, _reports []IOracleReport, _salt *big.Int, _extra uint8) (*types.Transaction, error)
	GetVotePeriod(opts *bind.CallOpts) (*big.Int, error)
	GetVoters(opts *bind.CallOpts) ([]common.Address, error)
	GetRound(opts *bind.CallOpts) (*big.Int, error)
	GetLastRoundBlock(opts *bind.CallOpts) (*big.Int, error)
	WatchNewRound(opts *bind.WatchOpts, sink chan<- *OracleNewRound) (event.Subscription, error)
	WatchNewSymbols(opts *bind.WatchOpts, sink chan<- *OracleNewSymbols) (event.Subscription, error)
	WatchPenalized(opts *bind.WatchOpts, sink chan<- *OraclePenalized, _participant []common.Address) (event.Subscription, error)
	WatchNoRevealPenalty(opts *bind.WatchOpts, sink chan<- *OracleNoRevealPenalty, _voter []common.Address) (event.Subscription, error)
	WatchSuccessfulVote(opts *bind.WatchOpts, sink chan<- *OracleSuccessfulVote, reporter []common.Address) (event.Subscription, error)
	WatchInvalidVote(opts *bind.WatchOpts, sink chan<- *OracleInvalidVote, reporter []common.Address) (event.Subscription, error)
	WatchTotalOracleRewards(opts *bind.WatchOpts, sink chan<- *OracleTotalOracleRewards) (event.Subscription, error)
	GetRoundData(opts *bind.CallOpts, _round *big.Int, _symbol string) (IOracleRoundData, error)
	LatestRoundData(opts *bind.CallOpts, _symbol string) (IOracleRoundData, error)
	GetDecimals(opts *bind.CallOpts) (uint8, error)
}
