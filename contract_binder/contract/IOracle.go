// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package oracle

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// IOracleReport is an auto generated low-level Go binding around an user-defined struct.
type IOracleReport struct {
	Price      *big.Int
	Confidence uint8
}

// IOracleRoundData is an auto generated low-level Go binding around an user-defined struct.
type IOracleRoundData struct {
	Round     *big.Int
	Price     *big.Int
	Timestamp *big.Int
	Success   bool
}

// OracleMetaData contains all meta data concerning the Oracle contract.
var OracleMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_voter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_round\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_nonRevealCount\",\"type\":\"uint256\"}],\"name\":\"CommitRevealMissed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"cause\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"reporter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"expValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"actualValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"extra\",\"type\":\"uint8\"}],\"name\":\"InvalidVote\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_round\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_timestamp\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_votePeriod\",\"type\":\"uint256\"}],\"name\":\"NewRound\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"_symbols\",\"type\":\"string[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_round\",\"type\":\"uint256\"}],\"name\":\"NewSymbols\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"reporter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"extra\",\"type\":\"uint8\"}],\"name\":\"NewVoter\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_voter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_round\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_missedReveal\",\"type\":\"uint256\"}],\"name\":\"NoRevealPenalty\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_participant\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_slashingAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"int256\",\"name\":\"_median\",\"type\":\"int256\"},{\"indexed\":false,\"internalType\":\"uint120\",\"name\":\"_reported\",\"type\":\"uint120\"}],\"name\":\"Penalized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"round\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"status\",\"type\":\"bool\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"PriceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"reporter\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"extra\",\"type\":\"uint8\"}],\"name\":\"SuccessfulVote\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"ntnReward\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"atnReward\",\"type\":\"uint256\"}],\"name\":\"TotalOracleRewards\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_ntnRewards\",\"type\":\"uint256\"}],\"name\":\"distributeRewards\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"finalize\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDecimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNewVotePeriod\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNewVoters\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getNonRevealThreshold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRound\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_round\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"getRoundData\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"round\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"internalType\":\"structIOracle.RoundData\",\"name\":\"data\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSymbols\",\"outputs\":[{\"internalType\":\"string[]\",\"name\":\"_symbols\",\"type\":\"string[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getVotePeriod\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getVoters\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"latestRoundData\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"round\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"success\",\"type\":\"bool\"}],\"internalType\":\"structIOracle.RoundData\",\"name\":\"data\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_resetInterval\",\"type\":\"uint256\"}],\"name\":\"setCommitRevealConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_operator\",\"type\":\"address\"}],\"name\":\"setOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"int256\",\"name\":\"_outlierSlashingThreshold\",\"type\":\"int256\"},{\"internalType\":\"int256\",\"name\":\"_outlierDetectionThreshold\",\"type\":\"int256\"},{\"internalType\":\"uint256\",\"name\":\"_baseSlashingRate\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_slashingRateCap\",\"type\":\"uint256\"}],\"name\":\"setSlashingConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string[]\",\"name\":\"_symbols\",\"type\":\"string[]\"}],\"name\":\"setSymbols\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_newVoters\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"_treasury\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"_validator\",\"type\":\"address[]\"}],\"name\":\"setVoters\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"updateVotersAndSymbol\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_commit\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint120\",\"name\":\"price\",\"type\":\"uint120\"},{\"internalType\":\"uint8\",\"name\":\"confidence\",\"type\":\"uint8\"}],\"internalType\":\"structIOracle.Report[]\",\"name\":\"_reports\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"_salt\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_extra\",\"type\":\"uint8\"}],\"name\":\"vote\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// OracleABI is the input ABI used to generate the binding from.
// Deprecated: Use OracleMetaData.ABI instead.
var OracleABI = OracleMetaData.ABI

// Oracle is an auto generated Go binding around an Ethereum contract.
type Oracle struct {
	OracleCaller     // Read-only binding to the contract
	OracleTransactor // Write-only binding to the contract
	OracleFilterer   // Log filterer for contract events
}

// OracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type OracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OracleSession struct {
	Contract     *Oracle           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OracleCallerSession struct {
	Contract *OracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// OracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OracleTransactorSession struct {
	Contract     *OracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// OracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type OracleRaw struct {
	Contract *Oracle // Generic contract binding to access the raw methods on
}

// OracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OracleCallerRaw struct {
	Contract *OracleCaller // Generic read-only contract binding to access the raw methods on
}

// OracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OracleTransactorRaw struct {
	Contract *OracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOracle creates a new instance of Oracle, bound to a specific deployed contract.
func NewOracle(address common.Address, backend bind.ContractBackend) (*Oracle, error) {
	contract, err := bindOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Oracle{OracleCaller: OracleCaller{contract: contract}, OracleTransactor: OracleTransactor{contract: contract}, OracleFilterer: OracleFilterer{contract: contract}}, nil
}

// NewOracleCaller creates a new read-only instance of Oracle, bound to a specific deployed contract.
func NewOracleCaller(address common.Address, caller bind.ContractCaller) (*OracleCaller, error) {
	contract, err := bindOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OracleCaller{contract: contract}, nil
}

// NewOracleTransactor creates a new write-only instance of Oracle, bound to a specific deployed contract.
func NewOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*OracleTransactor, error) {
	contract, err := bindOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OracleTransactor{contract: contract}, nil
}

// NewOracleFilterer creates a new log filterer instance of Oracle, bound to a specific deployed contract.
func NewOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*OracleFilterer, error) {
	contract, err := bindOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OracleFilterer{contract: contract}, nil
}

// bindOracle binds a generic wrapper to an already deployed contract.
func bindOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(OracleABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Oracle *OracleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Oracle.Contract.OracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Oracle *OracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Oracle.Contract.OracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Oracle *OracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Oracle.Contract.OracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Oracle *OracleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Oracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Oracle *OracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Oracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Oracle *OracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Oracle.Contract.contract.Transact(opts, method, params...)
}

// GetDecimals is a free data retrieval call binding the contract method 0xf0141d84.
//
// Solidity: function getDecimals() view returns(uint8)
func (_Oracle *OracleCaller) GetDecimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getDecimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// GetDecimals is a free data retrieval call binding the contract method 0xf0141d84.
//
// Solidity: function getDecimals() view returns(uint8)
func (_Oracle *OracleSession) GetDecimals() (uint8, error) {
	return _Oracle.Contract.GetDecimals(&_Oracle.CallOpts)
}

