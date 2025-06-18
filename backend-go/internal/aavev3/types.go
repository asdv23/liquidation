package aavev3

import (
	"fmt"
	"liquidation-bot/internal/models"
	"math/big"
	"time"
)

// uint256 totalCollateralBase,
// uint256 totalDebtBase,
// uint256 availableBorrowsBase,
// uint256 currentLiquidationThreshold,
// uint256 ltv,
// uint256 healthFactor
//
// UserAccountData 用户账户数据
type UserAccountData struct {
	TotalCollateralBase         *big.Int
	TotalDebtBase               *big.Int
	AvailableBorrowsBase        *big.Int
	CurrentLiquidationThreshold *big.Int
	Ltv                         *big.Int
	HealthFactor                *big.Int
}

// (vars.totalCollateralInBaseCurrency.percentMul(vars.avgLiquidationThreshold)).wadDiv(
//
//	    vars.totalDebtInBaseCurrencyvars.healthFactor = (vars.totalDebtInBaseCurrency == 0)
//		? type(uint256).max
//		: (vars.totalCollateralInBaseCurrency.percentMul(vars.avgLiquidationThreshold)).wadDiv(
//		  vars.totalDebtInBaseCurrency
//		);
//
// 计算手算的健康因子和合约里是否一致
func (uad *UserAccountData) checkCalcHealthFactor(healthFactor float64) (float64, bool) {
	x := new(big.Int)
	calcHealthFactor := formatHealthFactor(x.Lsh(big.NewInt(1), 256).Sub(x, big.NewInt(1)))
	if uad.TotalDebtBase.Sign() != 0 {
		y := new(big.Int)
		y = y.Mul(uad.TotalCollateralBase, uad.CurrentLiquidationThreshold).Mul(y, big.NewInt(1e14)).Div(y, uad.TotalDebtBase)
		calcHealthFactor = formatHealthFactor(y)
	}
	if fmt.Sprintf("%0.2f", calcHealthFactor) != fmt.Sprintf("%0.2f", healthFactor) {
		fmt.Println("calcHealthFactor", fmt.Sprintf("%0.2f", calcHealthFactor), "healthFactor", fmt.Sprintf("%0.2f", healthFactor))
		return calcHealthFactor, false
	}
	return calcHealthFactor, true
}

// uint256 currentATokenBalance,
// uint256 currentStableDebt,
// uint256 currentVariableDebt,
// uint256 principalStableDebt,
// uint256 scaledVariableDebt,
// uint256 stableBorrowRate,
// uint256 liquidityRate,
// uint40 stableRateLastUpdated,
// bool usageAsCollateralEnabled
//
// UserReserveData 用户储备数据
type UserReserveData struct {
	CurrentATokenBalance     *big.Int
	CurrentVariableDebt      *big.Int
	CurrentStableDebt        *big.Int
	PrincipalStableDebt      *big.Int
	ScaledVariableDebt       *big.Int
	StableBorrowRate         *big.Int
	LiquidityRate            *big.Int
	StableRateLastUpdated    *big.Int
	UsageAsCollateralEnabled bool
}

// LiquidationInfo 清算信息
type LiquidationInfo struct {
	User             string
	HealthFactor     float64
	LastUpdated      *time.Time
	CollateralAssets []string
	DebtAssets       []string
	CollateralPrices map[string]*big.Int
	DebtPrices       map[string]*big.Int
}

type ReserveInfo struct {
	Address  string
	Decimals *big.Int
	Symbol   string
}

type UpdateLiquidationInfo struct {
	User            string
	HealthFactor    float64
	LiquidationInfo *models.LiquidationInfo
}

type InputToken struct {
	TokenAddress string `json:"tokenAddress"`
	Amount       string `json:"amount"`
}

type OutputToken struct {
	TokenAddress string `json:"tokenAddress"`
	Proportion   string `json:"proportion"`
}

type QuotePayload struct {
	ChainID              string        `json:"chainId"`
	InputTokens          []InputToken  `json:"inputTokens"`
	OutputTokens         []OutputToken `json:"outputTokens"`
	UserAddr             string        `json:"userAddr"`
	SlippageLimitPercent string        `json:"slippageLimitPercent"`
	PathViz              string        `json:"pathViz"`
	PathVizImage         string        `json:"pathVizImage"`
}

type QuoteResponse struct {
	PathID string `json:"pathId"`
}

type AssemblePayload struct {
	UserAddr string `json:"userAddr"`
	PathID   string `json:"pathId"`
	Simulate bool   `json:"simulate"`
	Receiver string `json:"receiver"`
}

type AssembleResponse struct {
	Transaction *transaction `json:"transaction"`
}
type transaction struct {
	Value string `json:"value"`
	To    string `json:"to"`
	Data  string `json:"data"`
}
