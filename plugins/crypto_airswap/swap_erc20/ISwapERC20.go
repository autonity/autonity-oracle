// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package swaperc20

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

// Swaperc20MetaData contains all meta data concerning the Swaperc20 contract.
var Swaperc20MetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signerWallet\",\"type\":\"address\"}],\"name\":\"Authorize\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signerWallet\",\"type\":\"address\"}],\"name\":\"Cancel\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signerWallet\",\"type\":\"address\"}],\"name\":\"Revoke\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"bonusMax\",\"type\":\"uint256\"}],\"name\":\"SetBonusMax\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"bonusScale\",\"type\":\"uint256\"}],\"name\":\"SetBonusScale\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"protocolFee\",\"type\":\"uint256\"}],\"name\":\"SetProtocolFee\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"protocolFeeLight\",\"type\":\"uint256\"}],\"name\":\"SetProtocolFeeLight\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"feeWallet\",\"type\":\"address\"}],\"name\":\"SetProtocolFeeWallet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"staking\",\"type\":\"address\"}],\"name\":\"SetStaking\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"signerWallet\",\"type\":\"address\"}],\"name\":\"SwapERC20\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"authorize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"authorized\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"calculateProtocolFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256[]\",\"name\":\"nonces\",\"type\":\"uint256[]\"}],\"name\":\"cancel\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"senderWallet\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expiry\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"signerWallet\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"signerToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"signerAmount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"senderToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"senderAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"check\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"nonceUsed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"revoke\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expiry\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"signerWallet\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"signerToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"signerAmount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"senderToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"senderAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"swap\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expiry\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"signerWallet\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"signerToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"signerAmount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"senderToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"senderAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"swapAnySender\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"expiry\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"signerWallet\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"signerToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"signerAmount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"senderToken\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"senderAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"swapLight\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// Swaperc20ABI is the input ABI used to generate the binding from.
// Deprecated: Use Swaperc20MetaData.ABI instead.
var Swaperc20ABI = Swaperc20MetaData.ABI

// Swaperc20 is an auto generated Go binding around an Ethereum contract.
type Swaperc20 struct {
	Swaperc20Caller     // Read-only binding to the contract
	Swaperc20Transactor // Write-only binding to the contract
	Swaperc20Filterer   // Log filterer for contract events
}

// Swaperc20Caller is an auto generated read-only Go binding around an Ethereum contract.
type Swaperc20Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Swaperc20Transactor is an auto generated write-only Go binding around an Ethereum contract.
type Swaperc20Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Swaperc20Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type Swaperc20Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// Swaperc20Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type Swaperc20Session struct {
	Contract     *Swaperc20        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// Swaperc20CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type Swaperc20CallerSession struct {
	Contract *Swaperc20Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// Swaperc20TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type Swaperc20TransactorSession struct {
	Contract     *Swaperc20Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// Swaperc20Raw is an auto generated low-level Go binding around an Ethereum contract.
type Swaperc20Raw struct {
	Contract *Swaperc20 // Generic contract binding to access the raw methods on
}

// Swaperc20CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type Swaperc20CallerRaw struct {
	Contract *Swaperc20Caller // Generic read-only contract binding to access the raw methods on
}

// Swaperc20TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type Swaperc20TransactorRaw struct {
	Contract *Swaperc20Transactor // Generic write-only contract binding to access the raw methods on
}

