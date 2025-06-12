// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

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
	_ = abi.ConvertType
)

// PoolAddressesProviderMetaData contains all meta data concerning the PoolAddressesProvider contract.
var PoolAddressesProviderMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"marketId\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getACLAdmin\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getACLManager\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getAddress\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getMarketId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPool\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPoolConfigurator\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPoolDataProvider\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPriceOracle\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getPriceOracleSentinel\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setACLAdmin\",\"inputs\":[{\"name\":\"newAclAdmin\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setACLManager\",\"inputs\":[{\"name\":\"newAclManager\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAddress\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"newAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAddressAsProxy\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"newImplementationAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMarketId\",\"inputs\":[{\"name\":\"newMarketId\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setPoolConfiguratorImpl\",\"inputs\":[{\"name\":\"newPoolConfiguratorImpl\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setPoolDataProvider\",\"inputs\":[{\"name\":\"newDataProvider\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setPoolImpl\",\"inputs\":[{\"name\":\"newPoolImpl\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setPriceOracle\",\"inputs\":[{\"name\":\"newPriceOracle\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setPriceOracleSentinel\",\"inputs\":[{\"name\":\"newPriceOracleSentinel\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"ACLAdminUpdated\",\"inputs\":[{\"name\":\"oldAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ACLManagerUpdated\",\"inputs\":[{\"name\":\"oldAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AddressSet\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"oldAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"AddressSetAsProxy\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"proxyAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"oldImplementationAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"newImplementationAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MarketIdSet\",\"inputs\":[{\"name\":\"oldMarketId\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"},{\"name\":\"newMarketId\",\"type\":\"string\",\"indexed\":true,\"internalType\":\"string\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PoolConfiguratorUpdated\",\"inputs\":[{\"name\":\"oldAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PoolDataProviderUpdated\",\"inputs\":[{\"name\":\"oldAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PoolUpdated\",\"inputs\":[{\"name\":\"oldAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PriceOracleSentinelUpdated\",\"inputs\":[{\"name\":\"oldAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"PriceOracleUpdated\",\"inputs\":[{\"name\":\"oldAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ProxyCreated\",\"inputs\":[{\"name\":\"id\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"proxyAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"implementationAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false}]",
}

// PoolAddressesProviderABI is the input ABI used to generate the binding from.
// Deprecated: Use PoolAddressesProviderMetaData.ABI instead.
var PoolAddressesProviderABI = PoolAddressesProviderMetaData.ABI

// PoolAddressesProvider is an auto generated Go binding around an Ethereum contract.
type PoolAddressesProvider struct {
	PoolAddressesProviderCaller     // Read-only binding to the contract
	PoolAddressesProviderTransactor // Write-only binding to the contract
	PoolAddressesProviderFilterer   // Log filterer for contract events
}

// PoolAddressesProviderCaller is an auto generated read-only Go binding around an Ethereum contract.
type PoolAddressesProviderCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PoolAddressesProviderTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PoolAddressesProviderTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PoolAddressesProviderFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PoolAddressesProviderFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PoolAddressesProviderSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PoolAddressesProviderSession struct {
	Contract     *PoolAddressesProvider // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// PoolAddressesProviderCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PoolAddressesProviderCallerSession struct {
	Contract *PoolAddressesProviderCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// PoolAddressesProviderTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PoolAddressesProviderTransactorSession struct {
	Contract     *PoolAddressesProviderTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// PoolAddressesProviderRaw is an auto generated low-level Go binding around an Ethereum contract.
type PoolAddressesProviderRaw struct {
	Contract *PoolAddressesProvider // Generic contract binding to access the raw methods on
}

// PoolAddressesProviderCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PoolAddressesProviderCallerRaw struct {
	Contract *PoolAddressesProviderCaller // Generic read-only contract binding to access the raw methods on
}

// PoolAddressesProviderTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PoolAddressesProviderTransactorRaw struct {
	Contract *PoolAddressesProviderTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPoolAddressesProvider creates a new instance of PoolAddressesProvider, bound to a specific deployed contract.
func NewPoolAddressesProvider(address common.Address, backend bind.ContractBackend) (*PoolAddressesProvider, error) {
	contract, err := bindPoolAddressesProvider(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProvider{PoolAddressesProviderCaller: PoolAddressesProviderCaller{contract: contract}, PoolAddressesProviderTransactor: PoolAddressesProviderTransactor{contract: contract}, PoolAddressesProviderFilterer: PoolAddressesProviderFilterer{contract: contract}}, nil
}

// NewPoolAddressesProviderCaller creates a new read-only instance of PoolAddressesProvider, bound to a specific deployed contract.
func NewPoolAddressesProviderCaller(address common.Address, caller bind.ContractCaller) (*PoolAddressesProviderCaller, error) {
	contract, err := bindPoolAddressesProvider(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderCaller{contract: contract}, nil
}

// NewPoolAddressesProviderTransactor creates a new write-only instance of PoolAddressesProvider, bound to a specific deployed contract.
func NewPoolAddressesProviderTransactor(address common.Address, transactor bind.ContractTransactor) (*PoolAddressesProviderTransactor, error) {
	contract, err := bindPoolAddressesProvider(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderTransactor{contract: contract}, nil
}

// NewPoolAddressesProviderFilterer creates a new log filterer instance of PoolAddressesProvider, bound to a specific deployed contract.
func NewPoolAddressesProviderFilterer(address common.Address, filterer bind.ContractFilterer) (*PoolAddressesProviderFilterer, error) {
	contract, err := bindPoolAddressesProvider(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderFilterer{contract: contract}, nil
}

// bindPoolAddressesProvider binds a generic wrapper to an already deployed contract.
func bindPoolAddressesProvider(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PoolAddressesProviderMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PoolAddressesProvider *PoolAddressesProviderRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PoolAddressesProvider.Contract.PoolAddressesProviderCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PoolAddressesProvider *PoolAddressesProviderRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.PoolAddressesProviderTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PoolAddressesProvider *PoolAddressesProviderRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.PoolAddressesProviderTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PoolAddressesProvider *PoolAddressesProviderCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PoolAddressesProvider.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PoolAddressesProvider *PoolAddressesProviderTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PoolAddressesProvider *PoolAddressesProviderTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.contract.Transact(opts, method, params...)
}

// GetACLAdmin is a free data retrieval call binding the contract method 0x0e67178c.
//
// Solidity: function getACLAdmin() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) GetACLAdmin(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "getACLAdmin")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetACLAdmin is a free data retrieval call binding the contract method 0x0e67178c.
//
// Solidity: function getACLAdmin() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderSession) GetACLAdmin() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetACLAdmin(&_PoolAddressesProvider.CallOpts)
}

// GetACLAdmin is a free data retrieval call binding the contract method 0x0e67178c.
//
// Solidity: function getACLAdmin() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) GetACLAdmin() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetACLAdmin(&_PoolAddressesProvider.CallOpts)
}

// GetACLManager is a free data retrieval call binding the contract method 0x707cd716.
//
// Solidity: function getACLManager() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) GetACLManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "getACLManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetACLManager is a free data retrieval call binding the contract method 0x707cd716.
//
// Solidity: function getACLManager() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderSession) GetACLManager() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetACLManager(&_PoolAddressesProvider.CallOpts)
}

// GetACLManager is a free data retrieval call binding the contract method 0x707cd716.
//
// Solidity: function getACLManager() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) GetACLManager() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetACLManager(&_PoolAddressesProvider.CallOpts)
}

