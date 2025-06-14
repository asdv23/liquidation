package aavev3

import (
	aavev3 "liquidation-bot/bindings/aavev3"
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
