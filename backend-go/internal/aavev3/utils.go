package aavev3

import (
	"bytes"
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	MIN_DEBT_BASE = big.NewInt(2e8)
	USD_DECIMALS  = big.NewFloat(1e8)
	HF_DECIMALS   = big.NewFloat(1e18)
)

func formatHealthFactor(healthFactor *big.Int) float64 {
	if healthFactor == nil {
		return 0
	}

	// 将健康因子转换为浮点数
	f, _ := new(big.Float).SetString(healthFactor.String())
	f.Quo(f, HF_DECIMALS)
	result, _ := f.Float64()
	return result
}

// 辅助方法
func formatAmount(amount, decimals *big.Int) string {
	if amount == nil {
		return "0"
	}

	// 将大整数转换为浮点数并考虑精度
	f, _ := new(big.Float).SetString(amount.String())
	decimalsFactor, _ := new(big.Float).SetString(new(big.Int).Exp(big.NewInt(10), decimals, nil).String())
	f.Quo(f, decimalsFactor)

	result, _ := f.Float64()
	return fmt.Sprintf("%.8f", result)
}

// base = amount / 10^decimals * price
func amountToBase(amount, decimals *big.Int, price *big.Int) *big.Int {
	if amount == nil || decimals == nil || price == nil {
		return big.NewInt(0)
	}

	// 计算 USD 价值
	amountFloat, _ := new(big.Float).SetString(amount.String())
	priceFloat, _ := new(big.Float).SetString(price.String())

	// 考虑精度
	decimalsFactor, _ := new(big.Float).SetString(new(big.Int).Exp(big.NewInt(10), decimals, nil).String())

	value := new(big.Float)
	value.Quo(amountFloat, decimalsFactor)
	value.Mul(value, priceFloat)
	base, _ := value.Int(nil)
	return base
}

// usd = amount / 10^decimals * price / 10^8
func amountToUSD(amount, decimals *big.Int, price *big.Int) float64 {
	base := amountToBase(amount, decimals, price)
	baseFloat, _ := new(big.Float).SetString(base.String())
	usd := big.NewFloat(0).Quo(baseFloat, USD_DECIMALS)
	usdFloat, _ := usd.Float64()
	return usdFloat
}

// amount = usd / price * 10^decimals
func baseToAmount(base, decimals, price *big.Int) *big.Int {
	if price == nil || price.Sign() == 0 {
		return big.NewInt(0)
	}

	// 将 USD 转换为代币数量
	baseFloat, _ := new(big.Float).SetString(base.String())
	priceFloat, _ := new(big.Float).SetString(price.String())
	decimalsFactor, _ := new(big.Float).SetString(new(big.Int).Exp(big.NewInt(10), decimals, nil).String())

	// 计算代币数量
	amount := new(big.Float)
	amount.Quo(baseFloat, priceFloat)
	amount.Mul(amount, decimalsFactor)

	// 转换为大整数
	result, _ := amount.Int(nil)
	return result
}

func checkUSDEqual(old, new *big.Int) bool {
	oldUSD, _ := big.NewFloat(0).SetString(old.String())
	newUSD, _ := big.NewFloat(0).SetString(new.String())
	oldUSD.Quo(oldUSD, USD_DECIMALS)
	newUSD.Quo(newUSD, USD_DECIMALS)
	if fmt.Sprintf("%0.2f", oldUSD) != fmt.Sprintf("%0.2f", newUSD) {
		fmt.Println("oldUSD", fmt.Sprintf("%0.2f", oldUSD), "newUSD", fmt.Sprintf("%0.2f", newUSD))
		return false
	}
	return true
}

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
