package models

import (
	"time"

	"gorm.io/gorm"
)

type Reserve struct {
	ID uint `gorm:"primarykey"`
	// LoanID               uint `gorm:"foreignKey:LoanID;references:ID"`
	ChainName            string `gorm:"uniqueIndex:idx_chain_reserve,priority:1"`
	User                 string `gorm:"uniqueIndex:idx_chain_reserve,priority:2"`
	Reserve              string `gorm:"uniqueIndex:idx_chain_reserve,priority:3"`
	Amount               *BigInt
	AmountBase           *BigInt
	IsBorrowing          bool
	IsUsingAsCollateral  bool
	CurrentATokenBalance *BigInt
	CurrentVariableDebt  *BigInt
	CurrentStableDebt    *BigInt
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (t *Reserve) BeforeCreate(tx *gorm.DB) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now()
	}
	return nil
}

func (t *Reserve) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}
