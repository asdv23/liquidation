package aavev3

import (
	"fmt"
	"math/big"
)

var (
	MIN_DEBT_USD = big.NewFloat(2)
	USD_DECIMALS = big.NewFloat(1e8)
)

func formatHealthFactor(healthFactor *big.Int) float64 {
	if healthFactor == nil {
		return 0
	}

	// 将健康因子转换为浮点数
	f := new(big.Float).SetUint64(healthFactor.Uint64())
	f.Quo(f, new(big.Float).SetUint64(1e18))
	result, _ := f.Float64()
	return result
}

func parseAmount(amount string, decimals uint8) (*big.Int, error) {
	// 解析金额
	f, _, err := new(big.Float).Parse(amount, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to parse amount: %w", err)
	}

	// 转换为整数
	scale := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	f.Mul(f, scale)

	// 转换为 big.Int
	result, _ := f.Int(nil)
	return result, nil
}

// 辅助方法
func formatAmount(amount, decimals *big.Int) string {
	if amount == nil {
		return "0"
	}

	// 将大整数转换为浮点数并考虑精度
	f := new(big.Float).SetInt(amount)
	f.Quo(f, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), decimals, nil)))

	result, _ := f.Float64()
	return fmt.Sprintf("%.8f", result)
}

func amountToUSD(amount, decimals *big.Int, price *big.Int) float64 {
	if amount == nil || price == nil {
		return 0
	}

	// 计算 USD 价值
	value := new(big.Float).SetInt(amount)
	priceFloat := new(big.Float).SetInt(price)
	priceFloat.Quo(priceFloat, USD_DECIMALS)

	// 考虑精度
	decimalsFactor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), decimals, nil))
	value.Quo(value, decimalsFactor)

	result := new(big.Float).Mul(value, priceFloat)
	usdValue, _ := result.Float64()
	return usdValue
}

func USDToAmount(usd float64, decimals int, price *big.Int) *big.Int {
	if price == nil || price.Sign() == 0 {
		return big.NewInt(0)
	}

	// 将 USD 转换为代币数量
	usdFloat := new(big.Float).SetFloat64(usd)
	priceFloat := new(big.Float).SetInt(price)

	// 计算代币数量
	amount := new(big.Float).Quo(usdFloat, priceFloat)

	// 考虑精度
	decimalsFactor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))
	amount.Mul(amount, decimalsFactor)

	// 转换为大整数
	result, _ := amount.Int(nil)
	return result
}
