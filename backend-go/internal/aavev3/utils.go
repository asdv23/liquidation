package aavev3

import (
	"bytes"
	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// return (self.data >> ((reserveIndex << 1) + 1)) & 1 != 0;
func isUsingAsCollateral(config *aavev3.DataTypesUserConfigurationMap, reserveIndex int) bool {
	return config.Data.Bit((reserveIndex<<1)+1)&1 != 0
}

// return (self.data >> (reserveIndex << 1)) & 1 != 0;
func isBorrowing(config *aavev3.DataTypesUserConfigurationMap, reserveIndex int) bool {
	return config.Data.Bit(reserveIndex<<1)&1 != 0
}

// return (self.data >> (reserveIndex << 1)) & 3 != 0;
func isUsingAsCollateralOrBorrowing(config *aavev3.DataTypesUserConfigurationMap, reserveIndex int) bool {
	return isUsingAsCollateral(config, reserveIndex) || isBorrowing(config, reserveIndex)
}

func getSymbolAndDecimalsMulticall3Call3(abi *abi.ABI, asset common.Address) (bindings.Multicall3Call3, bindings.Multicall3Call3, error) {
	symbolCallData, err := abi.Pack("symbol")
	if err != nil {
		return bindings.Multicall3Call3{}, bindings.Multicall3Call3{}, err
	}
	decimalsCallData, err := abi.Pack("decimals")
	if err != nil {
		return bindings.Multicall3Call3{}, bindings.Multicall3Call3{}, err
	}
	return bindings.Multicall3Call3{
			Target:   asset,
			CallData: symbolCallData,
		}, bindings.Multicall3Call3{
			Target:   asset,
			CallData: decimalsCallData,
		}, nil
}

func decodeSymbol(returnData []byte, erc20Abi *abi.ABI) string {
	var symbol string
	err := erc20Abi.UnpackIntoInterface(&symbol, "symbol", returnData)
	if err != nil {
		return string(bytes.TrimRight(returnData, "\x00"))
	}

	return symbol
}
