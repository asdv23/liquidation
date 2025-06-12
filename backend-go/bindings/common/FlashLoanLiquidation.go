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

// FlashLoanLiquidationMetaData contains all meta data concerning the FlashLoanLiquidation contract.
var FlashLoanLiquidationMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"ADDRESSES_PROVIDER\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPoolAddressesProvider\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"POOL\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"UPGRADE_INTERFACE_VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"aave_v3_pool\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIPool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"dex\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIDex\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"executeLiquidation\",\"inputs\":[{\"name\":\"collateralAsset\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"debtAsset\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"debtToCover\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"executeOperation\",\"inputs\":[{\"name\":\"asset\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"premium\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_initiator\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"params\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_aave_v3_pool\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_dex\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"proxiableUUID\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"upgradeToAndCall\",\"inputs\":[{\"name\":\"newImplementation\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"withdrawToken\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Liquidation\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"asset\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"collateralAsset\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"collateralAmount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SwapWithAggregator\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"profitToken\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"profit\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"collateralAsset\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"collateralBalance\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Upgraded\",\"inputs\":[{\"name\":\"implementation\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressEmptyCode\",\"inputs\":[{\"name\":\"target\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC1967InvalidImplementation\",\"inputs\":[{\"name\":\"implementation\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ERC1967NonPayable\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"FailedCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidInitialization\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NotInitializing\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"UUPSUnauthorizedCallContext\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UUPSUnsupportedProxiableUUID\",\"inputs\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]",
}

// FlashLoanLiquidationABI is the input ABI used to generate the binding from.
// Deprecated: Use FlashLoanLiquidationMetaData.ABI instead.
var FlashLoanLiquidationABI = FlashLoanLiquidationMetaData.ABI

// FlashLoanLiquidation is an auto generated Go binding around an Ethereum contract.
type FlashLoanLiquidation struct {
	FlashLoanLiquidationCaller     // Read-only binding to the contract
	FlashLoanLiquidationTransactor // Write-only binding to the contract
	FlashLoanLiquidationFilterer   // Log filterer for contract events
}

// FlashLoanLiquidationCaller is an auto generated read-only Go binding around an Ethereum contract.
type FlashLoanLiquidationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlashLoanLiquidationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type FlashLoanLiquidationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlashLoanLiquidationFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type FlashLoanLiquidationFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// FlashLoanLiquidationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type FlashLoanLiquidationSession struct {
	Contract     *FlashLoanLiquidation // Generic contract binding to set the session for
	CallOpts     bind.CallOpts         // Call options to use throughout this session
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// FlashLoanLiquidationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type FlashLoanLiquidationCallerSession struct {
	Contract *FlashLoanLiquidationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts               // Call options to use throughout this session
}

// FlashLoanLiquidationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type FlashLoanLiquidationTransactorSession struct {
	Contract     *FlashLoanLiquidationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// FlashLoanLiquidationRaw is an auto generated low-level Go binding around an Ethereum contract.
type FlashLoanLiquidationRaw struct {
	Contract *FlashLoanLiquidation // Generic contract binding to access the raw methods on
}

// FlashLoanLiquidationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type FlashLoanLiquidationCallerRaw struct {
	Contract *FlashLoanLiquidationCaller // Generic read-only contract binding to access the raw methods on
}

// FlashLoanLiquidationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type FlashLoanLiquidationTransactorRaw struct {
	Contract *FlashLoanLiquidationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewFlashLoanLiquidation creates a new instance of FlashLoanLiquidation, bound to a specific deployed contract.
