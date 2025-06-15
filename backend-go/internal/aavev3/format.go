package aavev3

import (
	"fmt"
	"math/big"
)

var (
	MIN_DEBT_USD = big.NewFloat(2)
	USD_DECIMALS = big.NewFloat(1e8)
	HF_DECIMALS  = big.NewFloat(1e18)
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

func amountToUSD(amount, decimals *big.Int, price *big.Int) float64 {
	if amount == nil || price == nil {
		return 0
	}

	// 计算 USD 价值
	priceFloat, _ := new(big.Float).SetString(price.String())
	priceFloat.Quo(priceFloat, USD_DECIMALS)

	// 考虑精度
	decimalsFactor, _ := new(big.Float).SetString(new(big.Int).Exp(big.NewInt(10), decimals, nil).String())

	value, _ := new(big.Float).SetString(amount.String())
	value.Quo(value, decimalsFactor)
	value.Mul(value, priceFloat)
	usdValue, _ := value.Float64()
	return usdValue
}

func USDToAmount(usd float64, decimals, price *big.Int) *big.Int {
	if price == nil || price.Sign() == 0 {
		return big.NewInt(0)
	}

	// 将 USD 转换为代币数量
	usdFloat, _ := new(big.Float).SetString(fmt.Sprintf("%f", usd))
	priceFloat, _ := new(big.Float).SetString(price.String())

	// 计算代币数量
	amount := new(big.Float).Quo(usdFloat, priceFloat)

	// 考虑精度
	decimalsFactor, _ := new(big.Float).SetString(new(big.Int).Exp(big.NewInt(10), decimals, nil).String())
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
