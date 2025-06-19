package aavev3

import (
	"liquidation-bot/internal/models"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *DBWrapper {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(&models.Loan{})
	require.NoError(t, err)

	return &DBWrapper{db: db}
}

// createTestLoan 创建测试用的贷款数据
func createTestLoan(t *testing.T, db *DBWrapper, chainName, user string, healthFactor float64) {
	loan := &models.Loan{
		ChainName:    chainName,
		User:         user,
		IsActive:     true,
		HealthFactor: healthFactor,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		LiquidationInfo: &models.LiquidationInfo{
			TotalCollateralBase:  models.NewBigInt(big.NewInt(1000)),
			TotalDebtBase:        models.NewBigInt(big.NewInt(500)),
			LiquidationThreshold: models.NewBigInt(big.NewInt(800)),
			CollateralAsset:      "0x1234",
			CollateralAmount:     models.NewBigInt(big.NewInt(100)),
			CollateralAmountBase: models.NewBigInt(big.NewInt(1000)),
			DebtAsset:            "0x5678",
			DebtAmount:           models.NewBigInt(big.NewInt(50)),
			DebtAmountBase:       models.NewBigInt(big.NewInt(500)),
		},
	}
	err := db.db.Create(loan).Error
	require.NoError(t, err)
}

func TestBatchUpdateLoanLiquidationInfos(t *testing.T) {
	tests := []struct {
		name             string
		setupFunc        func(t *testing.T, db *DBWrapper)
		liquidationInfos []*UpdateLiquidationInfo
		expectedFunc     func(t *testing.T, db *DBWrapper)
		wantErr          bool
	}{
		{
			name: "成功更新多个用户",
			setupFunc: func(t *testing.T, db *DBWrapper) {
				createTestLoan(t, db, "chain1", "user1", 1.5)
				createTestLoan(t, db, "chain1", "user2", 1.2)
				createTestLoan(t, db, "chain1", "user3", 1.8)
			},
			liquidationInfos: []*UpdateLiquidationInfo{
				{
					User:         "user1",
					HealthFactor: 0.8,
					LiquidationInfo: &models.LiquidationInfo{
						TotalCollateralBase:  models.NewBigInt(big.NewInt(2000)),
						TotalDebtBase:        models.NewBigInt(big.NewInt(1000)),
						LiquidationThreshold: models.NewBigInt(big.NewInt(900)),
						CollateralAsset:      "0x1234",
						CollateralAmount:     models.NewBigInt(big.NewInt(200)),
						CollateralAmountBase: models.NewBigInt(big.NewInt(2000)),
						DebtAsset:            "0x5678",
						DebtAmount:           models.NewBigInt(big.NewInt(100)),
						DebtAmountBase:       models.NewBigInt(big.NewInt(1000)),
					},
				},
				{
					User:         "user2",
					HealthFactor: 0.9,
					LiquidationInfo: &models.LiquidationInfo{
						TotalCollateralBase:  models.NewBigInt(big.NewInt(3000)),
						TotalDebtBase:        models.NewBigInt(big.NewInt(1500)),
						LiquidationThreshold: models.NewBigInt(big.NewInt(850)),
						CollateralAsset:      "0x1234",
						CollateralAmount:     models.NewBigInt(big.NewInt(300)),
						CollateralAmountBase: models.NewBigInt(big.NewInt(3000)),
						DebtAsset:            "0x5678",
						DebtAmount:           models.NewBigInt(big.NewInt(150)),
						DebtAmountBase:       models.NewBigInt(big.NewInt(1500)),
					},
				},
			},
			expectedFunc: func(t *testing.T, db *DBWrapper) {
				// 验证 user1 的更新
				var loan1 models.Loan
				err := db.db.Where("chain_name = ? AND user = ?", "chain1", "user1").First(&loan1).Error
				require.NoError(t, err)
				assert.Equal(t, 0.8, loan1.HealthFactor)
				assert.Equal(t, "2000", loan1.LiquidationInfo.TotalCollateralBase.String())
				assert.Equal(t, "1000", loan1.LiquidationInfo.TotalDebtBase.String())

				// 验证 user2 的更新
				var loan2 models.Loan
				err = db.db.Where("chain_name = ? AND user = ?", "chain1", "user2").First(&loan2).Error
				require.NoError(t, err)
				assert.Equal(t, 0.9, loan2.HealthFactor)
				assert.Equal(t, "3000", loan2.LiquidationInfo.TotalCollateralBase.String())
				assert.Equal(t, "1500", loan2.LiquidationInfo.TotalDebtBase.String())

				// 验证 user3 未被更新
				var loan3 models.Loan
				err = db.db.Where("chain_name = ? AND user = ?", "chain1", "user3").First(&loan3).Error
				require.NoError(t, err)
				assert.Equal(t, 1.8, loan3.HealthFactor)
			},
			wantErr: false,
		},
		{
			name: "空列表不报错",
			setupFunc: func(t *testing.T, db *DBWrapper) {
				createTestLoan(t, db, "chain1", "user1", 1.5)
			},
			liquidationInfos: []*UpdateLiquidationInfo{},
			expectedFunc: func(t *testing.T, db *DBWrapper) {
				var loan models.Loan
				err := db.db.Where("chain_name = ? AND user = ?", "chain1", "user1").First(&loan).Error
				require.NoError(t, err)
				assert.Equal(t, 1.5, loan.HealthFactor)
			},
			wantErr: false,
		},
		{
			name: "更新不存在的用户",
			setupFunc: func(t *testing.T, db *DBWrapper) {
				createTestLoan(t, db, "chain1", "user1", 1.5)
			},
			liquidationInfos: []*UpdateLiquidationInfo{
				{
					User:         "nonexistent",
					HealthFactor: 0.8,
					LiquidationInfo: &models.LiquidationInfo{
						TotalCollateralBase:  models.NewBigInt(big.NewInt(2000)),
						TotalDebtBase:        models.NewBigInt(big.NewInt(1000)),
						LiquidationThreshold: models.NewBigInt(big.NewInt(900)),
						CollateralAsset:      "0x1234",
						CollateralAmount:     models.NewBigInt(big.NewInt(200)),
						CollateralAmountBase: models.NewBigInt(big.NewInt(2000)),
						DebtAsset:            "0x5678",
						DebtAmount:           models.NewBigInt(big.NewInt(100)),
						DebtAmountBase:       models.NewBigInt(big.NewInt(1000)),
					},
				},
			},
			expectedFunc: func(t *testing.T, db *DBWrapper) {
				// 验证现有用户未受影响
				var loan models.Loan
				err := db.db.Where("chain_name = ? AND user = ?", "chain1", "user1").First(&loan).Error
				require.NoError(t, err)
				assert.Equal(t, 1.5, loan.HealthFactor)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			tt.setupFunc(t, db)

			err := db.BatchUpdateLoanLiquidationInfos("chain1", tt.liquidationInfos)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			tt.expectedFunc(t, db)
		})
	}
}