// GetDecimals is a free data retrieval call binding the contract method 0xf0141d84.
//
// Solidity: function getDecimals() view returns(uint8)
func (_Oracle *OracleCallerSession) GetDecimals() (uint8, error) {
	return _Oracle.Contract.GetDecimals(&_Oracle.CallOpts)
}

// GetNewVotePeriod is a free data retrieval call binding the contract method 0x57eba759.
//
// Solidity: function getNewVotePeriod() view returns(uint256)
func (_Oracle *OracleCaller) GetNewVotePeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getNewVotePeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNewVotePeriod is a free data retrieval call binding the contract method 0x57eba759.
//
// Solidity: function getNewVotePeriod() view returns(uint256)
func (_Oracle *OracleSession) GetNewVotePeriod() (*big.Int, error) {
	return _Oracle.Contract.GetNewVotePeriod(&_Oracle.CallOpts)
}

// GetNewVotePeriod is a free data retrieval call binding the contract method 0x57eba759.
//
// Solidity: function getNewVotePeriod() view returns(uint256)
func (_Oracle *OracleCallerSession) GetNewVotePeriod() (*big.Int, error) {
	return _Oracle.Contract.GetNewVotePeriod(&_Oracle.CallOpts)
}

// GetNewVoters is a free data retrieval call binding the contract method 0x077945d3.
//
// Solidity: function getNewVoters() view returns(address[])
func (_Oracle *OracleCaller) GetNewVoters(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getNewVoters")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetNewVoters is a free data retrieval call binding the contract method 0x077945d3.
//
// Solidity: function getNewVoters() view returns(address[])
func (_Oracle *OracleSession) GetNewVoters() ([]common.Address, error) {
	return _Oracle.Contract.GetNewVoters(&_Oracle.CallOpts)
}

// GetNewVoters is a free data retrieval call binding the contract method 0x077945d3.
//
// Solidity: function getNewVoters() view returns(address[])
func (_Oracle *OracleCallerSession) GetNewVoters() ([]common.Address, error) {
	return _Oracle.Contract.GetNewVoters(&_Oracle.CallOpts)
}

// GetNonRevealThreshold is a free data retrieval call binding the contract method 0xed78349d.
//
// Solidity: function getNonRevealThreshold() view returns(uint256)
func (_Oracle *OracleCaller) GetNonRevealThreshold(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getNonRevealThreshold")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNonRevealThreshold is a free data retrieval call binding the contract method 0xed78349d.
//
// Solidity: function getNonRevealThreshold() view returns(uint256)
func (_Oracle *OracleSession) GetNonRevealThreshold() (*big.Int, error) {
	return _Oracle.Contract.GetNonRevealThreshold(&_Oracle.CallOpts)
}

// GetNonRevealThreshold is a free data retrieval call binding the contract method 0xed78349d.
//
// Solidity: function getNonRevealThreshold() view returns(uint256)
func (_Oracle *OracleCallerSession) GetNonRevealThreshold() (*big.Int, error) {
	return _Oracle.Contract.GetNonRevealThreshold(&_Oracle.CallOpts)
}

// GetRound is a free data retrieval call binding the contract method 0x9f8743f7.
//
// Solidity: function getRound() view returns(uint256)
func (_Oracle *OracleCaller) GetRound(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getRound")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRound is a free data retrieval call binding the contract method 0x9f8743f7.
//
// Solidity: function getRound() view returns(uint256)
func (_Oracle *OracleSession) GetRound() (*big.Int, error) {
	return _Oracle.Contract.GetRound(&_Oracle.CallOpts)
}

// GetRound is a free data retrieval call binding the contract method 0x9f8743f7.
//
// Solidity: function getRound() view returns(uint256)
func (_Oracle *OracleCallerSession) GetRound() (*big.Int, error) {
	return _Oracle.Contract.GetRound(&_Oracle.CallOpts)
}

// GetRoundData is a free data retrieval call binding the contract method 0x3c8510fd.
//
// Solidity: function getRoundData(uint256 _round, string _symbol) view returns((uint256,uint256,uint256,bool) data)
func (_Oracle *OracleCaller) GetRoundData(opts *bind.CallOpts, _round *big.Int, _symbol string) (IOracleRoundData, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getRoundData", _round, _symbol)

	if err != nil {
		return *new(IOracleRoundData), err
	}

	out0 := *abi.ConvertType(out[0], new(IOracleRoundData)).(*IOracleRoundData)

	return out0, err

}

// GetRoundData is a free data retrieval call binding the contract method 0x3c8510fd.
//
// Solidity: function getRoundData(uint256 _round, string _symbol) view returns((uint256,uint256,uint256,bool) data)
func (_Oracle *OracleSession) GetRoundData(_round *big.Int, _symbol string) (IOracleRoundData, error) {
	return _Oracle.Contract.GetRoundData(&_Oracle.CallOpts, _round, _symbol)
}

// GetRoundData is a free data retrieval call binding the contract method 0x3c8510fd.
//
// Solidity: function getRoundData(uint256 _round, string _symbol) view returns((uint256,uint256,uint256,bool) data)
func (_Oracle *OracleCallerSession) GetRoundData(_round *big.Int, _symbol string) (IOracleRoundData, error) {
	return _Oracle.Contract.GetRoundData(&_Oracle.CallOpts, _round, _symbol)
}

// GetSymbols is a free data retrieval call binding the contract method 0xdf7f710e.
//
// Solidity: function getSymbols() view returns(string[] _symbols)
func (_Oracle *OracleCaller) GetSymbols(opts *bind.CallOpts) ([]string, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getSymbols")

	if err != nil {
		return *new([]string), err
	}

	out0 := *abi.ConvertType(out[0], new([]string)).(*[]string)

	return out0, err

}

// GetSymbols is a free data retrieval call binding the contract method 0xdf7f710e.
//
// Solidity: function getSymbols() view returns(string[] _symbols)
func (_Oracle *OracleSession) GetSymbols() ([]string, error) {
	return _Oracle.Contract.GetSymbols(&_Oracle.CallOpts)
}

// GetSymbols is a free data retrieval call binding the contract method 0xdf7f710e.
//
// Solidity: function getSymbols() view returns(string[] _symbols)
func (_Oracle *OracleCallerSession) GetSymbols() ([]string, error) {
	return _Oracle.Contract.GetSymbols(&_Oracle.CallOpts)
}

// GetVotePeriod is a free data retrieval call binding the contract method 0xb78dec52.
//
// Solidity: function getVotePeriod() view returns(uint256)
func (_Oracle *OracleCaller) GetVotePeriod(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getVotePeriod")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetVotePeriod is a free data retrieval call binding the contract method 0xb78dec52.
//
// Solidity: function getVotePeriod() view returns(uint256)
func (_Oracle *OracleSession) GetVotePeriod() (*big.Int, error) {
	return _Oracle.Contract.GetVotePeriod(&_Oracle.CallOpts)
}

// GetVotePeriod is a free data retrieval call binding the contract method 0xb78dec52.
//
// Solidity: function getVotePeriod() view returns(uint256)
func (_Oracle *OracleCallerSession) GetVotePeriod() (*big.Int, error) {
	return _Oracle.Contract.GetVotePeriod(&_Oracle.CallOpts)
}

// GetVoters is a free data retrieval call binding the contract method 0xcdd72253.
//
// Solidity: function getVoters() view returns(address[])
func (_Oracle *OracleCaller) GetVoters(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getVoters")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetVoters is a free data retrieval call binding the contract method 0xcdd72253.
//
// Solidity: function getVoters() view returns(address[])
func (_Oracle *OracleSession) GetVoters() ([]common.Address, error) {
	return _Oracle.Contract.GetVoters(&_Oracle.CallOpts)
}

// GetVoters is a free data retrieval call binding the contract method 0xcdd72253.
//
// Solidity: function getVoters() view returns(address[])
func (_Oracle *OracleCallerSession) GetVoters() ([]common.Address, error) {
	return _Oracle.Contract.GetVoters(&_Oracle.CallOpts)
}

// LatestRoundData is a free data retrieval call binding the contract method 0x33f98c77.
//
// Solidity: function latestRoundData(string _symbol) view returns((uint256,uint256,uint256,bool) data)
func (_Oracle *OracleCaller) LatestRoundData(opts *bind.CallOpts, _symbol string) (IOracleRoundData, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "latestRoundData", _symbol)

	if err != nil {
		return *new(IOracleRoundData), err
	}

	out0 := *abi.ConvertType(out[0], new(IOracleRoundData)).(*IOracleRoundData)

	return out0, err

}

// LatestRoundData is a free data retrieval call binding the contract method 0x33f98c77.
//
// Solidity: function latestRoundData(string _symbol) view returns((uint256,uint256,uint256,bool) data)
func (_Oracle *OracleSession) LatestRoundData(_symbol string) (IOracleRoundData, error) {
	return _Oracle.Contract.LatestRoundData(&_Oracle.CallOpts, _symbol)
}

// LatestRoundData is a free data retrieval call binding the contract method 0x33f98c77.
//
// Solidity: function latestRoundData(string _symbol) view returns((uint256,uint256,uint256,bool) data)
func (_Oracle *OracleCallerSession) LatestRoundData(_symbol string) (IOracleRoundData, error) {
	return _Oracle.Contract.LatestRoundData(&_Oracle.CallOpts, _symbol)
}

// DistributeRewards is a paid mutator transaction binding the contract method 0x59974e38.
//
// Solidity: function distributeRewards(uint256 _ntnRewards) payable returns()
func (_Oracle *OracleTransactor) DistributeRewards(opts *bind.TransactOpts, _ntnRewards *big.Int) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "distributeRewards", _ntnRewards)
}

// DistributeRewards is a paid mutator transaction binding the contract method 0x59974e38.
//
// Solidity: function distributeRewards(uint256 _ntnRewards) payable returns()
func (_Oracle *OracleSession) DistributeRewards(_ntnRewards *big.Int) (*types.Transaction, error) {
	return _Oracle.Contract.DistributeRewards(&_Oracle.TransactOpts, _ntnRewards)
}

// DistributeRewards is a paid mutator transaction binding the contract method 0x59974e38.
//
// Solidity: function distributeRewards(uint256 _ntnRewards) payable returns()
func (_Oracle *OracleTransactorSession) DistributeRewards(_ntnRewards *big.Int) (*types.Transaction, error) {
	return _Oracle.Contract.DistributeRewards(&_Oracle.TransactOpts, _ntnRewards)
}

// Finalize is a paid mutator transaction binding the contract method 0x4bb278f3.
//
// Solidity: function finalize() returns(bool)
func (_Oracle *OracleTransactor) Finalize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "finalize")
}

