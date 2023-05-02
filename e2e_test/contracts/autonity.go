// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package autonity

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

// IAutonityCommitteeMember is an auto generated low-level Go binding around an user-defined struct.
type IAutonityCommitteeMember struct {
	Addr        common.Address
	VotingPower *big.Int
}

// AutonityMetaData contains all meta data concerning the Autonity contract.
var AutonityMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"getCommittee\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"votingPower\",\"type\":\"uint256\"}],\"internalType\":\"structIAutonity.CommitteeMember[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_size\",\"type\":\"uint256\"}],\"name\":\"setCommitteeSize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// AutonityABI is the input ABI used to generate the binding from.
// Deprecated: Use AutonityMetaData.ABI instead.
var AutonityABI = AutonityMetaData.ABI

// Autonity is an auto generated Go binding around an Ethereum contract.
type Autonity struct {
	AutonityCaller     // Read-only binding to the contract
	AutonityTransactor // Write-only binding to the contract
	AutonityFilterer   // Log filterer for contract events
}

// AutonityCaller is an auto generated read-only Go binding around an Ethereum contract.
type AutonityCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AutonityTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AutonityTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AutonityFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AutonityFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AutonitySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AutonitySession struct {
	Contract     *Autonity         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AutonityCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AutonityCallerSession struct {
	Contract *AutonityCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// AutonityTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AutonityTransactorSession struct {
	Contract     *AutonityTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// AutonityRaw is an auto generated low-level Go binding around an Ethereum contract.
type AutonityRaw struct {
	Contract *Autonity // Generic contract binding to access the raw methods on
}

// AutonityCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AutonityCallerRaw struct {
	Contract *AutonityCaller // Generic read-only contract binding to access the raw methods on
}

// AutonityTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AutonityTransactorRaw struct {
	Contract *AutonityTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAutonity creates a new instance of Autonity, bound to a specific deployed contract.
func NewAutonity(address common.Address, backend bind.ContractBackend) (*Autonity, error) {
	contract, err := bindAutonity(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Autonity{AutonityCaller: AutonityCaller{contract: contract}, AutonityTransactor: AutonityTransactor{contract: contract}, AutonityFilterer: AutonityFilterer{contract: contract}}, nil
}

// NewAutonityCaller creates a new read-only instance of Autonity, bound to a specific deployed contract.
func NewAutonityCaller(address common.Address, caller bind.ContractCaller) (*AutonityCaller, error) {
	contract, err := bindAutonity(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AutonityCaller{contract: contract}, nil
}

// NewAutonityTransactor creates a new write-only instance of Autonity, bound to a specific deployed contract.
func NewAutonityTransactor(address common.Address, transactor bind.ContractTransactor) (*AutonityTransactor, error) {
	contract, err := bindAutonity(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AutonityTransactor{contract: contract}, nil
}

// NewAutonityFilterer creates a new log filterer instance of Autonity, bound to a specific deployed contract.
func NewAutonityFilterer(address common.Address, filterer bind.ContractFilterer) (*AutonityFilterer, error) {
	contract, err := bindAutonity(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AutonityFilterer{contract: contract}, nil
}

// bindAutonity binds a generic wrapper to an already deployed contract.
func bindAutonity(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(AutonityABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Autonity *AutonityRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Autonity.Contract.AutonityCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Autonity *AutonityRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Autonity.Contract.AutonityTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Autonity *AutonityRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Autonity.Contract.AutonityTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Autonity *AutonityCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Autonity.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Autonity *AutonityTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Autonity.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Autonity *AutonityTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Autonity.Contract.contract.Transact(opts, method, params...)
}

// GetCommittee is a free data retrieval call binding the contract method 0xab8f6ffe.
//
// Solidity: function getCommittee() view returns((address,uint256)[])
func (_Autonity *AutonityCaller) GetCommittee(opts *bind.CallOpts) ([]IAutonityCommitteeMember, error) {
	var out []interface{}
	err := _Autonity.contract.Call(opts, &out, "getCommittee")

	if err != nil {
		return *new([]IAutonityCommitteeMember), err
	}

	out0 := *abi.ConvertType(out[0], new([]IAutonityCommitteeMember)).(*[]IAutonityCommitteeMember)

	return out0, err

}

// GetCommittee is a free data retrieval call binding the contract method 0xab8f6ffe.
//
// Solidity: function getCommittee() view returns((address,uint256)[])
func (_Autonity *AutonitySession) GetCommittee() ([]IAutonityCommitteeMember, error) {
	return _Autonity.Contract.GetCommittee(&_Autonity.CallOpts)
}

// GetCommittee is a free data retrieval call binding the contract method 0xab8f6ffe.
//
// Solidity: function getCommittee() view returns((address,uint256)[])
func (_Autonity *AutonityCallerSession) GetCommittee() ([]IAutonityCommitteeMember, error) {
	return _Autonity.Contract.GetCommittee(&_Autonity.CallOpts)
}

// SetCommitteeSize is a paid mutator transaction binding the contract method 0x8bac7dad.
//
// Solidity: function setCommitteeSize(uint256 _size) returns()
func (_Autonity *AutonityTransactor) SetCommitteeSize(opts *bind.TransactOpts, _size *big.Int) (*types.Transaction, error) {
	return _Autonity.contract.Transact(opts, "setCommitteeSize", _size)
}

// SetCommitteeSize is a paid mutator transaction binding the contract method 0x8bac7dad.
//
// Solidity: function setCommitteeSize(uint256 _size) returns()
func (_Autonity *AutonitySession) SetCommitteeSize(_size *big.Int) (*types.Transaction, error) {
	return _Autonity.Contract.SetCommitteeSize(&_Autonity.TransactOpts, _size)
}

// SetCommitteeSize is a paid mutator transaction binding the contract method 0x8bac7dad.
//
// Solidity: function setCommitteeSize(uint256 _size) returns()
func (_Autonity *AutonityTransactorSession) SetCommitteeSize(_size *big.Int) (*types.Transaction, error) {
	return _Autonity.Contract.SetCommitteeSize(&_Autonity.TransactOpts, _size)
}
