package aavev3

import (
	"fmt"
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
	return calcHealthFactor, fmt.Sprintf("%0.5f", calcHealthFactor) == fmt.Sprintf("%0.5f", healthFactor)
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

// LiquidationPair 清算对
type LiquidationPair struct {
	User            string
	CollateralAsset string
	DebtAsset       string
	DebtAmount      *big.Int
	Profit          *big.Int
}

// LiquidationParams 清算参数
type LiquidationParams struct {
	CollateralAmount *big.Int
	DebtAmount       *big.Int
	CollateralPrice  *big.Int
	DebtPrice        *big.Int
}

// UserAsset 用户资产
type UserAsset struct {
	Address      string
	IsCollateral bool
	Price        *big.Int
	Amount       *big.Int
	Decimals     uint8
}

// contracts/lib/aave-v3-origin/src/contracts/protocol/libraries/types/DataTypes.sol
// bit 0-15: LTV
// bit 16-31: Liq. threshold
// bit 32-47: Liq. bonus
// bit 48-55: Decimals
// bit 56: reserve is active
// bit 57: reserve is frozen
// bit 58: borrowing is enabled
// bit 59: DEPRECATED: stable rate borrowing enabled
// bit 60: asset is paused
// bit 61: borrowing in isolation mode is enabled
// bit 62: siloed borrowing enabled
// bit 63: flashloaning enabled
// bit 64-79: reserve factor
// bit 80-115: borrow cap in whole tokens, borrowCap == 0 => no cap
// bit 116-151: supply cap in whole tokens, supplyCap == 0 => no cap
// bit 152-167: liquidation protocol fee
// bit 168-175: DEPRECATED: eMode category
// bit 176-211: unbacked mint cap in whole tokens, unbackedMintCap == 0 => minting disabled
// bit 212-251: debt ceiling for isolation mode with (ReserveConfiguration::DEBT_CEILING_DECIMALS) decimals
// bit 252: virtual accounting is enabled for the reserve
// bit 253-255 unused
//
// ReserveConfiguration 储备配置
type ReserveConfiguration struct {
	Data *big.Int
}