// Finalize is a paid mutator transaction binding the contract method 0x4bb278f3.
//
// Solidity: function finalize() returns(bool)
func (_Oracle *OracleSession) Finalize() (*types.Transaction, error) {
	return _Oracle.Contract.Finalize(&_Oracle.TransactOpts)
}

// Finalize is a paid mutator transaction binding the contract method 0x4bb278f3.
//
// Solidity: function finalize() returns(bool)
func (_Oracle *OracleTransactorSession) Finalize() (*types.Transaction, error) {
	return _Oracle.Contract.Finalize(&_Oracle.TransactOpts)
}

// SetCommitRevealConfig is a paid mutator transaction binding the contract method 0x3f422ef3.
//
// Solidity: function setCommitRevealConfig(uint256 _threshold, uint256 _resetInterval) returns()
func (_Oracle *OracleTransactor) SetCommitRevealConfig(opts *bind.TransactOpts, _threshold *big.Int, _resetInterval *big.Int) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "setCommitRevealConfig", _threshold, _resetInterval)
}

// SetCommitRevealConfig is a paid mutator transaction binding the contract method 0x3f422ef3.
//
// Solidity: function setCommitRevealConfig(uint256 _threshold, uint256 _resetInterval) returns()
func (_Oracle *OracleSession) SetCommitRevealConfig(_threshold *big.Int, _resetInterval *big.Int) (*types.Transaction, error) {
	return _Oracle.Contract.SetCommitRevealConfig(&_Oracle.TransactOpts, _threshold, _resetInterval)
}

// SetCommitRevealConfig is a paid mutator transaction binding the contract method 0x3f422ef3.
//
// Solidity: function setCommitRevealConfig(uint256 _threshold, uint256 _resetInterval) returns()
func (_Oracle *OracleTransactorSession) SetCommitRevealConfig(_threshold *big.Int, _resetInterval *big.Int) (*types.Transaction, error) {
	return _Oracle.Contract.SetCommitRevealConfig(&_Oracle.TransactOpts, _threshold, _resetInterval)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address _operator) returns()
