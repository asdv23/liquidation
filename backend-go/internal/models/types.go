package models

import (
	"database/sql/driver"
	"math/big"
)

// BigInt 是一个包装了 big.Int 的类型，实现了 GORM 的 Valuer/Scanner 接口
type BigInt big.Int

func NewBigInt(i *big.Int) *BigInt {
	return (*BigInt)(i)
}

func (b *BigInt) String() string {
	if b == nil {
		return "0"
	}
	return b.BigInt().String()
}

func (b *BigInt) BigInt() *big.Int {
	if b == nil {
		return nil
	}
	return (*big.Int)(b)
}

// b = b + other
func (b *BigInt) Add(other *BigInt) *BigInt {
	if b == nil || other == nil {
		return NewBigInt(big.NewInt(0))
	}
	return NewBigInt(b.BigInt().Add(b.BigInt(), other.BigInt()))
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