func NewFlashLoanLiquidation(address common.Address, backend bind.ContractBackend) (*FlashLoanLiquidation, error) {
	contract, err := bindFlashLoanLiquidation(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &FlashLoanLiquidation{FlashLoanLiquidationCaller: FlashLoanLiquidationCaller{contract: contract}, FlashLoanLiquidationTransactor: FlashLoanLiquidationTransactor{contract: contract}, FlashLoanLiquidationFilterer: FlashLoanLiquidationFilterer{contract: contract}}, nil
}

// NewFlashLoanLiquidationCaller creates a new read-only instance of FlashLoanLiquidation, bound to a specific deployed contract.
func NewFlashLoanLiquidationCaller(address common.Address, caller bind.ContractCaller) (*FlashLoanLiquidationCaller, error) {
	contract, err := bindFlashLoanLiquidation(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &FlashLoanLiquidationCaller{contract: contract}, nil
}

// NewFlashLoanLiquidationTransactor creates a new write-only instance of FlashLoanLiquidation, bound to a specific deployed contract.
func NewFlashLoanLiquidationTransactor(address common.Address, transactor bind.ContractTransactor) (*FlashLoanLiquidationTransactor, error) {
	contract, err := bindFlashLoanLiquidation(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &FlashLoanLiquidationTransactor{contract: contract}, nil
}

// NewFlashLoanLiquidationFilterer creates a new log filterer instance of FlashLoanLiquidation, bound to a specific deployed contract.
func NewFlashLoanLiquidationFilterer(address common.Address, filterer bind.ContractFilterer) (*FlashLoanLiquidationFilterer, error) {
	contract, err := bindFlashLoanLiquidation(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &FlashLoanLiquidationFilterer{contract: contract}, nil
}

// bindFlashLoanLiquidation binds a generic wrapper to an already deployed contract.
func bindFlashLoanLiquidation(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := FlashLoanLiquidationMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FlashLoanLiquidation *FlashLoanLiquidationRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FlashLoanLiquidation.Contract.FlashLoanLiquidationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FlashLoanLiquidation *FlashLoanLiquidationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.FlashLoanLiquidationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FlashLoanLiquidation *FlashLoanLiquidationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.FlashLoanLiquidationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_FlashLoanLiquidation *FlashLoanLiquidationCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _FlashLoanLiquidation.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.contract.Transact(opts, method, params...)
}

// ADDRESSESPROVIDER is a free data retrieval call binding the contract method 0x0542975c.
//
// Solidity: function ADDRESSES_PROVIDER() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCaller) ADDRESSESPROVIDER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FlashLoanLiquidation.contract.Call(opts, &out, "ADDRESSES_PROVIDER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ADDRESSESPROVIDER is a free data retrieval call binding the contract method 0x0542975c.
//
// Solidity: function ADDRESSES_PROVIDER() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) ADDRESSESPROVIDER() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.ADDRESSESPROVIDER(&_FlashLoanLiquidation.CallOpts)
}

// ADDRESSESPROVIDER is a free data retrieval call binding the contract method 0x0542975c.
//
// Solidity: function ADDRESSES_PROVIDER() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCallerSession) ADDRESSESPROVIDER() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.ADDRESSESPROVIDER(&_FlashLoanLiquidation.CallOpts)
}

// POOL is a free data retrieval call binding the contract method 0x7535d246.
//
// Solidity: function POOL() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCaller) POOL(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FlashLoanLiquidation.contract.Call(opts, &out, "POOL")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// POOL is a free data retrieval call binding the contract method 0x7535d246.
//
// Solidity: function POOL() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) POOL() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.POOL(&_FlashLoanLiquidation.CallOpts)
}

// POOL is a free data retrieval call binding the contract method 0x7535d246.
//
// Solidity: function POOL() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCallerSession) POOL() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.POOL(&_FlashLoanLiquidation.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_FlashLoanLiquidation *FlashLoanLiquidationCaller) UPGRADEINTERFACEVERSION(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _FlashLoanLiquidation.contract.Call(opts, &out, "UPGRADE_INTERFACE_VERSION")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _FlashLoanLiquidation.Contract.UPGRADEINTERFACEVERSION(&_FlashLoanLiquidation.CallOpts)
}

// UPGRADEINTERFACEVERSION is a free data retrieval call binding the contract method 0xad3cb1cc.
//
// Solidity: function UPGRADE_INTERFACE_VERSION() view returns(string)
func (_FlashLoanLiquidation *FlashLoanLiquidationCallerSession) UPGRADEINTERFACEVERSION() (string, error) {
	return _FlashLoanLiquidation.Contract.UPGRADEINTERFACEVERSION(&_FlashLoanLiquidation.CallOpts)
}

