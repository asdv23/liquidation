package models

import (
	"time"

	"gorm.io/gorm"
)

type Loan struct {
	ID                      uint   `gorm:"primarykey"`
	ChainName               string `gorm:"index:idx_chain_active,priority:1"`
	User                    string `gorm:"uniqueIndex:idx_chain_user"`
	IsActive                bool   `gorm:"default:true;index:idx_chain_active,priority:2"`
	CreatedAt               time.Time
	UpdatedAt               time.Time
	NextCheckTime           *time.Time
	HealthFactor            float64
	LiquidationDiscoveredAt *time.Time
	LiquidationTxHash       *string
	LiquidationTime         *time.Time
	Liquidator              *string
	LiquidationDelay        *int64
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