func (_Oracle *OracleTransactor) SetOperator(opts *bind.TransactOpts, _operator common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "setOperator", _operator)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address _operator) returns()
func (_Oracle *OracleSession) SetOperator(_operator common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.SetOperator(&_Oracle.TransactOpts, _operator)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address _operator) returns()
func (_Oracle *OracleTransactorSession) SetOperator(_operator common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.SetOperator(&_Oracle.TransactOpts, _operator)
}

// SetSlashingConfig is a paid mutator transaction binding the contract method 0xda39fbfe.
//
// Solidity: function setSlashingConfig(int256 _outlierSlashingThreshold, int256 _outlierDetectionThreshold, uint256 _baseSlashingRate, uint256 _slashingRateCap) returns()
func (_Oracle *OracleTransactor) SetSlashingConfig(opts *bind.TransactOpts, _outlierSlashingThreshold *big.Int, _outlierDetectionThreshold *big.Int, _baseSlashingRate *big.Int, _slashingRateCap *big.Int) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "setSlashingConfig", _outlierSlashingThreshold, _outlierDetectionThreshold, _baseSlashingRate, _slashingRateCap)
}

// SetSlashingConfig is a paid mutator transaction binding the contract method 0xda39fbfe.
//
// Solidity: function setSlashingConfig(int256 _outlierSlashingThreshold, int256 _outlierDetectionThreshold, uint256 _baseSlashingRate, uint256 _slashingRateCap) returns()
func (_Oracle *OracleSession) SetSlashingConfig(_outlierSlashingThreshold *big.Int, _outlierDetectionThreshold *big.Int, _baseSlashingRate *big.Int, _slashingRateCap *big.Int) (*types.Transaction, error) {
	return _Oracle.Contract.SetSlashingConfig(&_Oracle.TransactOpts, _outlierSlashingThreshold, _outlierDetectionThreshold, _baseSlashingRate, _slashingRateCap)
}

// SetSlashingConfig is a paid mutator transaction binding the contract method 0xda39fbfe.
//
// Solidity: function setSlashingConfig(int256 _outlierSlashingThreshold, int256 _outlierDetectionThreshold, uint256 _baseSlashingRate, uint256 _slashingRateCap) returns()
func (_Oracle *OracleTransactorSession) SetSlashingConfig(_outlierSlashingThreshold *big.Int, _outlierDetectionThreshold *big.Int, _baseSlashingRate *big.Int, _slashingRateCap *big.Int) (*types.Transaction, error) {
	return _Oracle.Contract.SetSlashingConfig(&_Oracle.TransactOpts, _outlierSlashingThreshold, _outlierDetectionThreshold, _baseSlashingRate, _slashingRateCap)
}

// SetSymbols is a paid mutator transaction binding the contract method 0x8d4f75d2.
//
// Solidity: function setSymbols(string[] _symbols) returns()
func (_Oracle *OracleTransactor) SetSymbols(opts *bind.TransactOpts, _symbols []string) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "setSymbols", _symbols)
}

// SetSymbols is a paid mutator transaction binding the contract method 0x8d4f75d2.
//
// Solidity: function setSymbols(string[] _symbols) returns()
func (_Oracle *OracleSession) SetSymbols(_symbols []string) (*types.Transaction, error) {
	return _Oracle.Contract.SetSymbols(&_Oracle.TransactOpts, _symbols)
}

// SetSymbols is a paid mutator transaction binding the contract method 0x8d4f75d2.
//
// Solidity: function setSymbols(string[] _symbols) returns()
func (_Oracle *OracleTransactorSession) SetSymbols(_symbols []string) (*types.Transaction, error) {
	return _Oracle.Contract.SetSymbols(&_Oracle.TransactOpts, _symbols)
}

// SetVoters is a paid mutator transaction binding the contract method 0xda78110e.
//
// Solidity: function setVoters(address[] _newVoters, address[] _treasury, address[] _validator) returns()
func (_Oracle *OracleTransactor) SetVoters(opts *bind.TransactOpts, _newVoters []common.Address, _treasury []common.Address, _validator []common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "setVoters", _newVoters, _treasury, _validator)
}

// SetVoters is a paid mutator transaction binding the contract method 0xda78110e.
//
// Solidity: function setVoters(address[] _newVoters, address[] _treasury, address[] _validator) returns()
func (_Oracle *OracleSession) SetVoters(_newVoters []common.Address, _treasury []common.Address, _validator []common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.SetVoters(&_Oracle.TransactOpts, _newVoters, _treasury, _validator)
}

// SetVoters is a paid mutator transaction binding the contract method 0xda78110e.
//
// Solidity: function setVoters(address[] _newVoters, address[] _treasury, address[] _validator) returns()
func (_Oracle *OracleTransactorSession) SetVoters(_newVoters []common.Address, _treasury []common.Address, _validator []common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.SetVoters(&_Oracle.TransactOpts, _newVoters, _treasury, _validator)
}

// UpdateVotersAndSymbol is a paid mutator transaction binding the contract method 0x0f65875c.
//
// Solidity: function updateVotersAndSymbol() returns()
func (_Oracle *OracleTransactor) UpdateVotersAndSymbol(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "updateVotersAndSymbol")
}

// UpdateVotersAndSymbol is a paid mutator transaction binding the contract method 0x0f65875c.
//
// Solidity: function updateVotersAndSymbol() returns()
func (_Oracle *OracleSession) UpdateVotersAndSymbol() (*types.Transaction, error) {
	return _Oracle.Contract.UpdateVotersAndSymbol(&_Oracle.TransactOpts)
}

// UpdateVotersAndSymbol is a paid mutator transaction binding the contract method 0x0f65875c.
//
// Solidity: function updateVotersAndSymbol() returns()
func (_Oracle *OracleTransactorSession) UpdateVotersAndSymbol() (*types.Transaction, error) {
	return _Oracle.Contract.UpdateVotersAndSymbol(&_Oracle.TransactOpts)
}

// Vote is a paid mutator transaction binding the contract method 0x56833ebe.
//
// Solidity: function vote(uint256 _commit, (uint120,uint8)[] _reports, uint256 _salt, uint8 _extra) returns()
func (_Oracle *OracleTransactor) Vote(opts *bind.TransactOpts, _commit *big.Int, _reports []IOracleReport, _salt *big.Int, _extra uint8) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "vote", _commit, _reports, _salt, _extra)
}

