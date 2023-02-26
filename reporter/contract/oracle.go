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

// OracleMetaData contains all meta data concerning the Oracle contract.
var OracleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_committee\",\"type\":\"address[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"committee\",\"type\":\"address[]\"}],\"name\":\"UpdatedCommittee\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"round\",\"type\":\"uint256\"}],\"name\":\"UpdatedRound\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"symbols\",\"type\":\"string[]\"}],\"name\":\"UpdatedSymbols\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"commits\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"finalize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCommittee\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRound\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_round\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"getRoundData\",\"outputs\":[{\"internalType\":\"int256\",\"name\":\"\",\"type\":\"int256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getSymbols\",\"outputs\":[{\"internalType\":\"string[]\",\"name\":\"\",\"type\":\"string[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"latestData\",\"outputs\":[{\"internalType\":\"int256\",\"name\":\"\",\"type\":\"int256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"newSymbols\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"prevotes\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"round\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"_newCommittee\",\"type\":\"address[]\"}],\"name\":\"setCommittee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string[]\",\"name\":\"_symbols\",\"type\":\"string[]\"}],\"name\":\"setSymbols\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"symbols\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_commit\",\"type\":\"uint256\"},{\"internalType\":\"int256[]\",\"name\":\"_prevotes\",\"type\":\"int256[]\"}],\"name\":\"vote\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"voted\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
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