// GetAddress is a free data retrieval call binding the contract method 0x21f8a721.
//
// Solidity: function getAddress(bytes32 id) view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) GetAddress(opts *bind.CallOpts, id [32]byte) (common.Address, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "getAddress", id)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetAddress is a free data retrieval call binding the contract method 0x21f8a721.
//
// Solidity: function getAddress(bytes32 id) view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderSession) GetAddress(id [32]byte) (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetAddress(&_PoolAddressesProvider.CallOpts, id)
}

// GetAddress is a free data retrieval call binding the contract method 0x21f8a721.
//
// Solidity: function getAddress(bytes32 id) view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) GetAddress(id [32]byte) (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetAddress(&_PoolAddressesProvider.CallOpts, id)
}

// GetMarketId is a free data retrieval call binding the contract method 0x568ef470.
//
// Solidity: function getMarketId() view returns(string)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) GetMarketId(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "getMarketId")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// GetMarketId is a free data retrieval call binding the contract method 0x568ef470.
//
// Solidity: function getMarketId() view returns(string)
func (_PoolAddressesProvider *PoolAddressesProviderSession) GetMarketId() (string, error) {
	return _PoolAddressesProvider.Contract.GetMarketId(&_PoolAddressesProvider.CallOpts)
}

// GetMarketId is a free data retrieval call binding the contract method 0x568ef470.
//
// Solidity: function getMarketId() view returns(string)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) GetMarketId() (string, error) {
	return _PoolAddressesProvider.Contract.GetMarketId(&_PoolAddressesProvider.CallOpts)
}

// GetPool is a free data retrieval call binding the contract method 0x026b1d5f.
//
// Solidity: function getPool() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) GetPool(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "getPool")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPool is a free data retrieval call binding the contract method 0x026b1d5f.
//
// Solidity: function getPool() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderSession) GetPool() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPool(&_PoolAddressesProvider.CallOpts)
}

// GetPool is a free data retrieval call binding the contract method 0x026b1d5f.
//
// Solidity: function getPool() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) GetPool() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPool(&_PoolAddressesProvider.CallOpts)
}

// GetPoolConfigurator is a free data retrieval call binding the contract method 0x631adfca.
//
// Solidity: function getPoolConfigurator() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) GetPoolConfigurator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "getPoolConfigurator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPoolConfigurator is a free data retrieval call binding the contract method 0x631adfca.
//
// Solidity: function getPoolConfigurator() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderSession) GetPoolConfigurator() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPoolConfigurator(&_PoolAddressesProvider.CallOpts)
}

// GetPoolConfigurator is a free data retrieval call binding the contract method 0x631adfca.
//
// Solidity: function getPoolConfigurator() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) GetPoolConfigurator() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPoolConfigurator(&_PoolAddressesProvider.CallOpts)
}

// GetPoolDataProvider is a free data retrieval call binding the contract method 0xe860accb.
//
// Solidity: function getPoolDataProvider() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) GetPoolDataProvider(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "getPoolDataProvider")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPoolDataProvider is a free data retrieval call binding the contract method 0xe860accb.
//
// Solidity: function getPoolDataProvider() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderSession) GetPoolDataProvider() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPoolDataProvider(&_PoolAddressesProvider.CallOpts)
}

// GetPoolDataProvider is a free data retrieval call binding the contract method 0xe860accb.
//
// Solidity: function getPoolDataProvider() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) GetPoolDataProvider() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPoolDataProvider(&_PoolAddressesProvider.CallOpts)
}

// GetPriceOracle is a free data retrieval call binding the contract method 0xfca513a8.
//
// Solidity: function getPriceOracle() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) GetPriceOracle(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "getPriceOracle")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPriceOracle is a free data retrieval call binding the contract method 0xfca513a8.
//
// Solidity: function getPriceOracle() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderSession) GetPriceOracle() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPriceOracle(&_PoolAddressesProvider.CallOpts)
}

// GetPriceOracle is a free data retrieval call binding the contract method 0xfca513a8.
//
// Solidity: function getPriceOracle() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) GetPriceOracle() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPriceOracle(&_PoolAddressesProvider.CallOpts)
}

// GetPriceOracleSentinel is a free data retrieval call binding the contract method 0x5eb88d3d.
//
// Solidity: function getPriceOracleSentinel() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) GetPriceOracleSentinel(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "getPriceOracleSentinel")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPriceOracleSentinel is a free data retrieval call binding the contract method 0x5eb88d3d.
//
// Solidity: function getPriceOracleSentinel() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderSession) GetPriceOracleSentinel() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPriceOracleSentinel(&_PoolAddressesProvider.CallOpts)
}

