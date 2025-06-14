package models

import (
	"database/sql/driver"
	"math/big"
	"time"

	"gorm.io/gorm"
)

// BigInt 是一个包装了 big.Int 的类型，实现了 GORM 的 Valuer/Scanner 接口
type BigInt big.Int

func NewBigInt(i *big.Int) *BigInt {
	return (*BigInt)(i)
}

func (b *BigInt) BigInt() *big.Int {
	if b == nil {
		return nil
	}
	return (*big.Int)(b)
}

// Value 实现了 driver.Valuer 接口
func (b *BigInt) Value() (driver.Value, error) {
	if b == nil {
		return nil, nil
	}
	return (*big.Int)(b).String(), nil
}

// Scan 实现了 sql.Scanner 接口
func (b *BigInt) Scan(value interface{}) error {
	if value == nil {
		*b = BigInt{}
		return nil
	}

	var i big.Int
	_, ok := i.SetString(value.(string), 10)
	if !ok {
		return nil
	}
	*b = BigInt(i)
	return nil
}

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
	LiquidationThreshold *BigInt
	CollateralAsset      string
	CollateralAmount     *BigInt
	CollateralAmountBase float64
	DebtAsset            string
	DebtAmount           *BigInt
	DebtAmountBase       float64
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