// AaveV3Pool is a free data retrieval call binding the contract method 0xf5f8bbbc.
//
// Solidity: function aave_v3_pool() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCaller) AaveV3Pool(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FlashLoanLiquidation.contract.Call(opts, &out, "aave_v3_pool")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// AaveV3Pool is a free data retrieval call binding the contract method 0xf5f8bbbc.
//
// Solidity: function aave_v3_pool() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) AaveV3Pool() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.AaveV3Pool(&_FlashLoanLiquidation.CallOpts)
}

// AaveV3Pool is a free data retrieval call binding the contract method 0xf5f8bbbc.
//
// Solidity: function aave_v3_pool() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCallerSession) AaveV3Pool() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.AaveV3Pool(&_FlashLoanLiquidation.CallOpts)
}

// Dex is a free data retrieval call binding the contract method 0x692058c2.
//
// Solidity: function dex() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCaller) Dex(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FlashLoanLiquidation.contract.Call(opts, &out, "dex")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Dex is a free data retrieval call binding the contract method 0x692058c2.
//
// Solidity: function dex() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) Dex() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.Dex(&_FlashLoanLiquidation.CallOpts)
}

// Dex is a free data retrieval call binding the contract method 0x692058c2.
//
// Solidity: function dex() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCallerSession) Dex() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.Dex(&_FlashLoanLiquidation.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _FlashLoanLiquidation.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) Owner() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.Owner(&_FlashLoanLiquidation.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_FlashLoanLiquidation *FlashLoanLiquidationCallerSession) Owner() (common.Address, error) {
	return _FlashLoanLiquidation.Contract.Owner(&_FlashLoanLiquidation.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_FlashLoanLiquidation *FlashLoanLiquidationCaller) ProxiableUUID(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _FlashLoanLiquidation.contract.Call(opts, &out, "proxiableUUID")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) ProxiableUUID() ([32]byte, error) {
	return _FlashLoanLiquidation.Contract.ProxiableUUID(&_FlashLoanLiquidation.CallOpts)
}

// ProxiableUUID is a free data retrieval call binding the contract method 0x52d1902d.
//
// Solidity: function proxiableUUID() view returns(bytes32)
func (_FlashLoanLiquidation *FlashLoanLiquidationCallerSession) ProxiableUUID() ([32]byte, error) {
	return _FlashLoanLiquidation.Contract.ProxiableUUID(&_FlashLoanLiquidation.CallOpts)
}

// ExecuteLiquidation is a paid mutator transaction binding the contract method 0xa19dd4b2.
//
// Solidity: function executeLiquidation(address collateralAsset, address debtAsset, address user, uint256 debtToCover, bytes data) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactor) ExecuteLiquidation(opts *bind.TransactOpts, collateralAsset common.Address, debtAsset common.Address, user common.Address, debtToCover *big.Int, data []byte) (*types.Transaction, error) {
	return _FlashLoanLiquidation.contract.Transact(opts, "executeLiquidation", collateralAsset, debtAsset, user, debtToCover, data)
}

// ExecuteLiquidation is a paid mutator transaction binding the contract method 0xa19dd4b2.
//
// Solidity: function executeLiquidation(address collateralAsset, address debtAsset, address user, uint256 debtToCover, bytes data) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) ExecuteLiquidation(collateralAsset common.Address, debtAsset common.Address, user common.Address, debtToCover *big.Int, data []byte) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.ExecuteLiquidation(&_FlashLoanLiquidation.TransactOpts, collateralAsset, debtAsset, user, debtToCover, data)
}

// ExecuteLiquidation is a paid mutator transaction binding the contract method 0xa19dd4b2.
//
// Solidity: function executeLiquidation(address collateralAsset, address debtAsset, address user, uint256 debtToCover, bytes data) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactorSession) ExecuteLiquidation(collateralAsset common.Address, debtAsset common.Address, user common.Address, debtToCover *big.Int, data []byte) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.ExecuteLiquidation(&_FlashLoanLiquidation.TransactOpts, collateralAsset, debtAsset, user, debtToCover, data)
}