// GetPriceOracleSentinel is a free data retrieval call binding the contract method 0x5eb88d3d.
//
// Solidity: function getPriceOracleSentinel() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) GetPriceOracleSentinel() (common.Address, error) {
	return _PoolAddressesProvider.Contract.GetPriceOracleSentinel(&_PoolAddressesProvider.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _PoolAddressesProvider.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderSession) Owner() (common.Address, error) {
	return _PoolAddressesProvider.Contract.Owner(&_PoolAddressesProvider.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_PoolAddressesProvider *PoolAddressesProviderCallerSession) Owner() (common.Address, error) {
	return _PoolAddressesProvider.Contract.Owner(&_PoolAddressesProvider.CallOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) RenounceOwnership() (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.RenounceOwnership(&_PoolAddressesProvider.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.RenounceOwnership(&_PoolAddressesProvider.TransactOpts)
}

// SetACLAdmin is a paid mutator transaction binding the contract method 0x76d84ffc.
//
// Solidity: function setACLAdmin(address newAclAdmin) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetACLAdmin(opts *bind.TransactOpts, newAclAdmin common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setACLAdmin", newAclAdmin)
}

// SetACLAdmin is a paid mutator transaction binding the contract method 0x76d84ffc.
//
// Solidity: function setACLAdmin(address newAclAdmin) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetACLAdmin(newAclAdmin common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetACLAdmin(&_PoolAddressesProvider.TransactOpts, newAclAdmin)
}

// SetACLAdmin is a paid mutator transaction binding the contract method 0x76d84ffc.
//
// Solidity: function setACLAdmin(address newAclAdmin) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetACLAdmin(newAclAdmin common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetACLAdmin(&_PoolAddressesProvider.TransactOpts, newAclAdmin)
}

// SetACLManager is a paid mutator transaction binding the contract method 0xed301ca9.
//
// Solidity: function setACLManager(address newAclManager) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetACLManager(opts *bind.TransactOpts, newAclManager common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setACLManager", newAclManager)
}

// SetACLManager is a paid mutator transaction binding the contract method 0xed301ca9.
//
// Solidity: function setACLManager(address newAclManager) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetACLManager(newAclManager common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetACLManager(&_PoolAddressesProvider.TransactOpts, newAclManager)
}

// SetACLManager is a paid mutator transaction binding the contract method 0xed301ca9.
//
// Solidity: function setACLManager(address newAclManager) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetACLManager(newAclManager common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetACLManager(&_PoolAddressesProvider.TransactOpts, newAclManager)
}

// SetAddress is a paid mutator transaction binding the contract method 0xca446dd9.
//
// Solidity: function setAddress(bytes32 id, address newAddress) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetAddress(opts *bind.TransactOpts, id [32]byte, newAddress common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setAddress", id, newAddress)
}

// SetAddress is a paid mutator transaction binding the contract method 0xca446dd9.
//
// Solidity: function setAddress(bytes32 id, address newAddress) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetAddress(id [32]byte, newAddress common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetAddress(&_PoolAddressesProvider.TransactOpts, id, newAddress)
}

// SetAddress is a paid mutator transaction binding the contract method 0xca446dd9.
//
// Solidity: function setAddress(bytes32 id, address newAddress) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetAddress(id [32]byte, newAddress common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetAddress(&_PoolAddressesProvider.TransactOpts, id, newAddress)
}

// SetAddressAsProxy is a paid mutator transaction binding the contract method 0x5dcc528c.
//
// Solidity: function setAddressAsProxy(bytes32 id, address newImplementationAddress) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetAddressAsProxy(opts *bind.TransactOpts, id [32]byte, newImplementationAddress common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setAddressAsProxy", id, newImplementationAddress)
}

// SetAddressAsProxy is a paid mutator transaction binding the contract method 0x5dcc528c.
//
// Solidity: function setAddressAsProxy(bytes32 id, address newImplementationAddress) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetAddressAsProxy(id [32]byte, newImplementationAddress common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetAddressAsProxy(&_PoolAddressesProvider.TransactOpts, id, newImplementationAddress)
}

// SetAddressAsProxy is a paid mutator transaction binding the contract method 0x5dcc528c.
//
// Solidity: function setAddressAsProxy(bytes32 id, address newImplementationAddress) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetAddressAsProxy(id [32]byte, newImplementationAddress common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetAddressAsProxy(&_PoolAddressesProvider.TransactOpts, id, newImplementationAddress)
}

// SetMarketId is a paid mutator transaction binding the contract method 0xf67b1847.
//
// Solidity: function setMarketId(string newMarketId) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetMarketId(opts *bind.TransactOpts, newMarketId string) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setMarketId", newMarketId)
}

// SetMarketId is a paid mutator transaction binding the contract method 0xf67b1847.
//
// Solidity: function setMarketId(string newMarketId) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetMarketId(newMarketId string) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetMarketId(&_PoolAddressesProvider.TransactOpts, newMarketId)
}

// SetMarketId is a paid mutator transaction binding the contract method 0xf67b1847.
//
// Solidity: function setMarketId(string newMarketId) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetMarketId(newMarketId string) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetMarketId(&_PoolAddressesProvider.TransactOpts, newMarketId)
}

// SetPoolConfiguratorImpl is a paid mutator transaction binding the contract method 0xe4ca28b7.
//
// Solidity: function setPoolConfiguratorImpl(address newPoolConfiguratorImpl) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetPoolConfiguratorImpl(opts *bind.TransactOpts, newPoolConfiguratorImpl common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setPoolConfiguratorImpl", newPoolConfiguratorImpl)
}

// SetPoolConfiguratorImpl is a paid mutator transaction binding the contract method 0xe4ca28b7.
//
// Solidity: function setPoolConfiguratorImpl(address newPoolConfiguratorImpl) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetPoolConfiguratorImpl(newPoolConfiguratorImpl common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPoolConfiguratorImpl(&_PoolAddressesProvider.TransactOpts, newPoolConfiguratorImpl)
}

// SetPoolConfiguratorImpl is a paid mutator transaction binding the contract method 0xe4ca28b7.
//
// Solidity: function setPoolConfiguratorImpl(address newPoolConfiguratorImpl) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetPoolConfiguratorImpl(newPoolConfiguratorImpl common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPoolConfiguratorImpl(&_PoolAddressesProvider.TransactOpts, newPoolConfiguratorImpl)
}

// SetPoolDataProvider is a paid mutator transaction binding the contract method 0xe44e9ed1.
//
// Solidity: function setPoolDataProvider(address newDataProvider) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetPoolDataProvider(opts *bind.TransactOpts, newDataProvider common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setPoolDataProvider", newDataProvider)
}

// SetPoolDataProvider is a paid mutator transaction binding the contract method 0xe44e9ed1.
//
// Solidity: function setPoolDataProvider(address newDataProvider) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetPoolDataProvider(newDataProvider common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPoolDataProvider(&_PoolAddressesProvider.TransactOpts, newDataProvider)
}

