package aavev3

import (
	aavev3 "liquidation-bot/bindings/aavev3"
	"math/big"
	"testing"
)

func TestUserConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		data           *big.Int
		reserveIndex   int
		wantBorrowing  bool
		wantCollateral bool
	}{
		{
			name:           "测试 reserveIndex 0",
			data:           big.NewInt(3), // 二进制: 11
			reserveIndex:   0,
			wantBorrowing:  true,
			wantCollateral: true,
		},
		{
			name:           "测试 reserveIndex 1",
			data:           big.NewInt(12), // 二进制: 1100
			reserveIndex:   1,
			wantBorrowing:  true,
			wantCollateral: true,
		},
		{
			name:           "测试只有 borrowing",
			data:           big.NewInt(1), // 二进制: 01
			reserveIndex:   0,
			wantBorrowing:  true,
			wantCollateral: false,
		},
		{
			name:           "测试只有 collateral",
			data:           big.NewInt(2), // 二进制: 10
			reserveIndex:   0,
			wantBorrowing:  false,
			wantCollateral: true,
		},
		{
			name:           "测试都没有",
			data:           big.NewInt(0), // 二进制: 00
			reserveIndex:   0,
			wantBorrowing:  false,
			wantCollateral: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &aavev3.DataTypesUserConfigurationMap{
				Data: tt.data,
			}

			// 测试 isBorrowing
			if got := isBorrowing(config, tt.reserveIndex); got != tt.wantBorrowing {
				t.Errorf("isBorrowing() = %v, want %v", got, tt.wantBorrowing)
			}

			// 测试 isUsingAsCollateral
			if got := isUsingAsCollateral(config, tt.reserveIndex); got != tt.wantCollateral {
				t.Errorf("isUsingAsCollateral() = %v, want %v", got, tt.wantCollateral)
			}
		})
	}
}

// 测试边界情况
func TestUserConfigurationEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		reserveIndex   int
		wantBorrowing  bool
		wantCollateral bool
	}{
		{
			name:           "测试超出范围的 reserveIndex",
			reserveIndex:   256, // 超出 uint8 范围
			wantBorrowing:  false,
			wantCollateral: false,
		},
		{
			name:           "测试负数 reserveIndex",
			reserveIndex:   -1,
			wantBorrowing:  false,
			wantCollateral: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &aavev3.DataTypesUserConfigurationMap{
				Data: big.NewInt(0),
			}

			// 测试 isBorrowing
			if got := isBorrowing(config, tt.reserveIndex); got != tt.wantBorrowing {
				t.Errorf("isBorrowing() = %v, want %v", got, tt.wantBorrowing)
			}

			// 测试 isUsingAsCollateral
			if got := isUsingAsCollateral(config, tt.reserveIndex); got != tt.wantCollateral {
				t.Errorf("isUsingAsCollateral() = %v, want %v", got, tt.wantCollateral)
			}
		})
	}
}