// ExecuteOperation is a paid mutator transaction binding the contract method 0x1b11d0ff.
//
// Solidity: function executeOperation(address asset, uint256 amount, uint256 premium, address _initiator, bytes params) returns(bool)
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactor) ExecuteOperation(opts *bind.TransactOpts, asset common.Address, amount *big.Int, premium *big.Int, _initiator common.Address, params []byte) (*types.Transaction, error) {
	return _FlashLoanLiquidation.contract.Transact(opts, "executeOperation", asset, amount, premium, _initiator, params)
}

// ExecuteOperation is a paid mutator transaction binding the contract method 0x1b11d0ff.
//
// Solidity: function executeOperation(address asset, uint256 amount, uint256 premium, address _initiator, bytes params) returns(bool)
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) ExecuteOperation(asset common.Address, amount *big.Int, premium *big.Int, _initiator common.Address, params []byte) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.ExecuteOperation(&_FlashLoanLiquidation.TransactOpts, asset, amount, premium, _initiator, params)
}

// ExecuteOperation is a paid mutator transaction binding the contract method 0x1b11d0ff.
//
// Solidity: function executeOperation(address asset, uint256 amount, uint256 premium, address _initiator, bytes params) returns(bool)
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactorSession) ExecuteOperation(asset common.Address, amount *big.Int, premium *big.Int, _initiator common.Address, params []byte) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.ExecuteOperation(&_FlashLoanLiquidation.TransactOpts, asset, amount, premium, _initiator, params)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _aave_v3_pool, address _dex) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactor) Initialize(opts *bind.TransactOpts, _aave_v3_pool common.Address, _dex common.Address) (*types.Transaction, error) {
	return _FlashLoanLiquidation.contract.Transact(opts, "initialize", _aave_v3_pool, _dex)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _aave_v3_pool, address _dex) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) Initialize(_aave_v3_pool common.Address, _dex common.Address) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.Initialize(&_FlashLoanLiquidation.TransactOpts, _aave_v3_pool, _dex)
}

// Initialize is a paid mutator transaction binding the contract method 0x485cc955.
//
// Solidity: function initialize(address _aave_v3_pool, address _dex) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactorSession) Initialize(_aave_v3_pool common.Address, _dex common.Address) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.Initialize(&_FlashLoanLiquidation.TransactOpts, _aave_v3_pool, _dex)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _FlashLoanLiquidation.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) RenounceOwnership() (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.RenounceOwnership(&_FlashLoanLiquidation.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.RenounceOwnership(&_FlashLoanLiquidation.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _FlashLoanLiquidation.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.TransferOwnership(&_FlashLoanLiquidation.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.TransferOwnership(&_FlashLoanLiquidation.TransactOpts, newOwner)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactor) UpgradeToAndCall(opts *bind.TransactOpts, newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _FlashLoanLiquidation.contract.Transact(opts, "upgradeToAndCall", newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.UpgradeToAndCall(&_FlashLoanLiquidation.TransactOpts, newImplementation, data)
}

// UpgradeToAndCall is a paid mutator transaction binding the contract method 0x4f1ef286.
//
// Solidity: function upgradeToAndCall(address newImplementation, bytes data) payable returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactorSession) UpgradeToAndCall(newImplementation common.Address, data []byte) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.UpgradeToAndCall(&_FlashLoanLiquidation.TransactOpts, newImplementation, data)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x9e281a98.
//
// Solidity: function withdrawToken(address token, uint256 amount) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactor) WithdrawToken(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FlashLoanLiquidation.contract.Transact(opts, "withdrawToken", token, amount)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x9e281a98.
//
// Solidity: function withdrawToken(address token, uint256 amount) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationSession) WithdrawToken(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.WithdrawToken(&_FlashLoanLiquidation.TransactOpts, token, amount)
}

// WithdrawToken is a paid mutator transaction binding the contract method 0x9e281a98.
//
// Solidity: function withdrawToken(address token, uint256 amount) returns()
func (_FlashLoanLiquidation *FlashLoanLiquidationTransactorSession) WithdrawToken(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _FlashLoanLiquidation.Contract.WithdrawToken(&_FlashLoanLiquidation.TransactOpts, token, amount)
}

// FlashLoanLiquidationInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationInitializedIterator struct {
	Event *FlashLoanLiquidationInitialized // Event containing the contract specifics and raw log

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
func (it *FlashLoanLiquidationInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashLoanLiquidationInitialized)
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
		it.Event = new(FlashLoanLiquidationInitialized)
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
func (it *FlashLoanLiquidationInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashLoanLiquidationInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashLoanLiquidationInitialized represents a Initialized event raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationInitialized struct {
	Version uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) FilterInitialized(opts *bind.FilterOpts) (*FlashLoanLiquidationInitializedIterator, error) {

	logs, sub, err := _FlashLoanLiquidation.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &FlashLoanLiquidationInitializedIterator{contract: _FlashLoanLiquidation.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *FlashLoanLiquidationInitialized) (event.Subscription, error) {

	logs, sub, err := _FlashLoanLiquidation.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashLoanLiquidationInitialized)
				if err := _FlashLoanLiquidation.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xc7f505b2f371ae2175ee4913f4499e1f2633a7b5936321eed1cdaeb6115181d2.
//
// Solidity: event Initialized(uint64 version)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) ParseInitialized(log types.Log) (*FlashLoanLiquidationInitialized, error) {
	event := new(FlashLoanLiquidationInitialized)
	if err := _FlashLoanLiquidation.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlashLoanLiquidationLiquidationIterator is returned from FilterLiquidation and is used to iterate over the raw logs and unpacked data for Liquidation events raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationLiquidationIterator struct {
	Event *FlashLoanLiquidationLiquidation // Event containing the contract specifics and raw log

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
func (it *FlashLoanLiquidationLiquidationIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashLoanLiquidationLiquidation)
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
		it.Event = new(FlashLoanLiquidationLiquidation)
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
func (it *FlashLoanLiquidationLiquidationIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashLoanLiquidationLiquidationIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashLoanLiquidationLiquidation represents a Liquidation event raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationLiquidation struct {
	User             common.Address
	Asset            common.Address
	Amount           *big.Int
	CollateralAsset  common.Address
	CollateralAmount *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterLiquidation is a free log retrieval operation binding the contract event 0xad275fe017afe43f0a6e7b46bf3e3e9df267a25ade147288190402b123ade46b.
//
// Solidity: event Liquidation(address indexed user, address indexed asset, uint256 amount, address indexed collateralAsset, uint256 collateralAmount)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) FilterLiquidation(opts *bind.FilterOpts, user []common.Address, asset []common.Address, collateralAsset []common.Address) (*FlashLoanLiquidationLiquidationIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var assetRule []interface{}
	for _, assetItem := range asset {
		assetRule = append(assetRule, assetItem)
	}

	var collateralAssetRule []interface{}
	for _, collateralAssetItem := range collateralAsset {
		collateralAssetRule = append(collateralAssetRule, collateralAssetItem)
	}

	logs, sub, err := _FlashLoanLiquidation.contract.FilterLogs(opts, "Liquidation", userRule, assetRule, collateralAssetRule)
	if err != nil {
		return nil, err
	}
	return &FlashLoanLiquidationLiquidationIterator{contract: _FlashLoanLiquidation.contract, event: "Liquidation", logs: logs, sub: sub}, nil
}

// WatchLiquidation is a free log subscription operation binding the contract event 0xad275fe017afe43f0a6e7b46bf3e3e9df267a25ade147288190402b123ade46b.
//
// Solidity: event Liquidation(address indexed user, address indexed asset, uint256 amount, address indexed collateralAsset, uint256 collateralAmount)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) WatchLiquidation(opts *bind.WatchOpts, sink chan<- *FlashLoanLiquidationLiquidation, user []common.Address, asset []common.Address, collateralAsset []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var assetRule []interface{}
	for _, assetItem := range asset {
		assetRule = append(assetRule, assetItem)
	}

	var collateralAssetRule []interface{}
	for _, collateralAssetItem := range collateralAsset {
		collateralAssetRule = append(collateralAssetRule, collateralAssetItem)
	}

	logs, sub, err := _FlashLoanLiquidation.contract.WatchLogs(opts, "Liquidation", userRule, assetRule, collateralAssetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashLoanLiquidationLiquidation)
				if err := _FlashLoanLiquidation.contract.UnpackLog(event, "Liquidation", log); err != nil {
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

// ParseLiquidation is a log parse operation binding the contract event 0xad275fe017afe43f0a6e7b46bf3e3e9df267a25ade147288190402b123ade46b.
//
// Solidity: event Liquidation(address indexed user, address indexed asset, uint256 amount, address indexed collateralAsset, uint256 collateralAmount)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) ParseLiquidation(log types.Log) (*FlashLoanLiquidationLiquidation, error) {
	event := new(FlashLoanLiquidationLiquidation)
	if err := _FlashLoanLiquidation.contract.UnpackLog(event, "Liquidation", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlashLoanLiquidationOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationOwnershipTransferredIterator struct {
	Event *FlashLoanLiquidationOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *FlashLoanLiquidationOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashLoanLiquidationOwnershipTransferred)
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
		it.Event = new(FlashLoanLiquidationOwnershipTransferred)
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
func (it *FlashLoanLiquidationOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashLoanLiquidationOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashLoanLiquidationOwnershipTransferred represents a OwnershipTransferred event raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*FlashLoanLiquidationOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _FlashLoanLiquidation.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &FlashLoanLiquidationOwnershipTransferredIterator{contract: _FlashLoanLiquidation.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *FlashLoanLiquidationOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _FlashLoanLiquidation.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashLoanLiquidationOwnershipTransferred)
				if err := _FlashLoanLiquidation.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) ParseOwnershipTransferred(log types.Log) (*FlashLoanLiquidationOwnershipTransferred, error) {
	event := new(FlashLoanLiquidationOwnershipTransferred)
	if err := _FlashLoanLiquidation.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlashLoanLiquidationSwapWithAggregatorIterator is returned from FilterSwapWithAggregator and is used to iterate over the raw logs and unpacked data for SwapWithAggregator events raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationSwapWithAggregatorIterator struct {
	Event *FlashLoanLiquidationSwapWithAggregator // Event containing the contract specifics and raw log

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
func (it *FlashLoanLiquidationSwapWithAggregatorIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashLoanLiquidationSwapWithAggregator)
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
		it.Event = new(FlashLoanLiquidationSwapWithAggregator)
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
func (it *FlashLoanLiquidationSwapWithAggregatorIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashLoanLiquidationSwapWithAggregatorIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashLoanLiquidationSwapWithAggregator represents a SwapWithAggregator event raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationSwapWithAggregator struct {
	Target            common.Address
	ProfitToken       common.Address
	Profit            *big.Int
	CollateralAsset   common.Address
	CollateralBalance *big.Int
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterSwapWithAggregator is a free log retrieval operation binding the contract event 0x14233ca0bd4739a6a7f226fceb49ccfb881dfcbdf00875f38ae40ff5cc909d66.
//
// Solidity: event SwapWithAggregator(address indexed target, address indexed profitToken, uint256 profit, address indexed collateralAsset, uint256 collateralBalance)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) FilterSwapWithAggregator(opts *bind.FilterOpts, target []common.Address, profitToken []common.Address, collateralAsset []common.Address) (*FlashLoanLiquidationSwapWithAggregatorIterator, error) {

	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}
	var profitTokenRule []interface{}
	for _, profitTokenItem := range profitToken {
		profitTokenRule = append(profitTokenRule, profitTokenItem)
	}

	var collateralAssetRule []interface{}
	for _, collateralAssetItem := range collateralAsset {
		collateralAssetRule = append(collateralAssetRule, collateralAssetItem)
	}

	logs, sub, err := _FlashLoanLiquidation.contract.FilterLogs(opts, "SwapWithAggregator", targetRule, profitTokenRule, collateralAssetRule)
	if err != nil {
		return nil, err
	}
	return &FlashLoanLiquidationSwapWithAggregatorIterator{contract: _FlashLoanLiquidation.contract, event: "SwapWithAggregator", logs: logs, sub: sub}, nil
}

// WatchSwapWithAggregator is a free log subscription operation binding the contract event 0x14233ca0bd4739a6a7f226fceb49ccfb881dfcbdf00875f38ae40ff5cc909d66.
//
// Solidity: event SwapWithAggregator(address indexed target, address indexed profitToken, uint256 profit, address indexed collateralAsset, uint256 collateralBalance)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) WatchSwapWithAggregator(opts *bind.WatchOpts, sink chan<- *FlashLoanLiquidationSwapWithAggregator, target []common.Address, profitToken []common.Address, collateralAsset []common.Address) (event.Subscription, error) {

	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}
	var profitTokenRule []interface{}
	for _, profitTokenItem := range profitToken {
		profitTokenRule = append(profitTokenRule, profitTokenItem)
	}

	var collateralAssetRule []interface{}
	for _, collateralAssetItem := range collateralAsset {
		collateralAssetRule = append(collateralAssetRule, collateralAssetItem)
	}

	logs, sub, err := _FlashLoanLiquidation.contract.WatchLogs(opts, "SwapWithAggregator", targetRule, profitTokenRule, collateralAssetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashLoanLiquidationSwapWithAggregator)
				if err := _FlashLoanLiquidation.contract.UnpackLog(event, "SwapWithAggregator", log); err != nil {
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

// ParseSwapWithAggregator is a log parse operation binding the contract event 0x14233ca0bd4739a6a7f226fceb49ccfb881dfcbdf00875f38ae40ff5cc909d66.
//
// Solidity: event SwapWithAggregator(address indexed target, address indexed profitToken, uint256 profit, address indexed collateralAsset, uint256 collateralBalance)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) ParseSwapWithAggregator(log types.Log) (*FlashLoanLiquidationSwapWithAggregator, error) {
	event := new(FlashLoanLiquidationSwapWithAggregator)
	if err := _FlashLoanLiquidation.contract.UnpackLog(event, "SwapWithAggregator", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// FlashLoanLiquidationUpgradedIterator is returned from FilterUpgraded and is used to iterate over the raw logs and unpacked data for Upgraded events raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationUpgradedIterator struct {
	Event *FlashLoanLiquidationUpgraded // Event containing the contract specifics and raw log

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
func (it *FlashLoanLiquidationUpgradedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(FlashLoanLiquidationUpgraded)
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
		it.Event = new(FlashLoanLiquidationUpgraded)
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
func (it *FlashLoanLiquidationUpgradedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *FlashLoanLiquidationUpgradedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// FlashLoanLiquidationUpgraded represents a Upgraded event raised by the FlashLoanLiquidation contract.
type FlashLoanLiquidationUpgraded struct {
	Implementation common.Address
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterUpgraded is a free log retrieval operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) FilterUpgraded(opts *bind.FilterOpts, implementation []common.Address) (*FlashLoanLiquidationUpgradedIterator, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _FlashLoanLiquidation.contract.FilterLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return &FlashLoanLiquidationUpgradedIterator{contract: _FlashLoanLiquidation.contract, event: "Upgraded", logs: logs, sub: sub}, nil
}

// WatchUpgraded is a free log subscription operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) WatchUpgraded(opts *bind.WatchOpts, sink chan<- *FlashLoanLiquidationUpgraded, implementation []common.Address) (event.Subscription, error) {

	var implementationRule []interface{}
	for _, implementationItem := range implementation {
		implementationRule = append(implementationRule, implementationItem)
	}

	logs, sub, err := _FlashLoanLiquidation.contract.WatchLogs(opts, "Upgraded", implementationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(FlashLoanLiquidationUpgraded)
				if err := _FlashLoanLiquidation.contract.UnpackLog(event, "Upgraded", log); err != nil {
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

// ParseUpgraded is a log parse operation binding the contract event 0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b.
//
// Solidity: event Upgraded(address indexed implementation)
func (_FlashLoanLiquidation *FlashLoanLiquidationFilterer) ParseUpgraded(log types.Log) (*FlashLoanLiquidationUpgraded, error) {
	event := new(FlashLoanLiquidationUpgraded)
	if err := _FlashLoanLiquidation.contract.UnpackLog(event, "Upgraded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