// Vote is a paid mutator transaction binding the contract method 0x56833ebe.
//
// Solidity: function vote(uint256 _commit, (uint120,uint8)[] _reports, uint256 _salt, uint8 _extra) returns()
func (_Oracle *OracleSession) Vote(_commit *big.Int, _reports []IOracleReport, _salt *big.Int, _extra uint8) (*types.Transaction, error) {
	return _Oracle.Contract.Vote(&_Oracle.TransactOpts, _commit, _reports, _salt, _extra)
}

// Vote is a paid mutator transaction binding the contract method 0x56833ebe.
//
// Solidity: function vote(uint256 _commit, (uint120,uint8)[] _reports, uint256 _salt, uint8 _extra) returns()
func (_Oracle *OracleTransactorSession) Vote(_commit *big.Int, _reports []IOracleReport, _salt *big.Int, _extra uint8) (*types.Transaction, error) {
	return _Oracle.Contract.Vote(&_Oracle.TransactOpts, _commit, _reports, _salt, _extra)
}

// OracleCommitRevealMissedIterator is returned from FilterCommitRevealMissed and is used to iterate over the raw logs and unpacked data for CommitRevealMissed events raised by the Oracle contract.
type OracleCommitRevealMissedIterator struct {
	Event *OracleCommitRevealMissed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleCommitRevealMissedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleCommitRevealMissed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleCommitRevealMissed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleCommitRevealMissedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleCommitRevealMissedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleCommitRevealMissed represents a CommitRevealMissed event raised by the Oracle contract.
type OracleCommitRevealMissed struct {
	Voter          common.Address
	Round          *big.Int
	NonRevealCount *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterCommitRevealMissed is a free log retrieval operation binding the contract event 0x176956a4e941f6737f81a3c9a09d8571dd0438d86e25a432beb2013aced43092.
//
// Solidity: event CommitRevealMissed(address indexed _voter, uint256 _round, uint256 _nonRevealCount)
func (_Oracle *OracleFilterer) FilterCommitRevealMissed(opts *bind.FilterOpts, _voter []common.Address) (*OracleCommitRevealMissedIterator, error) {

	var _voterRule []interface{}
	for _, _voterItem := range _voter {
		_voterRule = append(_voterRule, _voterItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "CommitRevealMissed", _voterRule)
	if err != nil {
		return nil, err
	}
	return &OracleCommitRevealMissedIterator{contract: _Oracle.contract, event: "CommitRevealMissed", logs: logs, sub: sub}, nil
}

// WatchCommitRevealMissed is a free log subscription operation binding the contract event 0x176956a4e941f6737f81a3c9a09d8571dd0438d86e25a432beb2013aced43092.
//
// Solidity: event CommitRevealMissed(address indexed _voter, uint256 _round, uint256 _nonRevealCount)
func (_Oracle *OracleFilterer) WatchCommitRevealMissed(opts *bind.WatchOpts, sink chan<- *OracleCommitRevealMissed, _voter []common.Address) (event.Subscription, error) {

	var _voterRule []interface{}
	for _, _voterItem := range _voter {
		_voterRule = append(_voterRule, _voterItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "CommitRevealMissed", _voterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleCommitRevealMissed)
				if err := _Oracle.contract.UnpackLog(event, "CommitRevealMissed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCommitRevealMissed is a log parse operation binding the contract event 0x176956a4e941f6737f81a3c9a09d8571dd0438d86e25a432beb2013aced43092.
//
// Solidity: event CommitRevealMissed(address indexed _voter, uint256 _round, uint256 _nonRevealCount)
func (_Oracle *OracleFilterer) ParseCommitRevealMissed(log types.Log) (*OracleCommitRevealMissed, error) {
	event := new(OracleCommitRevealMissed)
	if err := _Oracle.contract.UnpackLog(event, "CommitRevealMissed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleInvalidVoteIterator is returned from FilterInvalidVote and is used to iterate over the raw logs and unpacked data for InvalidVote events raised by the Oracle contract.
type OracleInvalidVoteIterator struct {
	Event *OracleInvalidVote // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleInvalidVoteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleInvalidVote)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleInvalidVote)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleInvalidVoteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleInvalidVoteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleInvalidVote represents a InvalidVote event raised by the Oracle contract.
type OracleInvalidVote struct {
	Cause       string
	Reporter    common.Address
	ExpValue    *big.Int
	ActualValue *big.Int
	Extra       uint8
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterInvalidVote is a free log retrieval operation binding the contract event 0x04ca4e0efda95f8b780c116574d1521309010b38d8f7b75705495703a0f570b1.
//
// Solidity: event InvalidVote(string cause, address indexed reporter, uint256 expValue, uint256 actualValue, uint8 extra)
func (_Oracle *OracleFilterer) FilterInvalidVote(opts *bind.FilterOpts, reporter []common.Address) (*OracleInvalidVoteIterator, error) {

	var reporterRule []interface{}
	for _, reporterItem := range reporter {
		reporterRule = append(reporterRule, reporterItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "InvalidVote", reporterRule)
	if err != nil {
		return nil, err
	}
	return &OracleInvalidVoteIterator{contract: _Oracle.contract, event: "InvalidVote", logs: logs, sub: sub}, nil
}

// WatchInvalidVote is a free log subscription operation binding the contract event 0x04ca4e0efda95f8b780c116574d1521309010b38d8f7b75705495703a0f570b1.
//
// Solidity: event InvalidVote(string cause, address indexed reporter, uint256 expValue, uint256 actualValue, uint8 extra)
func (_Oracle *OracleFilterer) WatchInvalidVote(opts *bind.WatchOpts, sink chan<- *OracleInvalidVote, reporter []common.Address) (event.Subscription, error) {

	var reporterRule []interface{}
	for _, reporterItem := range reporter {
		reporterRule = append(reporterRule, reporterItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "InvalidVote", reporterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleInvalidVote)
				if err := _Oracle.contract.UnpackLog(event, "InvalidVote", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInvalidVote is a log parse operation binding the contract event 0x04ca4e0efda95f8b780c116574d1521309010b38d8f7b75705495703a0f570b1.
//
// Solidity: event InvalidVote(string cause, address indexed reporter, uint256 expValue, uint256 actualValue, uint8 extra)
func (_Oracle *OracleFilterer) ParseInvalidVote(log types.Log) (*OracleInvalidVote, error) {
	event := new(OracleInvalidVote)
	if err := _Oracle.contract.UnpackLog(event, "InvalidVote", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleNewRoundIterator is returned from FilterNewRound and is used to iterate over the raw logs and unpacked data for NewRound events raised by the Oracle contract.
type OracleNewRoundIterator struct {
	Event *OracleNewRound // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleNewRoundIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleNewRound)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleNewRound)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleNewRoundIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleNewRoundIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleNewRound represents a NewRound event raised by the Oracle contract.
type OracleNewRound struct {
	Round      *big.Int
	Timestamp  *big.Int
	VotePeriod *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterNewRound is a free log retrieval operation binding the contract event 0x5aec57d81928b24d30b1a2aec0d23d693412c37d7ec106b5d8259413716bb1f4.
//
// Solidity: event NewRound(uint256 _round, uint256 _timestamp, uint256 _votePeriod)
func (_Oracle *OracleFilterer) FilterNewRound(opts *bind.FilterOpts) (*OracleNewRoundIterator, error) {

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "NewRound")
	if err != nil {
		return nil, err
	}
	return &OracleNewRoundIterator{contract: _Oracle.contract, event: "NewRound", logs: logs, sub: sub}, nil
}

// WatchNewRound is a free log subscription operation binding the contract event 0x5aec57d81928b24d30b1a2aec0d23d693412c37d7ec106b5d8259413716bb1f4.
//
// Solidity: event NewRound(uint256 _round, uint256 _timestamp, uint256 _votePeriod)
func (_Oracle *OracleFilterer) WatchNewRound(opts *bind.WatchOpts, sink chan<- *OracleNewRound) (event.Subscription, error) {

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "NewRound")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleNewRound)
				if err := _Oracle.contract.UnpackLog(event, "NewRound", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewRound is a log parse operation binding the contract event 0x5aec57d81928b24d30b1a2aec0d23d693412c37d7ec106b5d8259413716bb1f4.
//
// Solidity: event NewRound(uint256 _round, uint256 _timestamp, uint256 _votePeriod)
func (_Oracle *OracleFilterer) ParseNewRound(log types.Log) (*OracleNewRound, error) {
	event := new(OracleNewRound)
	if err := _Oracle.contract.UnpackLog(event, "NewRound", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleNewSymbolsIterator is returned from FilterNewSymbols and is used to iterate over the raw logs and unpacked data for NewSymbols events raised by the Oracle contract.
type OracleNewSymbolsIterator struct {
	Event *OracleNewSymbols // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleNewSymbolsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleNewSymbols)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleNewSymbols)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleNewSymbolsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleNewSymbolsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleNewSymbols represents a NewSymbols event raised by the Oracle contract.
type OracleNewSymbols struct {
	Symbols []string
	Round   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterNewSymbols is a free log retrieval operation binding the contract event 0xaa278e424da680ce5dad66510415760e78e0bd87d45c786c6e88bdde82f9342d.
//
// Solidity: event NewSymbols(string[] _symbols, uint256 _round)
func (_Oracle *OracleFilterer) FilterNewSymbols(opts *bind.FilterOpts) (*OracleNewSymbolsIterator, error) {

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "NewSymbols")
	if err != nil {
		return nil, err
	}
	return &OracleNewSymbolsIterator{contract: _Oracle.contract, event: "NewSymbols", logs: logs, sub: sub}, nil
}

// WatchNewSymbols is a free log subscription operation binding the contract event 0xaa278e424da680ce5dad66510415760e78e0bd87d45c786c6e88bdde82f9342d.
//
// Solidity: event NewSymbols(string[] _symbols, uint256 _round)
func (_Oracle *OracleFilterer) WatchNewSymbols(opts *bind.WatchOpts, sink chan<- *OracleNewSymbols) (event.Subscription, error) {

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "NewSymbols")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleNewSymbols)
				if err := _Oracle.contract.UnpackLog(event, "NewSymbols", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewSymbols is a log parse operation binding the contract event 0xaa278e424da680ce5dad66510415760e78e0bd87d45c786c6e88bdde82f9342d.
//
// Solidity: event NewSymbols(string[] _symbols, uint256 _round)
func (_Oracle *OracleFilterer) ParseNewSymbols(log types.Log) (*OracleNewSymbols, error) {
	event := new(OracleNewSymbols)
	if err := _Oracle.contract.UnpackLog(event, "NewSymbols", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleNewVoterIterator is returned from FilterNewVoter and is used to iterate over the raw logs and unpacked data for NewVoter events raised by the Oracle contract.
type OracleNewVoterIterator struct {
	Event *OracleNewVoter // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleNewVoterIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleNewVoter)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleNewVoter)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleNewVoterIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleNewVoterIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleNewVoter represents a NewVoter event raised by the Oracle contract.
type OracleNewVoter struct {
	Reporter common.Address
	Extra    uint8
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterNewVoter is a free log retrieval operation binding the contract event 0xd2ec8e890a03083998d3e16f98044fd3dd13fe3e61b7bc2e58ee6da43b50af73.
//
// Solidity: event NewVoter(address reporter, uint8 extra)
func (_Oracle *OracleFilterer) FilterNewVoter(opts *bind.FilterOpts) (*OracleNewVoterIterator, error) {

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "NewVoter")
	if err != nil {
		return nil, err
	}
	return &OracleNewVoterIterator{contract: _Oracle.contract, event: "NewVoter", logs: logs, sub: sub}, nil
}

// WatchNewVoter is a free log subscription operation binding the contract event 0xd2ec8e890a03083998d3e16f98044fd3dd13fe3e61b7bc2e58ee6da43b50af73.
//
// Solidity: event NewVoter(address reporter, uint8 extra)
func (_Oracle *OracleFilterer) WatchNewVoter(opts *bind.WatchOpts, sink chan<- *OracleNewVoter) (event.Subscription, error) {

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "NewVoter")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleNewVoter)
				if err := _Oracle.contract.UnpackLog(event, "NewVoter", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewVoter is a log parse operation binding the contract event 0xd2ec8e890a03083998d3e16f98044fd3dd13fe3e61b7bc2e58ee6da43b50af73.
//
// Solidity: event NewVoter(address reporter, uint8 extra)
func (_Oracle *OracleFilterer) ParseNewVoter(log types.Log) (*OracleNewVoter, error) {
	event := new(OracleNewVoter)
	if err := _Oracle.contract.UnpackLog(event, "NewVoter", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleNoRevealPenaltyIterator is returned from FilterNoRevealPenalty and is used to iterate over the raw logs and unpacked data for NoRevealPenalty events raised by the Oracle contract.
type OracleNoRevealPenaltyIterator struct {
	Event *OracleNoRevealPenalty // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleNoRevealPenaltyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleNoRevealPenalty)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleNoRevealPenalty)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleNoRevealPenaltyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleNoRevealPenaltyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleNoRevealPenalty represents a NoRevealPenalty event raised by the Oracle contract.
type OracleNoRevealPenalty struct {
	Voter        common.Address
	Round        *big.Int
	MissedReveal *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterNoRevealPenalty is a free log retrieval operation binding the contract event 0x9e6b40f10c60d1ad09594f3b6ed7043d0e978f584d354ace6e1f6025660c42b1.
//
// Solidity: event NoRevealPenalty(address indexed _voter, uint256 _round, uint256 _missedReveal)
func (_Oracle *OracleFilterer) FilterNoRevealPenalty(opts *bind.FilterOpts, _voter []common.Address) (*OracleNoRevealPenaltyIterator, error) {

	var _voterRule []interface{}
	for _, _voterItem := range _voter {
		_voterRule = append(_voterRule, _voterItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "NoRevealPenalty", _voterRule)
	if err != nil {
		return nil, err
	}
	return &OracleNoRevealPenaltyIterator{contract: _Oracle.contract, event: "NoRevealPenalty", logs: logs, sub: sub}, nil
}

// WatchNoRevealPenalty is a free log subscription operation binding the contract event 0x9e6b40f10c60d1ad09594f3b6ed7043d0e978f584d354ace6e1f6025660c42b1.
//
// Solidity: event NoRevealPenalty(address indexed _voter, uint256 _round, uint256 _missedReveal)
func (_Oracle *OracleFilterer) WatchNoRevealPenalty(opts *bind.WatchOpts, sink chan<- *OracleNoRevealPenalty, _voter []common.Address) (event.Subscription, error) {

	var _voterRule []interface{}
	for _, _voterItem := range _voter {
		_voterRule = append(_voterRule, _voterItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "NoRevealPenalty", _voterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleNoRevealPenalty)
				if err := _Oracle.contract.UnpackLog(event, "NoRevealPenalty", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNoRevealPenalty is a log parse operation binding the contract event 0x9e6b40f10c60d1ad09594f3b6ed7043d0e978f584d354ace6e1f6025660c42b1.
//
// Solidity: event NoRevealPenalty(address indexed _voter, uint256 _round, uint256 _missedReveal)
func (_Oracle *OracleFilterer) ParseNoRevealPenalty(log types.Log) (*OracleNoRevealPenalty, error) {
	event := new(OracleNoRevealPenalty)
	if err := _Oracle.contract.UnpackLog(event, "NoRevealPenalty", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OraclePenalizedIterator is returned from FilterPenalized and is used to iterate over the raw logs and unpacked data for Penalized events raised by the Oracle contract.
type OraclePenalizedIterator struct {
	Event *OraclePenalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OraclePenalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OraclePenalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OraclePenalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OraclePenalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OraclePenalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OraclePenalized represents a Penalized event raised by the Oracle contract.
type OraclePenalized struct {
	Participant    common.Address
	SlashingAmount *big.Int
	Symbol         string
	Median         *big.Int
	Reported       *big.Int
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterPenalized is a free log retrieval operation binding the contract event 0x372858b237c8bd0714183e8351a461d6c3cb1ef83806181b36bf5943711f4f57.
//
// Solidity: event Penalized(address indexed _participant, uint256 _slashingAmount, string _symbol, int256 _median, uint120 _reported)
func (_Oracle *OracleFilterer) FilterPenalized(opts *bind.FilterOpts, _participant []common.Address) (*OraclePenalizedIterator, error) {

	var _participantRule []interface{}
	for _, _participantItem := range _participant {
		_participantRule = append(_participantRule, _participantItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "Penalized", _participantRule)
	if err != nil {
		return nil, err
	}
	return &OraclePenalizedIterator{contract: _Oracle.contract, event: "Penalized", logs: logs, sub: sub}, nil
}

// WatchPenalized is a free log subscription operation binding the contract event 0x372858b237c8bd0714183e8351a461d6c3cb1ef83806181b36bf5943711f4f57.
//
// Solidity: event Penalized(address indexed _participant, uint256 _slashingAmount, string _symbol, int256 _median, uint120 _reported)
func (_Oracle *OracleFilterer) WatchPenalized(opts *bind.WatchOpts, sink chan<- *OraclePenalized, _participant []common.Address) (event.Subscription, error) {

	var _participantRule []interface{}
	for _, _participantItem := range _participant {
		_participantRule = append(_participantRule, _participantItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "Penalized", _participantRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OraclePenalized)
				if err := _Oracle.contract.UnpackLog(event, "Penalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePenalized is a log parse operation binding the contract event 0x372858b237c8bd0714183e8351a461d6c3cb1ef83806181b36bf5943711f4f57.
//
// Solidity: event Penalized(address indexed _participant, uint256 _slashingAmount, string _symbol, int256 _median, uint120 _reported)
func (_Oracle *OracleFilterer) ParsePenalized(log types.Log) (*OraclePenalized, error) {
	event := new(OraclePenalized)
	if err := _Oracle.contract.UnpackLog(event, "Penalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OraclePriceUpdatedIterator is returned from FilterPriceUpdated and is used to iterate over the raw logs and unpacked data for PriceUpdated events raised by the Oracle contract.
type OraclePriceUpdatedIterator struct {
	Event *OraclePriceUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OraclePriceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OraclePriceUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OraclePriceUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OraclePriceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OraclePriceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OraclePriceUpdated represents a PriceUpdated event raised by the Oracle contract.
type OraclePriceUpdated struct {
	Price     *big.Int
	Round     *big.Int
	Symbol    common.Hash
	Status    bool
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPriceUpdated is a free log retrieval operation binding the contract event 0x5f2aa51aa7889ad71d9318fa7fd83c8ff3277434249bd06073f15986e197911c.
//
// Solidity: event PriceUpdated(uint256 price, uint256 round, string indexed symbol, bool status, uint256 timestamp)
func (_Oracle *OracleFilterer) FilterPriceUpdated(opts *bind.FilterOpts, symbol []string) (*OraclePriceUpdatedIterator, error) {

	var symbolRule []interface{}
	for _, symbolItem := range symbol {
		symbolRule = append(symbolRule, symbolItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "PriceUpdated", symbolRule)
	if err != nil {
		return nil, err
	}
	return &OraclePriceUpdatedIterator{contract: _Oracle.contract, event: "PriceUpdated", logs: logs, sub: sub}, nil
}

// WatchPriceUpdated is a free log subscription operation binding the contract event 0x5f2aa51aa7889ad71d9318fa7fd83c8ff3277434249bd06073f15986e197911c.
//
// Solidity: event PriceUpdated(uint256 price, uint256 round, string indexed symbol, bool status, uint256 timestamp)
func (_Oracle *OracleFilterer) WatchPriceUpdated(opts *bind.WatchOpts, sink chan<- *OraclePriceUpdated, symbol []string) (event.Subscription, error) {

	var symbolRule []interface{}
	for _, symbolItem := range symbol {
		symbolRule = append(symbolRule, symbolItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "PriceUpdated", symbolRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OraclePriceUpdated)
				if err := _Oracle.contract.UnpackLog(event, "PriceUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePriceUpdated is a log parse operation binding the contract event 0x5f2aa51aa7889ad71d9318fa7fd83c8ff3277434249bd06073f15986e197911c.
//
// Solidity: event PriceUpdated(uint256 price, uint256 round, string indexed symbol, bool status, uint256 timestamp)
func (_Oracle *OracleFilterer) ParsePriceUpdated(log types.Log) (*OraclePriceUpdated, error) {
	event := new(OraclePriceUpdated)
	if err := _Oracle.contract.UnpackLog(event, "PriceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleSuccessfulVoteIterator is returned from FilterSuccessfulVote and is used to iterate over the raw logs and unpacked data for SuccessfulVote events raised by the Oracle contract.
type OracleSuccessfulVoteIterator struct {
	Event *OracleSuccessfulVote // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleSuccessfulVoteIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleSuccessfulVote)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleSuccessfulVote)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleSuccessfulVoteIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleSuccessfulVoteIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleSuccessfulVote represents a SuccessfulVote event raised by the Oracle contract.
type OracleSuccessfulVote struct {
	Reporter common.Address
	Extra    uint8
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSuccessfulVote is a free log retrieval operation binding the contract event 0x8bdddd7f2f2c74679ffa6beb8f86aa18bfa5baf1bfaf534d0b66596babc53f08.
//
// Solidity: event SuccessfulVote(address indexed reporter, uint8 extra)
func (_Oracle *OracleFilterer) FilterSuccessfulVote(opts *bind.FilterOpts, reporter []common.Address) (*OracleSuccessfulVoteIterator, error) {

	var reporterRule []interface{}
	for _, reporterItem := range reporter {
		reporterRule = append(reporterRule, reporterItem)
	}

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "SuccessfulVote", reporterRule)
	if err != nil {
		return nil, err
	}
	return &OracleSuccessfulVoteIterator{contract: _Oracle.contract, event: "SuccessfulVote", logs: logs, sub: sub}, nil
}

// WatchSuccessfulVote is a free log subscription operation binding the contract event 0x8bdddd7f2f2c74679ffa6beb8f86aa18bfa5baf1bfaf534d0b66596babc53f08.
//
// Solidity: event SuccessfulVote(address indexed reporter, uint8 extra)
func (_Oracle *OracleFilterer) WatchSuccessfulVote(opts *bind.WatchOpts, sink chan<- *OracleSuccessfulVote, reporter []common.Address) (event.Subscription, error) {

	var reporterRule []interface{}
	for _, reporterItem := range reporter {
		reporterRule = append(reporterRule, reporterItem)
	}

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "SuccessfulVote", reporterRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleSuccessfulVote)
				if err := _Oracle.contract.UnpackLog(event, "SuccessfulVote", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSuccessfulVote is a log parse operation binding the contract event 0x8bdddd7f2f2c74679ffa6beb8f86aa18bfa5baf1bfaf534d0b66596babc53f08.
//
// Solidity: event SuccessfulVote(address indexed reporter, uint8 extra)
func (_Oracle *OracleFilterer) ParseSuccessfulVote(log types.Log) (*OracleSuccessfulVote, error) {
	event := new(OracleSuccessfulVote)
	if err := _Oracle.contract.UnpackLog(event, "SuccessfulVote", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleTotalOracleRewardsIterator is returned from FilterTotalOracleRewards and is used to iterate over the raw logs and unpacked data for TotalOracleRewards events raised by the Oracle contract.
type OracleTotalOracleRewardsIterator struct {
	Event *OracleTotalOracleRewards // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *OracleTotalOracleRewardsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleTotalOracleRewards)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(OracleTotalOracleRewards)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *OracleTotalOracleRewardsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleTotalOracleRewardsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleTotalOracleRewards represents a TotalOracleRewards event raised by the Oracle contract.
type OracleTotalOracleRewards struct {
	NtnReward *big.Int
	AtnReward *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTotalOracleRewards is a free log retrieval operation binding the contract event 0x3e5aaff9e8fd4293ae18127809c2d4069d87fe10c7de92aa39557a1edbd48fec.
//
// Solidity: event TotalOracleRewards(uint256 ntnReward, uint256 atnReward)
func (_Oracle *OracleFilterer) FilterTotalOracleRewards(opts *bind.FilterOpts) (*OracleTotalOracleRewardsIterator, error) {

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "TotalOracleRewards")
	if err != nil {
		return nil, err
	}
	return &OracleTotalOracleRewardsIterator{contract: _Oracle.contract, event: "TotalOracleRewards", logs: logs, sub: sub}, nil
}

// WatchTotalOracleRewards is a free log subscription operation binding the contract event 0x3e5aaff9e8fd4293ae18127809c2d4069d87fe10c7de92aa39557a1edbd48fec.
//
// Solidity: event TotalOracleRewards(uint256 ntnReward, uint256 atnReward)
func (_Oracle *OracleFilterer) WatchTotalOracleRewards(opts *bind.WatchOpts, sink chan<- *OracleTotalOracleRewards) (event.Subscription, error) {

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "TotalOracleRewards")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleTotalOracleRewards)
				if err := _Oracle.contract.UnpackLog(event, "TotalOracleRewards", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTotalOracleRewards is a log parse operation binding the contract event 0x3e5aaff9e8fd4293ae18127809c2d4069d87fe10c7de92aa39557a1edbd48fec.
//
// Solidity: event TotalOracleRewards(uint256 ntnReward, uint256 atnReward)
func (_Oracle *OracleFilterer) ParseTotalOracleRewards(log types.Log) (*OracleTotalOracleRewards, error) {
	event := new(OracleTotalOracleRewards)
	if err := _Oracle.contract.UnpackLog(event, "TotalOracleRewards", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