// SetPoolDataProvider is a paid mutator transaction binding the contract method 0xe44e9ed1.
//
// Solidity: function setPoolDataProvider(address newDataProvider) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetPoolDataProvider(newDataProvider common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPoolDataProvider(&_PoolAddressesProvider.TransactOpts, newDataProvider)
}

// SetPoolImpl is a paid mutator transaction binding the contract method 0xa1564406.
//
// Solidity: function setPoolImpl(address newPoolImpl) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetPoolImpl(opts *bind.TransactOpts, newPoolImpl common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setPoolImpl", newPoolImpl)
}

// SetPoolImpl is a paid mutator transaction binding the contract method 0xa1564406.
//
// Solidity: function setPoolImpl(address newPoolImpl) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetPoolImpl(newPoolImpl common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPoolImpl(&_PoolAddressesProvider.TransactOpts, newPoolImpl)
}

// SetPoolImpl is a paid mutator transaction binding the contract method 0xa1564406.
//
// Solidity: function setPoolImpl(address newPoolImpl) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetPoolImpl(newPoolImpl common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPoolImpl(&_PoolAddressesProvider.TransactOpts, newPoolImpl)
}

// SetPriceOracle is a paid mutator transaction binding the contract method 0x530e784f.
//
// Solidity: function setPriceOracle(address newPriceOracle) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetPriceOracle(opts *bind.TransactOpts, newPriceOracle common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setPriceOracle", newPriceOracle)
}

// SetPriceOracle is a paid mutator transaction binding the contract method 0x530e784f.
//
// Solidity: function setPriceOracle(address newPriceOracle) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetPriceOracle(newPriceOracle common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPriceOracle(&_PoolAddressesProvider.TransactOpts, newPriceOracle)
}

// SetPriceOracle is a paid mutator transaction binding the contract method 0x530e784f.
//
// Solidity: function setPriceOracle(address newPriceOracle) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetPriceOracle(newPriceOracle common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPriceOracle(&_PoolAddressesProvider.TransactOpts, newPriceOracle)
}

// SetPriceOracleSentinel is a paid mutator transaction binding the contract method 0x74944cec.
//
// Solidity: function setPriceOracleSentinel(address newPriceOracleSentinel) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) SetPriceOracleSentinel(opts *bind.TransactOpts, newPriceOracleSentinel common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "setPriceOracleSentinel", newPriceOracleSentinel)
}

// SetPriceOracleSentinel is a paid mutator transaction binding the contract method 0x74944cec.
//
// Solidity: function setPriceOracleSentinel(address newPriceOracleSentinel) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) SetPriceOracleSentinel(newPriceOracleSentinel common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPriceOracleSentinel(&_PoolAddressesProvider.TransactOpts, newPriceOracleSentinel)
}

// SetPriceOracleSentinel is a paid mutator transaction binding the contract method 0x74944cec.
//
// Solidity: function setPriceOracleSentinel(address newPriceOracleSentinel) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) SetPriceOracleSentinel(newPriceOracleSentinel common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.SetPriceOracleSentinel(&_PoolAddressesProvider.TransactOpts, newPriceOracleSentinel)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PoolAddressesProvider *PoolAddressesProviderSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.TransferOwnership(&_PoolAddressesProvider.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_PoolAddressesProvider *PoolAddressesProviderTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _PoolAddressesProvider.Contract.TransferOwnership(&_PoolAddressesProvider.TransactOpts, newOwner)
}

