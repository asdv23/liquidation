package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&Loan{})
	assert.NoError(t, err)

	return db
}

func TestLoanBeforeCreate(t *testing.T) {
	db := setupTestDB(t)

	loan := &Loan{
		ChainName: "ethereum",
		User:      "0x123",
		IsActive:  true,
	}

	err := db.Create(loan).Error
	assert.NoError(t, err)
	assert.False(t, loan.CreatedAt.IsZero())
	assert.False(t, loan.UpdatedAt.IsZero())
}

func TestLoanBeforeUpdate(t *testing.T) {
	db := setupTestDB(t)

	loan := &Loan{
		ChainName: "ethereum",
		User:      "0x123",
		IsActive:  true,
	}

	err := db.Create(loan).Error
	assert.NoError(t, err)

	oldUpdatedAt := loan.UpdatedAt
	time.Sleep(time.Second)

	loan.IsActive = false
	err = db.Save(loan).Error
	assert.NoError(t, err)
	assert.True(t, loan.UpdatedAt.After(oldUpdatedAt))
}

func TestLoanUniqueConstraint(t *testing.T) {
	db := setupTestDB(t)

	loan1 := &Loan{
		ChainName: "ethereum",
		User:      "0x123",
		IsActive:  true,
	}

	err := db.Create(loan1).Error
	assert.NoError(t, err)

	loan2 := &Loan{
		ChainName: "ethereum",
		User:      "0x123",
		IsActive:  true,
	}

	err = db.Create(loan2).Error
	assert.Error(t, err) // 应该违反唯一约束
}