// NewSwaperc20 creates a new instance of Swaperc20, bound to a specific deployed contract.
func NewSwaperc20(address common.Address, backend bind.ContractBackend) (*Swaperc20, error) {
	contract, err := bindSwaperc20(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Swaperc20{Swaperc20Caller: Swaperc20Caller{contract: contract}, Swaperc20Transactor: Swaperc20Transactor{contract: contract}, Swaperc20Filterer: Swaperc20Filterer{contract: contract}}, nil
}

// NewSwaperc20Caller creates a new read-only instance of Swaperc20, bound to a specific deployed contract.
func NewSwaperc20Caller(address common.Address, caller bind.ContractCaller) (*Swaperc20Caller, error) {
	contract, err := bindSwaperc20(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &Swaperc20Caller{contract: contract}, nil
}

// NewSwaperc20Transactor creates a new write-only instance of Swaperc20, bound to a specific deployed contract.
func NewSwaperc20Transactor(address common.Address, transactor bind.ContractTransactor) (*Swaperc20Transactor, error) {
	contract, err := bindSwaperc20(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &Swaperc20Transactor{contract: contract}, nil
}

// NewSwaperc20Filterer creates a new log filterer instance of Swaperc20, bound to a specific deployed contract.
func NewSwaperc20Filterer(address common.Address, filterer bind.ContractFilterer) (*Swaperc20Filterer, error) {
	contract, err := bindSwaperc20(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &Swaperc20Filterer{contract: contract}, nil
}

// bindSwaperc20 binds a generic wrapper to an already deployed contract.
func bindSwaperc20(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(Swaperc20ABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Swaperc20 *Swaperc20Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Swaperc20.Contract.Swaperc20Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Swaperc20 *Swaperc20Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Swaperc20.Contract.Swaperc20Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Swaperc20 *Swaperc20Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Swaperc20.Contract.Swaperc20Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Swaperc20 *Swaperc20CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Swaperc20.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Swaperc20 *Swaperc20TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Swaperc20.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Swaperc20 *Swaperc20TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Swaperc20.Contract.contract.Transact(opts, method, params...)
}

// Authorized is a free data retrieval call binding the contract method 0xb9181611.
//
// Solidity: function authorized(address ) view returns(address)
func (_Swaperc20 *Swaperc20Caller) Authorized(opts *bind.CallOpts, arg0 common.Address) (common.Address, error) {
	var out []interface{}
	err := _Swaperc20.contract.Call(opts, &out, "authorized", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Authorized is a free data retrieval call binding the contract method 0xb9181611.
//
// Solidity: function authorized(address ) view returns(address)
func (_Swaperc20 *Swaperc20Session) Authorized(arg0 common.Address) (common.Address, error) {
	return _Swaperc20.Contract.Authorized(&_Swaperc20.CallOpts, arg0)
}

// Authorized is a free data retrieval call binding the contract method 0xb9181611.
//
// Solidity: function authorized(address ) view returns(address)
func (_Swaperc20 *Swaperc20CallerSession) Authorized(arg0 common.Address) (common.Address, error) {
	return _Swaperc20.Contract.Authorized(&_Swaperc20.CallOpts, arg0)
}

// ProtocolFeeWallet is a free data retrieval call binding the contract method 0xcbf7c6c3.
//
// Solidity: function protocolFeeWallet() view returns(address)
func (_Swaperc20 *Swaperc20Caller) ProtocolFeeWallet(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Swaperc20.contract.Call(opts, &out, "protocolFeeWallet")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ProtocolFeeWallet is a free data retrieval call binding the contract method 0xcbf7c6c3.
//
// Solidity: function protocolFeeWallet() view returns(address)
func (_Swaperc20 *Swaperc20Session) ProtocolFeeWallet() (common.Address, error) {
	return _Swaperc20.Contract.ProtocolFeeWallet(&_Swaperc20.CallOpts)
}

// ProtocolFeeWallet is a free data retrieval call binding the contract method 0xcbf7c6c3.
//
// Solidity: function protocolFeeWallet() view returns(address)
func (_Swaperc20 *Swaperc20CallerSession) ProtocolFeeWallet() (common.Address, error) {
	return _Swaperc20.Contract.ProtocolFeeWallet(&_Swaperc20.CallOpts)
}

// CalculateProtocolFee is a free data retrieval call binding the contract method 0x52c5f1f5.
//
// Solidity: function calculateProtocolFee(address , uint256 ) view returns(uint256)
func (_Swaperc20 *Swaperc20Caller) CalculateProtocolFee(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Swaperc20.contract.Call(opts, &out, "calculateProtocolFee", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateProtocolFee is a free data retrieval call binding the contract method 0x52c5f1f5.
//
// Solidity: function calculateProtocolFee(address , uint256 ) view returns(uint256)
func (_Swaperc20 *Swaperc20Session) CalculateProtocolFee(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Swaperc20.Contract.CalculateProtocolFee(&_Swaperc20.CallOpts, arg0, arg1)
}

// CalculateProtocolFee is a free data retrieval call binding the contract method 0x52c5f1f5.
//
// Solidity: function calculateProtocolFee(address , uint256 ) view returns(uint256)
func (_Swaperc20 *Swaperc20CallerSession) CalculateProtocolFee(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Swaperc20.Contract.CalculateProtocolFee(&_Swaperc20.CallOpts, arg0, arg1)
}

// Check is a free data retrieval call binding the contract method 0xb9cb01b0.
//
// Solidity: function check(address senderWallet, uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) view returns(bytes32[])
func (_Swaperc20 *Swaperc20Caller) Check(opts *bind.CallOpts, senderWallet common.Address, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) ([][32]byte, error) {
	var out []interface{}
	err := _Swaperc20.contract.Call(opts, &out, "check", senderWallet, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// Check is a free data retrieval call binding the contract method 0xb9cb01b0.
//
// Solidity: function check(address senderWallet, uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) view returns(bytes32[])
func (_Swaperc20 *Swaperc20Session) Check(senderWallet common.Address, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) ([][32]byte, error) {
	return _Swaperc20.Contract.Check(&_Swaperc20.CallOpts, senderWallet, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// Check is a free data retrieval call binding the contract method 0xb9cb01b0.
//
// Solidity: function check(address senderWallet, uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) view returns(bytes32[])
func (_Swaperc20 *Swaperc20CallerSession) Check(senderWallet common.Address, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) ([][32]byte, error) {
	return _Swaperc20.Contract.Check(&_Swaperc20.CallOpts, senderWallet, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// NonceUsed is a free data retrieval call binding the contract method 0x1647795e.
//
// Solidity: function nonceUsed(address , uint256 ) view returns(bool)
func (_Swaperc20 *Swaperc20Caller) NonceUsed(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (bool, error) {
	var out []interface{}
	err := _Swaperc20.contract.Call(opts, &out, "nonceUsed", arg0, arg1)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// NonceUsed is a free data retrieval call binding the contract method 0x1647795e.
//
// Solidity: function nonceUsed(address , uint256 ) view returns(bool)
func (_Swaperc20 *Swaperc20Session) NonceUsed(arg0 common.Address, arg1 *big.Int) (bool, error) {
	return _Swaperc20.Contract.NonceUsed(&_Swaperc20.CallOpts, arg0, arg1)
}

// NonceUsed is a free data retrieval call binding the contract method 0x1647795e.
//
// Solidity: function nonceUsed(address , uint256 ) view returns(bool)
func (_Swaperc20 *Swaperc20CallerSession) NonceUsed(arg0 common.Address, arg1 *big.Int) (bool, error) {
	return _Swaperc20.Contract.NonceUsed(&_Swaperc20.CallOpts, arg0, arg1)
}

// Authorize is a paid mutator transaction binding the contract method 0xb6a5d7de.
//
// Solidity: function authorize(address sender) returns()
func (_Swaperc20 *Swaperc20Transactor) Authorize(opts *bind.TransactOpts, sender common.Address) (*types.Transaction, error) {
	return _Swaperc20.contract.Transact(opts, "authorize", sender)
}

// Authorize is a paid mutator transaction binding the contract method 0xb6a5d7de.
//
// Solidity: function authorize(address sender) returns()
func (_Swaperc20 *Swaperc20Session) Authorize(sender common.Address) (*types.Transaction, error) {
	return _Swaperc20.Contract.Authorize(&_Swaperc20.TransactOpts, sender)
}

// Authorize is a paid mutator transaction binding the contract method 0xb6a5d7de.
//
// Solidity: function authorize(address sender) returns()
func (_Swaperc20 *Swaperc20TransactorSession) Authorize(sender common.Address) (*types.Transaction, error) {
	return _Swaperc20.Contract.Authorize(&_Swaperc20.TransactOpts, sender)
}

// Cancel is a paid mutator transaction binding the contract method 0x2e340823.
//
// Solidity: function cancel(uint256[] nonces) returns()
func (_Swaperc20 *Swaperc20Transactor) Cancel(opts *bind.TransactOpts, nonces []*big.Int) (*types.Transaction, error) {
	return _Swaperc20.contract.Transact(opts, "cancel", nonces)
}

// Cancel is a paid mutator transaction binding the contract method 0x2e340823.
//
// Solidity: function cancel(uint256[] nonces) returns()
func (_Swaperc20 *Swaperc20Session) Cancel(nonces []*big.Int) (*types.Transaction, error) {
	return _Swaperc20.Contract.Cancel(&_Swaperc20.TransactOpts, nonces)
}

// Cancel is a paid mutator transaction binding the contract method 0x2e340823.
//
// Solidity: function cancel(uint256[] nonces) returns()
func (_Swaperc20 *Swaperc20TransactorSession) Cancel(nonces []*big.Int) (*types.Transaction, error) {
	return _Swaperc20.Contract.Cancel(&_Swaperc20.TransactOpts, nonces)
}

// Revoke is a paid mutator transaction binding the contract method 0xb6549f75.
//
// Solidity: function revoke() returns()
func (_Swaperc20 *Swaperc20Transactor) Revoke(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Swaperc20.contract.Transact(opts, "revoke")
}

// Revoke is a paid mutator transaction binding the contract method 0xb6549f75.
//
// Solidity: function revoke() returns()
func (_Swaperc20 *Swaperc20Session) Revoke() (*types.Transaction, error) {
	return _Swaperc20.Contract.Revoke(&_Swaperc20.TransactOpts)
}

// Revoke is a paid mutator transaction binding the contract method 0xb6549f75.
//
// Solidity: function revoke() returns()
func (_Swaperc20 *Swaperc20TransactorSession) Revoke() (*types.Transaction, error) {
	return _Swaperc20.Contract.Revoke(&_Swaperc20.TransactOpts)
}

// Swap is a paid mutator transaction binding the contract method 0x98956069.
//
// Solidity: function swap(address recipient, uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Swaperc20 *Swaperc20Transactor) Swap(opts *bind.TransactOpts, recipient common.Address, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Swaperc20.contract.Transact(opts, "swap", recipient, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// Swap is a paid mutator transaction binding the contract method 0x98956069.
//
// Solidity: function swap(address recipient, uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Swaperc20 *Swaperc20Session) Swap(recipient common.Address, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Swaperc20.Contract.Swap(&_Swaperc20.TransactOpts, recipient, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// Swap is a paid mutator transaction binding the contract method 0x98956069.
//
// Solidity: function swap(address recipient, uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Swaperc20 *Swaperc20TransactorSession) Swap(recipient common.Address, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Swaperc20.Contract.Swap(&_Swaperc20.TransactOpts, recipient, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// SwapAnySender is a paid mutator transaction binding the contract method 0x3eb1af24.
//
// Solidity: function swapAnySender(address recipient, uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Swaperc20 *Swaperc20Transactor) SwapAnySender(opts *bind.TransactOpts, recipient common.Address, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Swaperc20.contract.Transact(opts, "swapAnySender", recipient, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// SwapAnySender is a paid mutator transaction binding the contract method 0x3eb1af24.
//
// Solidity: function swapAnySender(address recipient, uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Swaperc20 *Swaperc20Session) SwapAnySender(recipient common.Address, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Swaperc20.Contract.SwapAnySender(&_Swaperc20.TransactOpts, recipient, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// SwapAnySender is a paid mutator transaction binding the contract method 0x3eb1af24.
//
// Solidity: function swapAnySender(address recipient, uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Swaperc20 *Swaperc20TransactorSession) SwapAnySender(recipient common.Address, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Swaperc20.Contract.SwapAnySender(&_Swaperc20.TransactOpts, recipient, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// SwapLight is a paid mutator transaction binding the contract method 0x46e4480d.
//
// Solidity: function swapLight(uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Swaperc20 *Swaperc20Transactor) SwapLight(opts *bind.TransactOpts, nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Swaperc20.contract.Transact(opts, "swapLight", nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// SwapLight is a paid mutator transaction binding the contract method 0x46e4480d.
//
// Solidity: function swapLight(uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Swaperc20 *Swaperc20Session) SwapLight(nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Swaperc20.Contract.SwapLight(&_Swaperc20.TransactOpts, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// SwapLight is a paid mutator transaction binding the contract method 0x46e4480d.
//
// Solidity: function swapLight(uint256 nonce, uint256 expiry, address signerWallet, address signerToken, uint256 signerAmount, address senderToken, uint256 senderAmount, uint8 v, bytes32 r, bytes32 s) returns()
func (_Swaperc20 *Swaperc20TransactorSession) SwapLight(nonce *big.Int, expiry *big.Int, signerWallet common.Address, signerToken common.Address, signerAmount *big.Int, senderToken common.Address, senderAmount *big.Int, v uint8, r [32]byte, s [32]byte) (*types.Transaction, error) {
	return _Swaperc20.Contract.SwapLight(&_Swaperc20.TransactOpts, nonce, expiry, signerWallet, signerToken, signerAmount, senderToken, senderAmount, v, r, s)
}

// Swaperc20AuthorizeIterator is returned from FilterAuthorize and is used to iterate over the raw logs and unpacked data for Authorize events raised by the Swaperc20 contract.
type Swaperc20AuthorizeIterator struct {
	Event *Swaperc20Authorize // Event containing the contract specifics and raw log

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
func (it *Swaperc20AuthorizeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20Authorize)
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
		it.Event = new(Swaperc20Authorize)
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
func (it *Swaperc20AuthorizeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20AuthorizeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20Authorize represents a Authorize event raised by the Swaperc20 contract.
type Swaperc20Authorize struct {
	Signer       common.Address
	SignerWallet common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterAuthorize is a free log retrieval operation binding the contract event 0x30468de898bda644e26bab66e5a2241a3aa6aaf527257f5ca54e0f65204ba14a.
//
// Solidity: event Authorize(address indexed signer, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) FilterAuthorize(opts *bind.FilterOpts, signer []common.Address, signerWallet []common.Address) (*Swaperc20AuthorizeIterator, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}
	var signerWalletRule []interface{}
	for _, signerWalletItem := range signerWallet {
		signerWalletRule = append(signerWalletRule, signerWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "Authorize", signerRule, signerWalletRule)
	if err != nil {
		return nil, err
	}
	return &Swaperc20AuthorizeIterator{contract: _Swaperc20.contract, event: "Authorize", logs: logs, sub: sub}, nil
}

// WatchAuthorize is a free log subscription operation binding the contract event 0x30468de898bda644e26bab66e5a2241a3aa6aaf527257f5ca54e0f65204ba14a.
//
// Solidity: event Authorize(address indexed signer, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) WatchAuthorize(opts *bind.WatchOpts, sink chan<- *Swaperc20Authorize, signer []common.Address, signerWallet []common.Address) (event.Subscription, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}
	var signerWalletRule []interface{}
	for _, signerWalletItem := range signerWallet {
		signerWalletRule = append(signerWalletRule, signerWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "Authorize", signerRule, signerWalletRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20Authorize)
				if err := _Swaperc20.contract.UnpackLog(event, "Authorize", log); err != nil {
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

// ParseAuthorize is a log parse operation binding the contract event 0x30468de898bda644e26bab66e5a2241a3aa6aaf527257f5ca54e0f65204ba14a.
//
// Solidity: event Authorize(address indexed signer, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) ParseAuthorize(log types.Log) (*Swaperc20Authorize, error) {
	event := new(Swaperc20Authorize)
	if err := _Swaperc20.contract.UnpackLog(event, "Authorize", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Swaperc20CancelIterator is returned from FilterCancel and is used to iterate over the raw logs and unpacked data for Cancel events raised by the Swaperc20 contract.
type Swaperc20CancelIterator struct {
	Event *Swaperc20Cancel // Event containing the contract specifics and raw log

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
func (it *Swaperc20CancelIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20Cancel)
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
		it.Event = new(Swaperc20Cancel)
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
func (it *Swaperc20CancelIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20CancelIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20Cancel represents a Cancel event raised by the Swaperc20 contract.
type Swaperc20Cancel struct {
	Nonce        *big.Int
	SignerWallet common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterCancel is a free log retrieval operation binding the contract event 0x8dd3c361eb2366ff27c2db0eb07b9261f1d052570742ab8c9a0c326f37aa576d.
//
// Solidity: event Cancel(uint256 indexed nonce, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) FilterCancel(opts *bind.FilterOpts, nonce []*big.Int, signerWallet []common.Address) (*Swaperc20CancelIterator, error) {

	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}
	var signerWalletRule []interface{}
	for _, signerWalletItem := range signerWallet {
		signerWalletRule = append(signerWalletRule, signerWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "Cancel", nonceRule, signerWalletRule)
	if err != nil {
		return nil, err
	}
	return &Swaperc20CancelIterator{contract: _Swaperc20.contract, event: "Cancel", logs: logs, sub: sub}, nil
}

// WatchCancel is a free log subscription operation binding the contract event 0x8dd3c361eb2366ff27c2db0eb07b9261f1d052570742ab8c9a0c326f37aa576d.
//
// Solidity: event Cancel(uint256 indexed nonce, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) WatchCancel(opts *bind.WatchOpts, sink chan<- *Swaperc20Cancel, nonce []*big.Int, signerWallet []common.Address) (event.Subscription, error) {

	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}
	var signerWalletRule []interface{}
	for _, signerWalletItem := range signerWallet {
		signerWalletRule = append(signerWalletRule, signerWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "Cancel", nonceRule, signerWalletRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20Cancel)
				if err := _Swaperc20.contract.UnpackLog(event, "Cancel", log); err != nil {
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

// ParseCancel is a log parse operation binding the contract event 0x8dd3c361eb2366ff27c2db0eb07b9261f1d052570742ab8c9a0c326f37aa576d.
//
// Solidity: event Cancel(uint256 indexed nonce, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) ParseCancel(log types.Log) (*Swaperc20Cancel, error) {
	event := new(Swaperc20Cancel)
	if err := _Swaperc20.contract.UnpackLog(event, "Cancel", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Swaperc20RevokeIterator is returned from FilterRevoke and is used to iterate over the raw logs and unpacked data for Revoke events raised by the Swaperc20 contract.
type Swaperc20RevokeIterator struct {
	Event *Swaperc20Revoke // Event containing the contract specifics and raw log

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
func (it *Swaperc20RevokeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20Revoke)
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
		it.Event = new(Swaperc20Revoke)
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
func (it *Swaperc20RevokeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20RevokeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20Revoke represents a Revoke event raised by the Swaperc20 contract.
type Swaperc20Revoke struct {
	Signer       common.Address
	SignerWallet common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterRevoke is a free log retrieval operation binding the contract event 0xd7426110292f20fe59e73ccf52124e0f5440a756507c91c7b0a6c50e1eb1a23a.
//
// Solidity: event Revoke(address indexed signer, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) FilterRevoke(opts *bind.FilterOpts, signer []common.Address, signerWallet []common.Address) (*Swaperc20RevokeIterator, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}
	var signerWalletRule []interface{}
	for _, signerWalletItem := range signerWallet {
		signerWalletRule = append(signerWalletRule, signerWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "Revoke", signerRule, signerWalletRule)
	if err != nil {
		return nil, err
	}
	return &Swaperc20RevokeIterator{contract: _Swaperc20.contract, event: "Revoke", logs: logs, sub: sub}, nil
}

// WatchRevoke is a free log subscription operation binding the contract event 0xd7426110292f20fe59e73ccf52124e0f5440a756507c91c7b0a6c50e1eb1a23a.
//
// Solidity: event Revoke(address indexed signer, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) WatchRevoke(opts *bind.WatchOpts, sink chan<- *Swaperc20Revoke, signer []common.Address, signerWallet []common.Address) (event.Subscription, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}
	var signerWalletRule []interface{}
	for _, signerWalletItem := range signerWallet {
		signerWalletRule = append(signerWalletRule, signerWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "Revoke", signerRule, signerWalletRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20Revoke)
				if err := _Swaperc20.contract.UnpackLog(event, "Revoke", log); err != nil {
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

// ParseRevoke is a log parse operation binding the contract event 0xd7426110292f20fe59e73ccf52124e0f5440a756507c91c7b0a6c50e1eb1a23a.
//
// Solidity: event Revoke(address indexed signer, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) ParseRevoke(log types.Log) (*Swaperc20Revoke, error) {
	event := new(Swaperc20Revoke)
	if err := _Swaperc20.contract.UnpackLog(event, "Revoke", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Swaperc20SetBonusMaxIterator is returned from FilterSetBonusMax and is used to iterate over the raw logs and unpacked data for SetBonusMax events raised by the Swaperc20 contract.
type Swaperc20SetBonusMaxIterator struct {
	Event *Swaperc20SetBonusMax // Event containing the contract specifics and raw log

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
func (it *Swaperc20SetBonusMaxIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20SetBonusMax)
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
		it.Event = new(Swaperc20SetBonusMax)
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
func (it *Swaperc20SetBonusMaxIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20SetBonusMaxIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20SetBonusMax represents a SetBonusMax event raised by the Swaperc20 contract.
type Swaperc20SetBonusMax struct {
	BonusMax *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSetBonusMax is a free log retrieval operation binding the contract event 0xb113403a9e8b9f0173354acc3a5d210c86be40bb7259c19c55cea02227c5026f.
//
// Solidity: event SetBonusMax(uint256 bonusMax)
func (_Swaperc20 *Swaperc20Filterer) FilterSetBonusMax(opts *bind.FilterOpts) (*Swaperc20SetBonusMaxIterator, error) {

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "SetBonusMax")
	if err != nil {
		return nil, err
	}
	return &Swaperc20SetBonusMaxIterator{contract: _Swaperc20.contract, event: "SetBonusMax", logs: logs, sub: sub}, nil
}

// WatchSetBonusMax is a free log subscription operation binding the contract event 0xb113403a9e8b9f0173354acc3a5d210c86be40bb7259c19c55cea02227c5026f.
//
// Solidity: event SetBonusMax(uint256 bonusMax)
func (_Swaperc20 *Swaperc20Filterer) WatchSetBonusMax(opts *bind.WatchOpts, sink chan<- *Swaperc20SetBonusMax) (event.Subscription, error) {

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "SetBonusMax")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20SetBonusMax)
				if err := _Swaperc20.contract.UnpackLog(event, "SetBonusMax", log); err != nil {
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

// ParseSetBonusMax is a log parse operation binding the contract event 0xb113403a9e8b9f0173354acc3a5d210c86be40bb7259c19c55cea02227c5026f.
//
// Solidity: event SetBonusMax(uint256 bonusMax)
func (_Swaperc20 *Swaperc20Filterer) ParseSetBonusMax(log types.Log) (*Swaperc20SetBonusMax, error) {
	event := new(Swaperc20SetBonusMax)
	if err := _Swaperc20.contract.UnpackLog(event, "SetBonusMax", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Swaperc20SetBonusScaleIterator is returned from FilterSetBonusScale and is used to iterate over the raw logs and unpacked data for SetBonusScale events raised by the Swaperc20 contract.
type Swaperc20SetBonusScaleIterator struct {
	Event *Swaperc20SetBonusScale // Event containing the contract specifics and raw log

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
func (it *Swaperc20SetBonusScaleIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20SetBonusScale)
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
		it.Event = new(Swaperc20SetBonusScale)
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
func (it *Swaperc20SetBonusScaleIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20SetBonusScaleIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20SetBonusScale represents a SetBonusScale event raised by the Swaperc20 contract.
type Swaperc20SetBonusScale struct {
	BonusScale *big.Int
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSetBonusScale is a free log retrieval operation binding the contract event 0xcc5b12dfbda3644d5f3190b40ad8215d4aaac870df5c8112735085679d7cc333.
//
// Solidity: event SetBonusScale(uint256 bonusScale)
func (_Swaperc20 *Swaperc20Filterer) FilterSetBonusScale(opts *bind.FilterOpts) (*Swaperc20SetBonusScaleIterator, error) {

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "SetBonusScale")
	if err != nil {
		return nil, err
	}
	return &Swaperc20SetBonusScaleIterator{contract: _Swaperc20.contract, event: "SetBonusScale", logs: logs, sub: sub}, nil
}

// WatchSetBonusScale is a free log subscription operation binding the contract event 0xcc5b12dfbda3644d5f3190b40ad8215d4aaac870df5c8112735085679d7cc333.
//
// Solidity: event SetBonusScale(uint256 bonusScale)
func (_Swaperc20 *Swaperc20Filterer) WatchSetBonusScale(opts *bind.WatchOpts, sink chan<- *Swaperc20SetBonusScale) (event.Subscription, error) {

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "SetBonusScale")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20SetBonusScale)
				if err := _Swaperc20.contract.UnpackLog(event, "SetBonusScale", log); err != nil {
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

// ParseSetBonusScale is a log parse operation binding the contract event 0xcc5b12dfbda3644d5f3190b40ad8215d4aaac870df5c8112735085679d7cc333.
//
// Solidity: event SetBonusScale(uint256 bonusScale)
func (_Swaperc20 *Swaperc20Filterer) ParseSetBonusScale(log types.Log) (*Swaperc20SetBonusScale, error) {
	event := new(Swaperc20SetBonusScale)
	if err := _Swaperc20.contract.UnpackLog(event, "SetBonusScale", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Swaperc20SetProtocolFeeIterator is returned from FilterSetProtocolFee and is used to iterate over the raw logs and unpacked data for SetProtocolFee events raised by the Swaperc20 contract.
type Swaperc20SetProtocolFeeIterator struct {
	Event *Swaperc20SetProtocolFee // Event containing the contract specifics and raw log

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
func (it *Swaperc20SetProtocolFeeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20SetProtocolFee)
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
		it.Event = new(Swaperc20SetProtocolFee)
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
func (it *Swaperc20SetProtocolFeeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20SetProtocolFeeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20SetProtocolFee represents a SetProtocolFee event raised by the Swaperc20 contract.
type Swaperc20SetProtocolFee struct {
	ProtocolFee *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSetProtocolFee is a free log retrieval operation binding the contract event 0xdc0410a296e1e33943a772020d333d5f99319d7fcad932a484c53889f7aaa2b1.
//
// Solidity: event SetProtocolFee(uint256 protocolFee)
func (_Swaperc20 *Swaperc20Filterer) FilterSetProtocolFee(opts *bind.FilterOpts) (*Swaperc20SetProtocolFeeIterator, error) {

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "SetProtocolFee")
	if err != nil {
		return nil, err
	}
	return &Swaperc20SetProtocolFeeIterator{contract: _Swaperc20.contract, event: "SetProtocolFee", logs: logs, sub: sub}, nil
}

// WatchSetProtocolFee is a free log subscription operation binding the contract event 0xdc0410a296e1e33943a772020d333d5f99319d7fcad932a484c53889f7aaa2b1.
//
// Solidity: event SetProtocolFee(uint256 protocolFee)
func (_Swaperc20 *Swaperc20Filterer) WatchSetProtocolFee(opts *bind.WatchOpts, sink chan<- *Swaperc20SetProtocolFee) (event.Subscription, error) {

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "SetProtocolFee")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20SetProtocolFee)
				if err := _Swaperc20.contract.UnpackLog(event, "SetProtocolFee", log); err != nil {
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

// ParseSetProtocolFee is a log parse operation binding the contract event 0xdc0410a296e1e33943a772020d333d5f99319d7fcad932a484c53889f7aaa2b1.
//
// Solidity: event SetProtocolFee(uint256 protocolFee)
func (_Swaperc20 *Swaperc20Filterer) ParseSetProtocolFee(log types.Log) (*Swaperc20SetProtocolFee, error) {
	event := new(Swaperc20SetProtocolFee)
	if err := _Swaperc20.contract.UnpackLog(event, "SetProtocolFee", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Swaperc20SetProtocolFeeLightIterator is returned from FilterSetProtocolFeeLight and is used to iterate over the raw logs and unpacked data for SetProtocolFeeLight events raised by the Swaperc20 contract.
type Swaperc20SetProtocolFeeLightIterator struct {
	Event *Swaperc20SetProtocolFeeLight // Event containing the contract specifics and raw log

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
func (it *Swaperc20SetProtocolFeeLightIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20SetProtocolFeeLight)
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
		it.Event = new(Swaperc20SetProtocolFeeLight)
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
func (it *Swaperc20SetProtocolFeeLightIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20SetProtocolFeeLightIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20SetProtocolFeeLight represents a SetProtocolFeeLight event raised by the Swaperc20 contract.
type Swaperc20SetProtocolFeeLight struct {
	ProtocolFeeLight *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterSetProtocolFeeLight is a free log retrieval operation binding the contract event 0x312cc1a9b7287129a22395b9572a3c9ed09ce456f02b519efb34e12bb429eed0.
//
// Solidity: event SetProtocolFeeLight(uint256 protocolFeeLight)
func (_Swaperc20 *Swaperc20Filterer) FilterSetProtocolFeeLight(opts *bind.FilterOpts) (*Swaperc20SetProtocolFeeLightIterator, error) {

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "SetProtocolFeeLight")
	if err != nil {
		return nil, err
	}
	return &Swaperc20SetProtocolFeeLightIterator{contract: _Swaperc20.contract, event: "SetProtocolFeeLight", logs: logs, sub: sub}, nil
}

// WatchSetProtocolFeeLight is a free log subscription operation binding the contract event 0x312cc1a9b7287129a22395b9572a3c9ed09ce456f02b519efb34e12bb429eed0.
//
// Solidity: event SetProtocolFeeLight(uint256 protocolFeeLight)
func (_Swaperc20 *Swaperc20Filterer) WatchSetProtocolFeeLight(opts *bind.WatchOpts, sink chan<- *Swaperc20SetProtocolFeeLight) (event.Subscription, error) {

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "SetProtocolFeeLight")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20SetProtocolFeeLight)
				if err := _Swaperc20.contract.UnpackLog(event, "SetProtocolFeeLight", log); err != nil {
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

// ParseSetProtocolFeeLight is a log parse operation binding the contract event 0x312cc1a9b7287129a22395b9572a3c9ed09ce456f02b519efb34e12bb429eed0.
//
// Solidity: event SetProtocolFeeLight(uint256 protocolFeeLight)
func (_Swaperc20 *Swaperc20Filterer) ParseSetProtocolFeeLight(log types.Log) (*Swaperc20SetProtocolFeeLight, error) {
	event := new(Swaperc20SetProtocolFeeLight)
	if err := _Swaperc20.contract.UnpackLog(event, "SetProtocolFeeLight", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Swaperc20SetProtocolFeeWalletIterator is returned from FilterSetProtocolFeeWallet and is used to iterate over the raw logs and unpacked data for SetProtocolFeeWallet events raised by the Swaperc20 contract.
type Swaperc20SetProtocolFeeWalletIterator struct {
	Event *Swaperc20SetProtocolFeeWallet // Event containing the contract specifics and raw log

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
func (it *Swaperc20SetProtocolFeeWalletIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20SetProtocolFeeWallet)
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
		it.Event = new(Swaperc20SetProtocolFeeWallet)
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
func (it *Swaperc20SetProtocolFeeWalletIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20SetProtocolFeeWalletIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20SetProtocolFeeWallet represents a SetProtocolFeeWallet event raised by the Swaperc20 contract.
type Swaperc20SetProtocolFeeWallet struct {
	FeeWallet common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterSetProtocolFeeWallet is a free log retrieval operation binding the contract event 0x8b2a800ce9e2e7ccdf4741ae0e41b1f16983192291080ae3b78ac4296ddf598a.
//
// Solidity: event SetProtocolFeeWallet(address indexed feeWallet)
func (_Swaperc20 *Swaperc20Filterer) FilterSetProtocolFeeWallet(opts *bind.FilterOpts, feeWallet []common.Address) (*Swaperc20SetProtocolFeeWalletIterator, error) {

	var feeWalletRule []interface{}
	for _, feeWalletItem := range feeWallet {
		feeWalletRule = append(feeWalletRule, feeWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "SetProtocolFeeWallet", feeWalletRule)
	if err != nil {
		return nil, err
	}
	return &Swaperc20SetProtocolFeeWalletIterator{contract: _Swaperc20.contract, event: "SetProtocolFeeWallet", logs: logs, sub: sub}, nil
}

// WatchSetProtocolFeeWallet is a free log subscription operation binding the contract event 0x8b2a800ce9e2e7ccdf4741ae0e41b1f16983192291080ae3b78ac4296ddf598a.
//
// Solidity: event SetProtocolFeeWallet(address indexed feeWallet)
func (_Swaperc20 *Swaperc20Filterer) WatchSetProtocolFeeWallet(opts *bind.WatchOpts, sink chan<- *Swaperc20SetProtocolFeeWallet, feeWallet []common.Address) (event.Subscription, error) {

	var feeWalletRule []interface{}
	for _, feeWalletItem := range feeWallet {
		feeWalletRule = append(feeWalletRule, feeWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "SetProtocolFeeWallet", feeWalletRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20SetProtocolFeeWallet)
				if err := _Swaperc20.contract.UnpackLog(event, "SetProtocolFeeWallet", log); err != nil {
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

// ParseSetProtocolFeeWallet is a log parse operation binding the contract event 0x8b2a800ce9e2e7ccdf4741ae0e41b1f16983192291080ae3b78ac4296ddf598a.
//
// Solidity: event SetProtocolFeeWallet(address indexed feeWallet)
func (_Swaperc20 *Swaperc20Filterer) ParseSetProtocolFeeWallet(log types.Log) (*Swaperc20SetProtocolFeeWallet, error) {
	event := new(Swaperc20SetProtocolFeeWallet)
	if err := _Swaperc20.contract.UnpackLog(event, "SetProtocolFeeWallet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Swaperc20SetStakingIterator is returned from FilterSetStaking and is used to iterate over the raw logs and unpacked data for SetStaking events raised by the Swaperc20 contract.
type Swaperc20SetStakingIterator struct {
	Event *Swaperc20SetStaking // Event containing the contract specifics and raw log

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
func (it *Swaperc20SetStakingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20SetStaking)
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
		it.Event = new(Swaperc20SetStaking)
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
func (it *Swaperc20SetStakingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20SetStakingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20SetStaking represents a SetStaking event raised by the Swaperc20 contract.
type Swaperc20SetStaking struct {
	Staking common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterSetStaking is a free log retrieval operation binding the contract event 0x58fd5d9c33114e6edf8ea5d30956f8d1a4ab112b004f99928b4bcf1b87d66662.
//
// Solidity: event SetStaking(address indexed staking)
func (_Swaperc20 *Swaperc20Filterer) FilterSetStaking(opts *bind.FilterOpts, staking []common.Address) (*Swaperc20SetStakingIterator, error) {

	var stakingRule []interface{}
	for _, stakingItem := range staking {
		stakingRule = append(stakingRule, stakingItem)
	}

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "SetStaking", stakingRule)
	if err != nil {
		return nil, err
	}
	return &Swaperc20SetStakingIterator{contract: _Swaperc20.contract, event: "SetStaking", logs: logs, sub: sub}, nil
}

// WatchSetStaking is a free log subscription operation binding the contract event 0x58fd5d9c33114e6edf8ea5d30956f8d1a4ab112b004f99928b4bcf1b87d66662.
//
// Solidity: event SetStaking(address indexed staking)
func (_Swaperc20 *Swaperc20Filterer) WatchSetStaking(opts *bind.WatchOpts, sink chan<- *Swaperc20SetStaking, staking []common.Address) (event.Subscription, error) {

	var stakingRule []interface{}
	for _, stakingItem := range staking {
		stakingRule = append(stakingRule, stakingItem)
	}

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "SetStaking", stakingRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20SetStaking)
				if err := _Swaperc20.contract.UnpackLog(event, "SetStaking", log); err != nil {
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

// ParseSetStaking is a log parse operation binding the contract event 0x58fd5d9c33114e6edf8ea5d30956f8d1a4ab112b004f99928b4bcf1b87d66662.
//
// Solidity: event SetStaking(address indexed staking)
func (_Swaperc20 *Swaperc20Filterer) ParseSetStaking(log types.Log) (*Swaperc20SetStaking, error) {
	event := new(Swaperc20SetStaking)
	if err := _Swaperc20.contract.UnpackLog(event, "SetStaking", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// Swaperc20SwapERC20Iterator is returned from FilterSwapERC20 and is used to iterate over the raw logs and unpacked data for SwapERC20 events raised by the Swaperc20 contract.
type Swaperc20SwapERC20Iterator struct {
	Event *Swaperc20SwapERC20 // Event containing the contract specifics and raw log

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
func (it *Swaperc20SwapERC20Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(Swaperc20SwapERC20)
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
		it.Event = new(Swaperc20SwapERC20)
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
func (it *Swaperc20SwapERC20Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *Swaperc20SwapERC20Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// Swaperc20SwapERC20 represents a SwapERC20 event raised by the Swaperc20 contract.
type Swaperc20SwapERC20 struct {
	Nonce        *big.Int
	SignerWallet common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterSwapERC20 is a free log retrieval operation binding the contract event 0x4294f3cfba9ff22cfa9cb602947f7656aa160c0a6c8fa406a28e12bed6bf2093.
//
// Solidity: event SwapERC20(uint256 indexed nonce, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) FilterSwapERC20(opts *bind.FilterOpts, nonce []*big.Int, signerWallet []common.Address) (*Swaperc20SwapERC20Iterator, error) {

	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}
	var signerWalletRule []interface{}
	for _, signerWalletItem := range signerWallet {
		signerWalletRule = append(signerWalletRule, signerWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.FilterLogs(opts, "SwapERC20", nonceRule, signerWalletRule)
	if err != nil {
		return nil, err
	}
	return &Swaperc20SwapERC20Iterator{contract: _Swaperc20.contract, event: "SwapERC20", logs: logs, sub: sub}, nil
}

// WatchSwapERC20 is a free log subscription operation binding the contract event 0x4294f3cfba9ff22cfa9cb602947f7656aa160c0a6c8fa406a28e12bed6bf2093.
//
// Solidity: event SwapERC20(uint256 indexed nonce, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) WatchSwapERC20(opts *bind.WatchOpts, sink chan<- *Swaperc20SwapERC20, nonce []*big.Int, signerWallet []common.Address) (event.Subscription, error) {

	var nonceRule []interface{}
	for _, nonceItem := range nonce {
		nonceRule = append(nonceRule, nonceItem)
	}
	var signerWalletRule []interface{}
	for _, signerWalletItem := range signerWallet {
		signerWalletRule = append(signerWalletRule, signerWalletItem)
	}

	logs, sub, err := _Swaperc20.contract.WatchLogs(opts, "SwapERC20", nonceRule, signerWalletRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(Swaperc20SwapERC20)
				if err := _Swaperc20.contract.UnpackLog(event, "SwapERC20", log); err != nil {
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

// ParseSwapERC20 is a log parse operation binding the contract event 0x4294f3cfba9ff22cfa9cb602947f7656aa160c0a6c8fa406a28e12bed6bf2093.
//
// Solidity: event SwapERC20(uint256 indexed nonce, address indexed signerWallet)
func (_Swaperc20 *Swaperc20Filterer) ParseSwapERC20(log types.Log) (*Swaperc20SwapERC20, error) {
	event := new(Swaperc20SwapERC20)
	if err := _Swaperc20.contract.UnpackLog(event, "SwapERC20", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
