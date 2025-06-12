package aavev3

import (
	aavev3 "liquidation-bot/bindings/aavev3"
	"math"
)

func isUsingAsCollateral(config *aavev3.DataTypesUserConfigurationMap, reserveIndex int) bool {
	if reserveIndex > math.MaxInt8 {
		//   string public constant INVALID_RESERVE_INDEX = '74'; // 'Invalid reserve index'
		return false
	}

	return config.Data.Bit((reserveIndex<<1)+1)&1 != 0
}

func isBorrowing(config *aavev3.DataTypesUserConfigurationMap, reserveIndex int) bool {
	if reserveIndex > math.MaxInt8 {
		//   string public constant INVALID_RESERVE_INDEX = '74'; // 'Invalid reserve index'
		return false
	}

	return config.Data.Bit(reserveIndex<<1)&1 != 0
}