// Commits is a free data retrieval call binding the contract method 0x7b43a8e6.
//
// Solidity: function commits(address ) view returns(uint256)
func (_Oracle *OracleCaller) Commits(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "commits", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Commits is a free data retrieval call binding the contract method 0x7b43a8e6.
//
// Solidity: function commits(address ) view returns(uint256)
func (_Oracle *OracleSession) Commits(arg0 common.Address) (*big.Int, error) {
	return _Oracle.Contract.Commits(&_Oracle.CallOpts, arg0)
}

// Commits is a free data retrieval call binding the contract method 0x7b43a8e6.
//
// Solidity: function commits(address ) view returns(uint256)
func (_Oracle *OracleCallerSession) Commits(arg0 common.Address) (*big.Int, error) {
	return _Oracle.Contract.Commits(&_Oracle.CallOpts, arg0)
}

// GetCommittee is a free data retrieval call binding the contract method 0xab8f6ffe.
//
// Solidity: function getCommittee() view returns(address[])
func (_Oracle *OracleCaller) GetCommittee(opts *bind.CallOpts) ([]common.Address, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getCommittee")

	if err != nil {
		return *new([]common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)

	return out0, err

}

// GetCommittee is a free data retrieval call binding the contract method 0xab8f6ffe.
//
// Solidity: function getCommittee() view returns(address[])
func (_Oracle *OracleSession) GetCommittee() ([]common.Address, error) {
	return _Oracle.Contract.GetCommittee(&_Oracle.CallOpts)
}

// GetCommittee is a free data retrieval call binding the contract method 0xab8f6ffe.
//
// Solidity: function getCommittee() view returns(address[])
func (_Oracle *OracleCallerSession) GetCommittee() ([]common.Address, error) {
	return _Oracle.Contract.GetCommittee(&_Oracle.CallOpts)
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
// Solidity: function getRoundData(uint256 _round, string _symbol) view returns(int256, uint256, uint256)
func (_Oracle *OracleCaller) GetRoundData(opts *bind.CallOpts, _round *big.Int, _symbol string) (*big.Int, *big.Int, *big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "getRoundData", _round, _symbol)

	if err != nil {
		return *new(*big.Int), *new(*big.Int), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	out2 := *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return out0, out1, out2, err

}

// GetRoundData is a free data retrieval call binding the contract method 0x3c8510fd.
//
// Solidity: function getRoundData(uint256 _round, string _symbol) view returns(int256, uint256, uint256)
func (_Oracle *OracleSession) GetRoundData(_round *big.Int, _symbol string) (*big.Int, *big.Int, *big.Int, error) {
	return _Oracle.Contract.GetRoundData(&_Oracle.CallOpts, _round, _symbol)
}

// GetRoundData is a free data retrieval call binding the contract method 0x3c8510fd.
//
// Solidity: function getRoundData(uint256 _round, string _symbol) view returns(int256, uint256, uint256)
func (_Oracle *OracleCallerSession) GetRoundData(_round *big.Int, _symbol string) (*big.Int, *big.Int, *big.Int, error) {
	return _Oracle.Contract.GetRoundData(&_Oracle.CallOpts, _round, _symbol)
}

// GetSymbols is a free data retrieval call binding the contract method 0xdf7f710e.
//
// Solidity: function getSymbols() view returns(string[])
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
// Solidity: function getSymbols() view returns(string[])
func (_Oracle *OracleSession) GetSymbols() ([]string, error) {
	return _Oracle.Contract.GetSymbols(&_Oracle.CallOpts)
}

// GetSymbols is a free data retrieval call binding the contract method 0xdf7f710e.
//
// Solidity: function getSymbols() view returns(string[])
func (_Oracle *OracleCallerSession) GetSymbols() ([]string, error) {
	return _Oracle.Contract.GetSymbols(&_Oracle.CallOpts)
}

// LatestData is a free data retrieval call binding the contract method 0x9250cfaf.
//
// Solidity: function latestData(string _symbol) view returns(int256, uint256, uint256)
func (_Oracle *OracleCaller) LatestData(opts *bind.CallOpts, _symbol string) (*big.Int, *big.Int, *big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "latestData", _symbol)

	if err != nil {
		return *new(*big.Int), *new(*big.Int), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	out2 := *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return out0, out1, out2, err

}

// LatestData is a free data retrieval call binding the contract method 0x9250cfaf.
//
// Solidity: function latestData(string _symbol) view returns(int256, uint256, uint256)
func (_Oracle *OracleSession) LatestData(_symbol string) (*big.Int, *big.Int, *big.Int, error) {
	return _Oracle.Contract.LatestData(&_Oracle.CallOpts, _symbol)
}

// LatestData is a free data retrieval call binding the contract method 0x9250cfaf.
//
// Solidity: function latestData(string _symbol) view returns(int256, uint256, uint256)
func (_Oracle *OracleCallerSession) LatestData(_symbol string) (*big.Int, *big.Int, *big.Int, error) {
	return _Oracle.Contract.LatestData(&_Oracle.CallOpts, _symbol)
}

// NewSymbols is a free data retrieval call binding the contract method 0x5281b5c6.
//
// Solidity: function newSymbols(uint256 ) view returns(string)
func (_Oracle *OracleCaller) NewSymbols(opts *bind.CallOpts, arg0 *big.Int) (string, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "newSymbols", arg0)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// NewSymbols is a free data retrieval call binding the contract method 0x5281b5c6.
//
// Solidity: function newSymbols(uint256 ) view returns(string)
func (_Oracle *OracleSession) NewSymbols(arg0 *big.Int) (string, error) {
	return _Oracle.Contract.NewSymbols(&_Oracle.CallOpts, arg0)
}

// NewSymbols is a free data retrieval call binding the contract method 0x5281b5c6.
//
// Solidity: function newSymbols(uint256 ) view returns(string)
func (_Oracle *OracleCallerSession) NewSymbols(arg0 *big.Int) (string, error) {
	return _Oracle.Contract.NewSymbols(&_Oracle.CallOpts, arg0)
}

// Prevotes is a free data retrieval call binding the contract method 0x220610ab.
//
// Solidity: function prevotes(string , address ) view returns(uint256)
func (_Oracle *OracleCaller) Prevotes(opts *bind.CallOpts, arg0 string, arg1 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "prevotes", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Prevotes is a free data retrieval call binding the contract method 0x220610ab.
//
// Solidity: function prevotes(string , address ) view returns(uint256)
func (_Oracle *OracleSession) Prevotes(arg0 string, arg1 common.Address) (*big.Int, error) {
	return _Oracle.Contract.Prevotes(&_Oracle.CallOpts, arg0, arg1)
}

// Prevotes is a free data retrieval call binding the contract method 0x220610ab.
//
// Solidity: function prevotes(string , address ) view returns(uint256)
func (_Oracle *OracleCallerSession) Prevotes(arg0 string, arg1 common.Address) (*big.Int, error) {
	return _Oracle.Contract.Prevotes(&_Oracle.CallOpts, arg0, arg1)
}

// Round is a free data retrieval call binding the contract method 0x146ca531.
//
// Solidity: function round() view returns(uint256)
func (_Oracle *OracleCaller) Round(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "round")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Round is a free data retrieval call binding the contract method 0x146ca531.
//
// Solidity: function round() view returns(uint256)
func (_Oracle *OracleSession) Round() (*big.Int, error) {
	return _Oracle.Contract.Round(&_Oracle.CallOpts)
}

// Round is a free data retrieval call binding the contract method 0x146ca531.
//
// Solidity: function round() view returns(uint256)
func (_Oracle *OracleCallerSession) Round() (*big.Int, error) {
	return _Oracle.Contract.Round(&_Oracle.CallOpts)
}

// Symbols is a free data retrieval call binding the contract method 0xccce413b.
//
// Solidity: function symbols(uint256 ) view returns(string)
func (_Oracle *OracleCaller) Symbols(opts *bind.CallOpts, arg0 *big.Int) (string, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "symbols", arg0)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbols is a free data retrieval call binding the contract method 0xccce413b.
//
// Solidity: function symbols(uint256 ) view returns(string)
func (_Oracle *OracleSession) Symbols(arg0 *big.Int) (string, error) {
	return _Oracle.Contract.Symbols(&_Oracle.CallOpts, arg0)
}

// Symbols is a free data retrieval call binding the contract method 0xccce413b.
//
// Solidity: function symbols(uint256 ) view returns(string)
func (_Oracle *OracleCallerSession) Symbols(arg0 *big.Int) (string, error) {
	return _Oracle.Contract.Symbols(&_Oracle.CallOpts, arg0)
}

// Voted is a free data retrieval call binding the contract method 0xaec2ccae.
//
// Solidity: function voted(address ) view returns(uint256)
func (_Oracle *OracleCaller) Voted(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _Oracle.contract.Call(opts, &out, "voted", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Voted is a free data retrieval call binding the contract method 0xaec2ccae.
//
// Solidity: function voted(address ) view returns(uint256)
func (_Oracle *OracleSession) Voted(arg0 common.Address) (*big.Int, error) {
	return _Oracle.Contract.Voted(&_Oracle.CallOpts, arg0)
}

// Voted is a free data retrieval call binding the contract method 0xaec2ccae.
//
// Solidity: function voted(address ) view returns(uint256)
func (_Oracle *OracleCallerSession) Voted(arg0 common.Address) (*big.Int, error) {
	return _Oracle.Contract.Voted(&_Oracle.CallOpts, arg0)
}

// Finalize is a paid mutator transaction binding the contract method 0x4bb278f3.
//
// Solidity: function finalize() returns()
func (_Oracle *OracleTransactor) Finalize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "finalize")
}

// Finalize is a paid mutator transaction binding the contract method 0x4bb278f3.
//
// Solidity: function finalize() returns()
func (_Oracle *OracleSession) Finalize() (*types.Transaction, error) {
	return _Oracle.Contract.Finalize(&_Oracle.TransactOpts)
}

// Finalize is a paid mutator transaction binding the contract method 0x4bb278f3.
//
// Solidity: function finalize() returns()
func (_Oracle *OracleTransactorSession) Finalize() (*types.Transaction, error) {
	return _Oracle.Contract.Finalize(&_Oracle.TransactOpts)
}

// SetCommittee is a paid mutator transaction binding the contract method 0xe08b14ed.
//
// Solidity: function setCommittee(address[] _newCommittee) returns()
func (_Oracle *OracleTransactor) SetCommittee(opts *bind.TransactOpts, _newCommittee []common.Address) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "setCommittee", _newCommittee)
}

// SetCommittee is a paid mutator transaction binding the contract method 0xe08b14ed.
//
// Solidity: function setCommittee(address[] _newCommittee) returns()
func (_Oracle *OracleSession) SetCommittee(_newCommittee []common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.SetCommittee(&_Oracle.TransactOpts, _newCommittee)
}

// SetCommittee is a paid mutator transaction binding the contract method 0xe08b14ed.
//
// Solidity: function setCommittee(address[] _newCommittee) returns()
func (_Oracle *OracleTransactorSession) SetCommittee(_newCommittee []common.Address) (*types.Transaction, error) {
	return _Oracle.Contract.SetCommittee(&_Oracle.TransactOpts, _newCommittee)
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

// Vote is a paid mutator transaction binding the contract method 0x94803508.
//
// Solidity: function vote(uint256 _commit, int256[] _prevotes, uint256 _salt) returns()
func (_Oracle *OracleTransactor) Vote(opts *bind.TransactOpts, _commit *big.Int, _prevotes []*big.Int, _salt *big.Int) (*types.Transaction, error) {
	return _Oracle.contract.Transact(opts, "vote", _commit, _prevotes, _salt)
}

// Vote is a paid mutator transaction binding the contract method 0x94803508.
//
// Solidity: function vote(uint256 _commit, int256[] _prevotes, uint256 _salt) returns()
func (_Oracle *OracleSession) Vote(_commit *big.Int, _prevotes []*big.Int, _salt *big.Int) (*types.Transaction, error) {
	return _Oracle.Contract.Vote(&_Oracle.TransactOpts, _commit, _prevotes, _salt)
}

// Vote is a paid mutator transaction binding the contract method 0x94803508.
//
// Solidity: function vote(uint256 _commit, int256[] _prevotes, uint256 _salt) returns()
func (_Oracle *OracleTransactorSession) Vote(_commit *big.Int, _prevotes []*big.Int, _salt *big.Int) (*types.Transaction, error) {
	return _Oracle.Contract.Vote(&_Oracle.TransactOpts, _commit, _prevotes, _salt)
}

// OracleUpdatedCommitteeIterator is returned from FilterUpdatedCommittee and is used to iterate over the raw logs and unpacked data for UpdatedCommittee events raised by the Oracle contract.
type OracleUpdatedCommitteeIterator struct {
	Event *OracleUpdatedCommittee // Event containing the contract specifics and raw log

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
func (it *OracleUpdatedCommitteeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleUpdatedCommittee)
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
		it.Event = new(OracleUpdatedCommittee)
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
func (it *OracleUpdatedCommitteeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleUpdatedCommitteeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleUpdatedCommittee represents a UpdatedCommittee event raised by the Oracle contract.
type OracleUpdatedCommittee struct {
	Committee []common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterUpdatedCommittee is a free log retrieval operation binding the contract event 0x8e0e962b9f0451981ba928b023e18a9e70aa6709ff576587d32053b0310bde22.
//
// Solidity: event UpdatedCommittee(address[] committee)
func (_Oracle *OracleFilterer) FilterUpdatedCommittee(opts *bind.FilterOpts) (*OracleUpdatedCommitteeIterator, error) {

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "UpdatedCommittee")
	if err != nil {
		return nil, err
	}
	return &OracleUpdatedCommitteeIterator{contract: _Oracle.contract, event: "UpdatedCommittee", logs: logs, sub: sub}, nil
}

// WatchUpdatedCommittee is a free log subscription operation binding the contract event 0x8e0e962b9f0451981ba928b023e18a9e70aa6709ff576587d32053b0310bde22.
//
// Solidity: event UpdatedCommittee(address[] committee)
func (_Oracle *OracleFilterer) WatchUpdatedCommittee(opts *bind.WatchOpts, sink chan<- *OracleUpdatedCommittee) (event.Subscription, error) {

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "UpdatedCommittee")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleUpdatedCommittee)
				if err := _Oracle.contract.UnpackLog(event, "UpdatedCommittee", log); err != nil {
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

// ParseUpdatedCommittee is a log parse operation binding the contract event 0x8e0e962b9f0451981ba928b023e18a9e70aa6709ff576587d32053b0310bde22.
//
// Solidity: event UpdatedCommittee(address[] committee)
func (_Oracle *OracleFilterer) ParseUpdatedCommittee(log types.Log) (*OracleUpdatedCommittee, error) {
	event := new(OracleUpdatedCommittee)
	if err := _Oracle.contract.UnpackLog(event, "UpdatedCommittee", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleUpdatedRoundIterator is returned from FilterUpdatedRound and is used to iterate over the raw logs and unpacked data for UpdatedRound events raised by the Oracle contract.
type OracleUpdatedRoundIterator struct {
	Event *OracleUpdatedRound // Event containing the contract specifics and raw log

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
func (it *OracleUpdatedRoundIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleUpdatedRound)
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
		it.Event = new(OracleUpdatedRound)
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
func (it *OracleUpdatedRoundIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleUpdatedRoundIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleUpdatedRound represents a UpdatedRound event raised by the Oracle contract.
type OracleUpdatedRound struct {
	Round *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterUpdatedRound is a free log retrieval operation binding the contract event 0x346e872acb7563c524fd9f17511221d966247f7c39abdc52fcafccc289f5c217.
//
// Solidity: event UpdatedRound(uint256 round)
func (_Oracle *OracleFilterer) FilterUpdatedRound(opts *bind.FilterOpts) (*OracleUpdatedRoundIterator, error) {

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "UpdatedRound")
	if err != nil {
		return nil, err
	}
	return &OracleUpdatedRoundIterator{contract: _Oracle.contract, event: "UpdatedRound", logs: logs, sub: sub}, nil
}

// WatchUpdatedRound is a free log subscription operation binding the contract event 0x346e872acb7563c524fd9f17511221d966247f7c39abdc52fcafccc289f5c217.
//
// Solidity: event UpdatedRound(uint256 round)
func (_Oracle *OracleFilterer) WatchUpdatedRound(opts *bind.WatchOpts, sink chan<- *OracleUpdatedRound) (event.Subscription, error) {

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "UpdatedRound")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleUpdatedRound)
				if err := _Oracle.contract.UnpackLog(event, "UpdatedRound", log); err != nil {
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

// ParseUpdatedRound is a log parse operation binding the contract event 0x346e872acb7563c524fd9f17511221d966247f7c39abdc52fcafccc289f5c217.
//
// Solidity: event UpdatedRound(uint256 round)
func (_Oracle *OracleFilterer) ParseUpdatedRound(log types.Log) (*OracleUpdatedRound, error) {
	event := new(OracleUpdatedRound)
	if err := _Oracle.contract.UnpackLog(event, "UpdatedRound", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// OracleUpdatedSymbolsIterator is returned from FilterUpdatedSymbols and is used to iterate over the raw logs and unpacked data for UpdatedSymbols events raised by the Oracle contract.
type OracleUpdatedSymbolsIterator struct {
	Event *OracleUpdatedSymbols // Event containing the contract specifics and raw log

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
func (it *OracleUpdatedSymbolsIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(OracleUpdatedSymbols)
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
		it.Event = new(OracleUpdatedSymbols)
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
func (it *OracleUpdatedSymbolsIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *OracleUpdatedSymbolsIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// OracleUpdatedSymbols represents a UpdatedSymbols event raised by the Oracle contract.
type OracleUpdatedSymbols struct {
	Symbols []string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUpdatedSymbols is a free log retrieval operation binding the contract event 0xc527c8b5c6cf5a59f545e484a86b7569cd349c27526aa816b2ab4677d4b9a46d.
//
// Solidity: event UpdatedSymbols(string[] symbols)
func (_Oracle *OracleFilterer) FilterUpdatedSymbols(opts *bind.FilterOpts) (*OracleUpdatedSymbolsIterator, error) {

	logs, sub, err := _Oracle.contract.FilterLogs(opts, "UpdatedSymbols")
	if err != nil {
		return nil, err
	}
	return &OracleUpdatedSymbolsIterator{contract: _Oracle.contract, event: "UpdatedSymbols", logs: logs, sub: sub}, nil
}

// WatchUpdatedSymbols is a free log subscription operation binding the contract event 0xc527c8b5c6cf5a59f545e484a86b7569cd349c27526aa816b2ab4677d4b9a46d.
//
// Solidity: event UpdatedSymbols(string[] symbols)
func (_Oracle *OracleFilterer) WatchUpdatedSymbols(opts *bind.WatchOpts, sink chan<- *OracleUpdatedSymbols) (event.Subscription, error) {

	logs, sub, err := _Oracle.contract.WatchLogs(opts, "UpdatedSymbols")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(OracleUpdatedSymbols)
				if err := _Oracle.contract.UnpackLog(event, "UpdatedSymbols", log); err != nil {
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

// ParseUpdatedSymbols is a log parse operation binding the contract event 0xc527c8b5c6cf5a59f545e484a86b7569cd349c27526aa816b2ab4677d4b9a46d.
//
// Solidity: event UpdatedSymbols(string[] symbols)
func (_Oracle *OracleFilterer) ParseUpdatedSymbols(log types.Log) (*OracleUpdatedSymbols, error) {
	event := new(OracleUpdatedSymbols)
	if err := _Oracle.contract.UnpackLog(event, "UpdatedSymbols", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
