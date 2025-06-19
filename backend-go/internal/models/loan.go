package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Loan struct {
	ID                uint   `gorm:"primarykey"`
	ChainName         string `gorm:"uniqueIndex:idx_chain_user,priority:1;index:idx_chain_active,priority:1"`
	User              string `gorm:"uniqueIndex:idx_chain_user,priority:2"`
	IsActive          bool   `gorm:"default:true;index:idx_chain_active,priority:2"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	HealthFactor      float64
	LiquidationInfo   *LiquidationInfo `gorm:"embedded;embeddedPrefix:liquidation_"`
	LiquidationTxHash *string
	LiquidationTime   *time.Time
	Liquidator        *string
	LiquidationDelay  *int64
}

// LiquidationInfo 清算信息
type LiquidationInfo struct {
	TotalCollateralBase  *BigInt
	TotalDebtBase        *BigInt
	LiquidationThreshold *BigInt // should in outside together with health factor
	CollateralAsset      string
	CollateralAmount     *BigInt
	CollateralAmountBase *BigInt
	DebtAsset            string
	DebtAmount           *BigInt
	DebtAmountBase       *BigInt
}

func (l *LiquidationInfo) String() string {
	return fmt.Sprintf("TotalCollateralBase: %s, TotalDebtBase: %s, LiquidationThreshold: %s, CollateralAsset: %s, CollateralAmount: %s, CollateralAmountBase: %s, DebtAsset: %s, DebtAmount: %s, DebtAmountBase: %s",
		l.TotalCollateralBase.String(), l.TotalDebtBase.String(), l.LiquidationThreshold.String(), l.CollateralAsset, l.CollateralAmount.String(), l.CollateralAmountBase.String(), l.DebtAsset, l.DebtAmount.String(), l.DebtAmountBase.String())
}

func (old *LiquidationInfo) Cmp(new *LiquidationInfo) bool {
	return old.DebtAsset == new.DebtAsset &&
		old.DebtAmount == new.DebtAmount &&
		old.CollateralAsset == new.CollateralAsset &&
		old.CollateralAmount == new.CollateralAmount
}

func (l *Loan) BeforeCreate(tx *gorm.DB) error {
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now()
	}
	if l.UpdatedAt.IsZero() {
		l.UpdatedAt = time.Now()
	}
	return nil
}

func (l *Loan) BeforeUpdate(tx *gorm.DB) error {
	l.UpdatedAt = time.Now()
	return nil
}