// PoolAddressesProviderACLAdminUpdatedIterator is returned from FilterACLAdminUpdated and is used to iterate over the raw logs and unpacked data for ACLAdminUpdated events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderACLAdminUpdatedIterator struct {
	Event *PoolAddressesProviderACLAdminUpdated // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderACLAdminUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderACLAdminUpdated)
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
		it.Event = new(PoolAddressesProviderACLAdminUpdated)
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
func (it *PoolAddressesProviderACLAdminUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderACLAdminUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderACLAdminUpdated represents a ACLAdminUpdated event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderACLAdminUpdated struct {
	OldAddress common.Address
	NewAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterACLAdminUpdated is a free log retrieval operation binding the contract event 0xe9cf53972264dc95304fd424458745019ddfca0e37ae8f703d74772c41ad115b.
//
// Solidity: event ACLAdminUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterACLAdminUpdated(opts *bind.FilterOpts, oldAddress []common.Address, newAddress []common.Address) (*PoolAddressesProviderACLAdminUpdatedIterator, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "ACLAdminUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderACLAdminUpdatedIterator{contract: _PoolAddressesProvider.contract, event: "ACLAdminUpdated", logs: logs, sub: sub}, nil
}

// WatchACLAdminUpdated is a free log subscription operation binding the contract event 0xe9cf53972264dc95304fd424458745019ddfca0e37ae8f703d74772c41ad115b.
//
// Solidity: event ACLAdminUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchACLAdminUpdated(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderACLAdminUpdated, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "ACLAdminUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderACLAdminUpdated)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "ACLAdminUpdated", log); err != nil {
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

// ParseACLAdminUpdated is a log parse operation binding the contract event 0xe9cf53972264dc95304fd424458745019ddfca0e37ae8f703d74772c41ad115b.
//
// Solidity: event ACLAdminUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParseACLAdminUpdated(log types.Log) (*PoolAddressesProviderACLAdminUpdated, error) {
	event := new(PoolAddressesProviderACLAdminUpdated)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "ACLAdminUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderACLManagerUpdatedIterator is returned from FilterACLManagerUpdated and is used to iterate over the raw logs and unpacked data for ACLManagerUpdated events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderACLManagerUpdatedIterator struct {
	Event *PoolAddressesProviderACLManagerUpdated // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderACLManagerUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderACLManagerUpdated)
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
		it.Event = new(PoolAddressesProviderACLManagerUpdated)
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
func (it *PoolAddressesProviderACLManagerUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderACLManagerUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderACLManagerUpdated represents a ACLManagerUpdated event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderACLManagerUpdated struct {
	OldAddress common.Address
	NewAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterACLManagerUpdated is a free log retrieval operation binding the contract event 0xb30efa04327bb8a537d61cc1e5c48095345ad18ef7cc04e6bacf7dfb6caaf507.
//
// Solidity: event ACLManagerUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterACLManagerUpdated(opts *bind.FilterOpts, oldAddress []common.Address, newAddress []common.Address) (*PoolAddressesProviderACLManagerUpdatedIterator, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "ACLManagerUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderACLManagerUpdatedIterator{contract: _PoolAddressesProvider.contract, event: "ACLManagerUpdated", logs: logs, sub: sub}, nil
}

// WatchACLManagerUpdated is a free log subscription operation binding the contract event 0xb30efa04327bb8a537d61cc1e5c48095345ad18ef7cc04e6bacf7dfb6caaf507.
//
// Solidity: event ACLManagerUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchACLManagerUpdated(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderACLManagerUpdated, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "ACLManagerUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderACLManagerUpdated)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "ACLManagerUpdated", log); err != nil {
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

// ParseACLManagerUpdated is a log parse operation binding the contract event 0xb30efa04327bb8a537d61cc1e5c48095345ad18ef7cc04e6bacf7dfb6caaf507.
//
// Solidity: event ACLManagerUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParseACLManagerUpdated(log types.Log) (*PoolAddressesProviderACLManagerUpdated, error) {
	event := new(PoolAddressesProviderACLManagerUpdated)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "ACLManagerUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderAddressSetIterator is returned from FilterAddressSet and is used to iterate over the raw logs and unpacked data for AddressSet events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderAddressSetIterator struct {
	Event *PoolAddressesProviderAddressSet // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderAddressSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderAddressSet)
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
		it.Event = new(PoolAddressesProviderAddressSet)
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
func (it *PoolAddressesProviderAddressSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderAddressSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderAddressSet represents a AddressSet event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderAddressSet struct {
	Id         [32]byte
	OldAddress common.Address
	NewAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterAddressSet is a free log retrieval operation binding the contract event 0x9ef0e8c8e52743bb38b83b17d9429141d494b8041ca6d616a6c77cebae9cd8b7.
//
// Solidity: event AddressSet(bytes32 indexed id, address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterAddressSet(opts *bind.FilterOpts, id [][32]byte, oldAddress []common.Address, newAddress []common.Address) (*PoolAddressesProviderAddressSetIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "AddressSet", idRule, oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderAddressSetIterator{contract: _PoolAddressesProvider.contract, event: "AddressSet", logs: logs, sub: sub}, nil
}

// WatchAddressSet is a free log subscription operation binding the contract event 0x9ef0e8c8e52743bb38b83b17d9429141d494b8041ca6d616a6c77cebae9cd8b7.
//
// Solidity: event AddressSet(bytes32 indexed id, address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchAddressSet(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderAddressSet, id [][32]byte, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "AddressSet", idRule, oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderAddressSet)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "AddressSet", log); err != nil {
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

// ParseAddressSet is a log parse operation binding the contract event 0x9ef0e8c8e52743bb38b83b17d9429141d494b8041ca6d616a6c77cebae9cd8b7.
//
// Solidity: event AddressSet(bytes32 indexed id, address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParseAddressSet(log types.Log) (*PoolAddressesProviderAddressSet, error) {
	event := new(PoolAddressesProviderAddressSet)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "AddressSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderAddressSetAsProxyIterator is returned from FilterAddressSetAsProxy and is used to iterate over the raw logs and unpacked data for AddressSetAsProxy events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderAddressSetAsProxyIterator struct {
	Event *PoolAddressesProviderAddressSetAsProxy // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderAddressSetAsProxyIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderAddressSetAsProxy)
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
		it.Event = new(PoolAddressesProviderAddressSetAsProxy)
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
func (it *PoolAddressesProviderAddressSetAsProxyIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderAddressSetAsProxyIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderAddressSetAsProxy represents a AddressSetAsProxy event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderAddressSetAsProxy struct {
	Id                       [32]byte
	ProxyAddress             common.Address
	OldImplementationAddress common.Address
	NewImplementationAddress common.Address
	Raw                      types.Log // Blockchain specific contextual infos
}

// FilterAddressSetAsProxy is a free log retrieval operation binding the contract event 0x3bbd45b5429b385e3fb37ad5cd1cd1435a3c8ec32196c7937597365a3fd3e99c.
//
// Solidity: event AddressSetAsProxy(bytes32 indexed id, address indexed proxyAddress, address oldImplementationAddress, address indexed newImplementationAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterAddressSetAsProxy(opts *bind.FilterOpts, id [][32]byte, proxyAddress []common.Address, newImplementationAddress []common.Address) (*PoolAddressesProviderAddressSetAsProxyIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var proxyAddressRule []interface{}
	for _, proxyAddressItem := range proxyAddress {
		proxyAddressRule = append(proxyAddressRule, proxyAddressItem)
	}

	var newImplementationAddressRule []interface{}
	for _, newImplementationAddressItem := range newImplementationAddress {
		newImplementationAddressRule = append(newImplementationAddressRule, newImplementationAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "AddressSetAsProxy", idRule, proxyAddressRule, newImplementationAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderAddressSetAsProxyIterator{contract: _PoolAddressesProvider.contract, event: "AddressSetAsProxy", logs: logs, sub: sub}, nil
}

// WatchAddressSetAsProxy is a free log subscription operation binding the contract event 0x3bbd45b5429b385e3fb37ad5cd1cd1435a3c8ec32196c7937597365a3fd3e99c.
//
// Solidity: event AddressSetAsProxy(bytes32 indexed id, address indexed proxyAddress, address oldImplementationAddress, address indexed newImplementationAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchAddressSetAsProxy(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderAddressSetAsProxy, id [][32]byte, proxyAddress []common.Address, newImplementationAddress []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var proxyAddressRule []interface{}
	for _, proxyAddressItem := range proxyAddress {
		proxyAddressRule = append(proxyAddressRule, proxyAddressItem)
	}

	var newImplementationAddressRule []interface{}
	for _, newImplementationAddressItem := range newImplementationAddress {
		newImplementationAddressRule = append(newImplementationAddressRule, newImplementationAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "AddressSetAsProxy", idRule, proxyAddressRule, newImplementationAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderAddressSetAsProxy)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "AddressSetAsProxy", log); err != nil {
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

// ParseAddressSetAsProxy is a log parse operation binding the contract event 0x3bbd45b5429b385e3fb37ad5cd1cd1435a3c8ec32196c7937597365a3fd3e99c.
//
// Solidity: event AddressSetAsProxy(bytes32 indexed id, address indexed proxyAddress, address oldImplementationAddress, address indexed newImplementationAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParseAddressSetAsProxy(log types.Log) (*PoolAddressesProviderAddressSetAsProxy, error) {
	event := new(PoolAddressesProviderAddressSetAsProxy)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "AddressSetAsProxy", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderMarketIdSetIterator is returned from FilterMarketIdSet and is used to iterate over the raw logs and unpacked data for MarketIdSet events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderMarketIdSetIterator struct {
	Event *PoolAddressesProviderMarketIdSet // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderMarketIdSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderMarketIdSet)
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
		it.Event = new(PoolAddressesProviderMarketIdSet)
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
func (it *PoolAddressesProviderMarketIdSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderMarketIdSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderMarketIdSet represents a MarketIdSet event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderMarketIdSet struct {
	OldMarketId common.Hash
	NewMarketId common.Hash
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterMarketIdSet is a free log retrieval operation binding the contract event 0xe685c8cdecc6030c45030fd54778812cb84ed8e4467c38294403d68ba7860823.
//
// Solidity: event MarketIdSet(string indexed oldMarketId, string indexed newMarketId)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterMarketIdSet(opts *bind.FilterOpts, oldMarketId []string, newMarketId []string) (*PoolAddressesProviderMarketIdSetIterator, error) {

	var oldMarketIdRule []interface{}
	for _, oldMarketIdItem := range oldMarketId {
		oldMarketIdRule = append(oldMarketIdRule, oldMarketIdItem)
	}
	var newMarketIdRule []interface{}
	for _, newMarketIdItem := range newMarketId {
		newMarketIdRule = append(newMarketIdRule, newMarketIdItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "MarketIdSet", oldMarketIdRule, newMarketIdRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderMarketIdSetIterator{contract: _PoolAddressesProvider.contract, event: "MarketIdSet", logs: logs, sub: sub}, nil
}

// WatchMarketIdSet is a free log subscription operation binding the contract event 0xe685c8cdecc6030c45030fd54778812cb84ed8e4467c38294403d68ba7860823.
//
// Solidity: event MarketIdSet(string indexed oldMarketId, string indexed newMarketId)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchMarketIdSet(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderMarketIdSet, oldMarketId []string, newMarketId []string) (event.Subscription, error) {

	var oldMarketIdRule []interface{}
	for _, oldMarketIdItem := range oldMarketId {
		oldMarketIdRule = append(oldMarketIdRule, oldMarketIdItem)
	}
	var newMarketIdRule []interface{}
	for _, newMarketIdItem := range newMarketId {
		newMarketIdRule = append(newMarketIdRule, newMarketIdItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "MarketIdSet", oldMarketIdRule, newMarketIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderMarketIdSet)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "MarketIdSet", log); err != nil {
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

// ParseMarketIdSet is a log parse operation binding the contract event 0xe685c8cdecc6030c45030fd54778812cb84ed8e4467c38294403d68ba7860823.
//
// Solidity: event MarketIdSet(string indexed oldMarketId, string indexed newMarketId)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParseMarketIdSet(log types.Log) (*PoolAddressesProviderMarketIdSet, error) {
	event := new(PoolAddressesProviderMarketIdSet)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "MarketIdSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderOwnershipTransferredIterator struct {
	Event *PoolAddressesProviderOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderOwnershipTransferred)
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
		it.Event = new(PoolAddressesProviderOwnershipTransferred)
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
func (it *PoolAddressesProviderOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderOwnershipTransferred represents a OwnershipTransferred event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*PoolAddressesProviderOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderOwnershipTransferredIterator{contract: _PoolAddressesProvider.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderOwnershipTransferred)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParseOwnershipTransferred(log types.Log) (*PoolAddressesProviderOwnershipTransferred, error) {
	event := new(PoolAddressesProviderOwnershipTransferred)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderPoolConfiguratorUpdatedIterator is returned from FilterPoolConfiguratorUpdated and is used to iterate over the raw logs and unpacked data for PoolConfiguratorUpdated events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPoolConfiguratorUpdatedIterator struct {
	Event *PoolAddressesProviderPoolConfiguratorUpdated // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderPoolConfiguratorUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderPoolConfiguratorUpdated)
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
		it.Event = new(PoolAddressesProviderPoolConfiguratorUpdated)
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
func (it *PoolAddressesProviderPoolConfiguratorUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderPoolConfiguratorUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderPoolConfiguratorUpdated represents a PoolConfiguratorUpdated event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPoolConfiguratorUpdated struct {
	OldAddress common.Address
	NewAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterPoolConfiguratorUpdated is a free log retrieval operation binding the contract event 0x8932892569eba59c8382a089d9b732d1f49272878775235761a2a6b0309cd465.
//
// Solidity: event PoolConfiguratorUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterPoolConfiguratorUpdated(opts *bind.FilterOpts, oldAddress []common.Address, newAddress []common.Address) (*PoolAddressesProviderPoolConfiguratorUpdatedIterator, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "PoolConfiguratorUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderPoolConfiguratorUpdatedIterator{contract: _PoolAddressesProvider.contract, event: "PoolConfiguratorUpdated", logs: logs, sub: sub}, nil
}

// WatchPoolConfiguratorUpdated is a free log subscription operation binding the contract event 0x8932892569eba59c8382a089d9b732d1f49272878775235761a2a6b0309cd465.
//
// Solidity: event PoolConfiguratorUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchPoolConfiguratorUpdated(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderPoolConfiguratorUpdated, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "PoolConfiguratorUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderPoolConfiguratorUpdated)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "PoolConfiguratorUpdated", log); err != nil {
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

// ParsePoolConfiguratorUpdated is a log parse operation binding the contract event 0x8932892569eba59c8382a089d9b732d1f49272878775235761a2a6b0309cd465.
//
// Solidity: event PoolConfiguratorUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParsePoolConfiguratorUpdated(log types.Log) (*PoolAddressesProviderPoolConfiguratorUpdated, error) {
	event := new(PoolAddressesProviderPoolConfiguratorUpdated)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "PoolConfiguratorUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderPoolDataProviderUpdatedIterator is returned from FilterPoolDataProviderUpdated and is used to iterate over the raw logs and unpacked data for PoolDataProviderUpdated events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPoolDataProviderUpdatedIterator struct {
	Event *PoolAddressesProviderPoolDataProviderUpdated // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderPoolDataProviderUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderPoolDataProviderUpdated)
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
		it.Event = new(PoolAddressesProviderPoolDataProviderUpdated)
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
func (it *PoolAddressesProviderPoolDataProviderUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderPoolDataProviderUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderPoolDataProviderUpdated represents a PoolDataProviderUpdated event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPoolDataProviderUpdated struct {
	OldAddress common.Address
	NewAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterPoolDataProviderUpdated is a free log retrieval operation binding the contract event 0xc853974cfbf81487a14a23565917bee63f527853bcb5fa54f2ae1cdf8a38356d.
//
// Solidity: event PoolDataProviderUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterPoolDataProviderUpdated(opts *bind.FilterOpts, oldAddress []common.Address, newAddress []common.Address) (*PoolAddressesProviderPoolDataProviderUpdatedIterator, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "PoolDataProviderUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderPoolDataProviderUpdatedIterator{contract: _PoolAddressesProvider.contract, event: "PoolDataProviderUpdated", logs: logs, sub: sub}, nil
}

// WatchPoolDataProviderUpdated is a free log subscription operation binding the contract event 0xc853974cfbf81487a14a23565917bee63f527853bcb5fa54f2ae1cdf8a38356d.
//
// Solidity: event PoolDataProviderUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchPoolDataProviderUpdated(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderPoolDataProviderUpdated, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "PoolDataProviderUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderPoolDataProviderUpdated)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "PoolDataProviderUpdated", log); err != nil {
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

// ParsePoolDataProviderUpdated is a log parse operation binding the contract event 0xc853974cfbf81487a14a23565917bee63f527853bcb5fa54f2ae1cdf8a38356d.
//
// Solidity: event PoolDataProviderUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParsePoolDataProviderUpdated(log types.Log) (*PoolAddressesProviderPoolDataProviderUpdated, error) {
	event := new(PoolAddressesProviderPoolDataProviderUpdated)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "PoolDataProviderUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderPoolUpdatedIterator is returned from FilterPoolUpdated and is used to iterate over the raw logs and unpacked data for PoolUpdated events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPoolUpdatedIterator struct {
	Event *PoolAddressesProviderPoolUpdated // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderPoolUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderPoolUpdated)
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
		it.Event = new(PoolAddressesProviderPoolUpdated)
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
func (it *PoolAddressesProviderPoolUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderPoolUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderPoolUpdated represents a PoolUpdated event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPoolUpdated struct {
	OldAddress common.Address
	NewAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterPoolUpdated is a free log retrieval operation binding the contract event 0x90affc163f1a2dfedcd36aa02ed992eeeba8100a4014f0b4cdc20ea265a66627.
//
// Solidity: event PoolUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterPoolUpdated(opts *bind.FilterOpts, oldAddress []common.Address, newAddress []common.Address) (*PoolAddressesProviderPoolUpdatedIterator, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "PoolUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderPoolUpdatedIterator{contract: _PoolAddressesProvider.contract, event: "PoolUpdated", logs: logs, sub: sub}, nil
}

// WatchPoolUpdated is a free log subscription operation binding the contract event 0x90affc163f1a2dfedcd36aa02ed992eeeba8100a4014f0b4cdc20ea265a66627.
//
// Solidity: event PoolUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchPoolUpdated(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderPoolUpdated, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "PoolUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderPoolUpdated)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "PoolUpdated", log); err != nil {
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

// ParsePoolUpdated is a log parse operation binding the contract event 0x90affc163f1a2dfedcd36aa02ed992eeeba8100a4014f0b4cdc20ea265a66627.
//
// Solidity: event PoolUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParsePoolUpdated(log types.Log) (*PoolAddressesProviderPoolUpdated, error) {
	event := new(PoolAddressesProviderPoolUpdated)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "PoolUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderPriceOracleSentinelUpdatedIterator is returned from FilterPriceOracleSentinelUpdated and is used to iterate over the raw logs and unpacked data for PriceOracleSentinelUpdated events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPriceOracleSentinelUpdatedIterator struct {
	Event *PoolAddressesProviderPriceOracleSentinelUpdated // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderPriceOracleSentinelUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderPriceOracleSentinelUpdated)
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
		it.Event = new(PoolAddressesProviderPriceOracleSentinelUpdated)
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
func (it *PoolAddressesProviderPriceOracleSentinelUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderPriceOracleSentinelUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderPriceOracleSentinelUpdated represents a PriceOracleSentinelUpdated event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPriceOracleSentinelUpdated struct {
	OldAddress common.Address
	NewAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterPriceOracleSentinelUpdated is a free log retrieval operation binding the contract event 0x5326514eeca90494a14bedabcff812a0e683029ee85d1e23824d44fd14cd6ae7.
//
// Solidity: event PriceOracleSentinelUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterPriceOracleSentinelUpdated(opts *bind.FilterOpts, oldAddress []common.Address, newAddress []common.Address) (*PoolAddressesProviderPriceOracleSentinelUpdatedIterator, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "PriceOracleSentinelUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderPriceOracleSentinelUpdatedIterator{contract: _PoolAddressesProvider.contract, event: "PriceOracleSentinelUpdated", logs: logs, sub: sub}, nil
}

// WatchPriceOracleSentinelUpdated is a free log subscription operation binding the contract event 0x5326514eeca90494a14bedabcff812a0e683029ee85d1e23824d44fd14cd6ae7.
//
// Solidity: event PriceOracleSentinelUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchPriceOracleSentinelUpdated(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderPriceOracleSentinelUpdated, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "PriceOracleSentinelUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderPriceOracleSentinelUpdated)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "PriceOracleSentinelUpdated", log); err != nil {
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

// ParsePriceOracleSentinelUpdated is a log parse operation binding the contract event 0x5326514eeca90494a14bedabcff812a0e683029ee85d1e23824d44fd14cd6ae7.
//
// Solidity: event PriceOracleSentinelUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParsePriceOracleSentinelUpdated(log types.Log) (*PoolAddressesProviderPriceOracleSentinelUpdated, error) {
	event := new(PoolAddressesProviderPriceOracleSentinelUpdated)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "PriceOracleSentinelUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderPriceOracleUpdatedIterator is returned from FilterPriceOracleUpdated and is used to iterate over the raw logs and unpacked data for PriceOracleUpdated events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPriceOracleUpdatedIterator struct {
	Event *PoolAddressesProviderPriceOracleUpdated // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderPriceOracleUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderPriceOracleUpdated)
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
		it.Event = new(PoolAddressesProviderPriceOracleUpdated)
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
func (it *PoolAddressesProviderPriceOracleUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderPriceOracleUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderPriceOracleUpdated represents a PriceOracleUpdated event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderPriceOracleUpdated struct {
	OldAddress common.Address
	NewAddress common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterPriceOracleUpdated is a free log retrieval operation binding the contract event 0x56b5f80d8cac1479698aa7d01605fd6111e90b15fc4d2b377417f46034876cbd.
//
// Solidity: event PriceOracleUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterPriceOracleUpdated(opts *bind.FilterOpts, oldAddress []common.Address, newAddress []common.Address) (*PoolAddressesProviderPriceOracleUpdatedIterator, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "PriceOracleUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderPriceOracleUpdatedIterator{contract: _PoolAddressesProvider.contract, event: "PriceOracleUpdated", logs: logs, sub: sub}, nil
}

// WatchPriceOracleUpdated is a free log subscription operation binding the contract event 0x56b5f80d8cac1479698aa7d01605fd6111e90b15fc4d2b377417f46034876cbd.
//
// Solidity: event PriceOracleUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchPriceOracleUpdated(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderPriceOracleUpdated, oldAddress []common.Address, newAddress []common.Address) (event.Subscription, error) {

	var oldAddressRule []interface{}
	for _, oldAddressItem := range oldAddress {
		oldAddressRule = append(oldAddressRule, oldAddressItem)
	}
	var newAddressRule []interface{}
	for _, newAddressItem := range newAddress {
		newAddressRule = append(newAddressRule, newAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "PriceOracleUpdated", oldAddressRule, newAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderPriceOracleUpdated)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "PriceOracleUpdated", log); err != nil {
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

// ParsePriceOracleUpdated is a log parse operation binding the contract event 0x56b5f80d8cac1479698aa7d01605fd6111e90b15fc4d2b377417f46034876cbd.
//
// Solidity: event PriceOracleUpdated(address indexed oldAddress, address indexed newAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParsePriceOracleUpdated(log types.Log) (*PoolAddressesProviderPriceOracleUpdated, error) {
	event := new(PoolAddressesProviderPriceOracleUpdated)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "PriceOracleUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PoolAddressesProviderProxyCreatedIterator is returned from FilterProxyCreated and is used to iterate over the raw logs and unpacked data for ProxyCreated events raised by the PoolAddressesProvider contract.
type PoolAddressesProviderProxyCreatedIterator struct {
	Event *PoolAddressesProviderProxyCreated // Event containing the contract specifics and raw log

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
func (it *PoolAddressesProviderProxyCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PoolAddressesProviderProxyCreated)
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
		it.Event = new(PoolAddressesProviderProxyCreated)
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
func (it *PoolAddressesProviderProxyCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PoolAddressesProviderProxyCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PoolAddressesProviderProxyCreated represents a ProxyCreated event raised by the PoolAddressesProvider contract.
type PoolAddressesProviderProxyCreated struct {
	Id                    [32]byte
	ProxyAddress          common.Address
	ImplementationAddress common.Address
	Raw                   types.Log // Blockchain specific contextual infos
}

// FilterProxyCreated is a free log retrieval operation binding the contract event 0x4a465a9bd819d9662563c1e11ae958f8109e437e7f4bf1c6ef0b9a7b3f35d478.
//
// Solidity: event ProxyCreated(bytes32 indexed id, address indexed proxyAddress, address indexed implementationAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) FilterProxyCreated(opts *bind.FilterOpts, id [][32]byte, proxyAddress []common.Address, implementationAddress []common.Address) (*PoolAddressesProviderProxyCreatedIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var proxyAddressRule []interface{}
	for _, proxyAddressItem := range proxyAddress {
		proxyAddressRule = append(proxyAddressRule, proxyAddressItem)
	}
	var implementationAddressRule []interface{}
	for _, implementationAddressItem := range implementationAddress {
		implementationAddressRule = append(implementationAddressRule, implementationAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.FilterLogs(opts, "ProxyCreated", idRule, proxyAddressRule, implementationAddressRule)
	if err != nil {
		return nil, err
	}
	return &PoolAddressesProviderProxyCreatedIterator{contract: _PoolAddressesProvider.contract, event: "ProxyCreated", logs: logs, sub: sub}, nil
}

// WatchProxyCreated is a free log subscription operation binding the contract event 0x4a465a9bd819d9662563c1e11ae958f8109e437e7f4bf1c6ef0b9a7b3f35d478.
//
// Solidity: event ProxyCreated(bytes32 indexed id, address indexed proxyAddress, address indexed implementationAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) WatchProxyCreated(opts *bind.WatchOpts, sink chan<- *PoolAddressesProviderProxyCreated, id [][32]byte, proxyAddress []common.Address, implementationAddress []common.Address) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var proxyAddressRule []interface{}
	for _, proxyAddressItem := range proxyAddress {
		proxyAddressRule = append(proxyAddressRule, proxyAddressItem)
	}
	var implementationAddressRule []interface{}
	for _, implementationAddressItem := range implementationAddress {
		implementationAddressRule = append(implementationAddressRule, implementationAddressItem)
	}

	logs, sub, err := _PoolAddressesProvider.contract.WatchLogs(opts, "ProxyCreated", idRule, proxyAddressRule, implementationAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PoolAddressesProviderProxyCreated)
				if err := _PoolAddressesProvider.contract.UnpackLog(event, "ProxyCreated", log); err != nil {
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

// ParseProxyCreated is a log parse operation binding the contract event 0x4a465a9bd819d9662563c1e11ae958f8109e437e7f4bf1c6ef0b9a7b3f35d478.
//
// Solidity: event ProxyCreated(bytes32 indexed id, address indexed proxyAddress, address indexed implementationAddress)
func (_PoolAddressesProvider *PoolAddressesProviderFilterer) ParseProxyCreated(log types.Log) (*PoolAddressesProviderProxyCreated, error) {
	event := new(PoolAddressesProviderProxyCreated)
	if err := _PoolAddressesProvider.contract.UnpackLog(event, "ProxyCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
