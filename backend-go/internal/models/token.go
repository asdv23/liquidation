package models

import (
	"time"

	"gorm.io/gorm"
)

type Token struct {
	ID        uint   `gorm:"primarykey"`
	ChainName string `gorm:"uniqueIndex:idx_chain_address"`
	Address   string `gorm:"uniqueIndex:idx_chain_address"`
	Symbol    string
	Decimals  *BigInt
	Price     *BigInt
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Token) BeforeCreate(tx *gorm.DB) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = time.Now()
	}
	return nil
}

func (t *Token) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}
