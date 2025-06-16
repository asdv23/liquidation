package aavev3

import (
	"context"
	"fmt"
	"liquidation-bot/internal/models"
	"math/big"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBWrapper struct {
	db *gorm.DB
}

func NewDBWrapper(db *gorm.DB) (*DBWrapper, error) {
	w := &DBWrapper{db: db}
	return w, nil
}

func (w *DBWrapper) GetDB() *gorm.DB {
	return w.db
}

func (w *DBWrapper) GetTokenInfoMap(chainName string) (map[string]*models.Token, error) {
	token := make([]*models.Token, 0)
	if err := w.db.Where(&models.Token{ChainName: chainName}).Find(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to get token infos: %w", err)
	}
	tokenMap := make(map[string]*models.Token, 0)
	for _, t := range token {
		tokenMap[t.Address] = t
	}
	return tokenMap, nil
}

func (w *DBWrapper) GetTokenInfo(chainName string, address string) (*models.Token, error) {
	token := &models.Token{}
	if err := w.db.Where(&models.Token{ChainName: chainName, Address: address}).First(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to get token info: %w", err)
	}
	return token, nil
}

func (w *DBWrapper) AddTokenInfo(chainName string, address string, symbol string, decimals, price *big.Int) (*models.Token, error) {
	token := &models.Token{
		ChainName: chainName,
		Address:   address,
		Symbol:    symbol,
		Decimals:  (*models.BigInt)(decimals),
		Price:     (*models.BigInt)(price),
	}
	if err := w.db.Where(&models.Token{ChainName: chainName, Address: address}).
		Assign(token).
		FirstOrCreate(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to add token info: %w", err)
	}

	return token, nil
}

func (w *DBWrapper) UpdateTokenPrice(chainName string, address string, price *models.BigInt) error {
	if err := w.db.Model(&models.Token{}).Where(&models.Token{ChainName: chainName, Address: address}).
		Where("price <> ?", price).
		Update("price", price).Error; err != nil {
		return fmt.Errorf("failed to update token price: %w", err)
	}
	return nil
}

// user -> loan
func (w *DBWrapper) ChainActiveLoans(chainName string) ([]*models.Loan, error) {
	loans := make([]*models.Loan, 0)
	if err := w.db.Where(&models.Loan{ChainName: chainName, IsActive: true}).Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loans: %w", err)
	}
	return loans, nil
}

func (w *DBWrapper) GetActiveLoansByToken(chainName string, tokenAddress string) ([]*models.Loan, error) {
	loans := make([]*models.Loan, 0)

	reserves := make([]*models.Reserve, 0)
	if err := w.db.Model(&models.Reserve{}).Where("chain_name = ? AND reserve = ?", chainName, tokenAddress).
		Where("is_using_as_collateral = ? OR is_borrowing = ?", true, true).
		Find(&reserves).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loans by token: %w", err)
	}

	userMap := make(map[string]struct{}, 0)
	for _, reserve := range reserves {
		userMap[reserve.User] = struct{}{}
	}

	users := make([]string, 0)
	for user := range userMap {
		users = append(users, user)
	}

	if err := w.db.Where(&models.Loan{ChainName: chainName, IsActive: true}).Where("user IN (?)", users).
		Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loans by token: %w", err)
	}
	return loans, nil
}

func (w *DBWrapper) GetLiquidationLoans(ctx context.Context, chainName string) ([]*models.Loan, error) {
	loans := make([]*models.Loan, 0)
	if err := w.db.Where(&models.Loan{ChainName: chainName, IsActive: true}).Where(
		"health_factor < ? AND health_factor > 0", 1,
	).Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to get liquidation loans: %w", err)
	}
	return loans, nil
}

func (w *DBWrapper) GetNullLiquidationLoans(ctx context.Context, chainName string) ([]*models.Loan, error) {
	loans := make([]*models.Loan, 0)
	if err := w.db.Where(&models.Loan{ChainName: chainName, IsActive: true}).Where(
		"liquidation_collateral_asset IS NULL",
	).Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to get liquidation loans: %w", err)
	}
	return loans, nil
}

func (w *DBWrapper) GetLoan(ctx context.Context, chainName, user string) (*models.Loan, error) {
	loan := &models.Loan{}
	if err := w.db.Where(&models.Loan{ChainName: chainName, User: user}).First(&loan).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loan: %w", err)
	}
	return loan, nil
}

func (w *DBWrapper) CreateOrUpdateActiveLoan(chainName string, user string) (*models.Loan, error) {
	loan := models.Loan{
		ChainName: chainName,
		User:      user,
		IsActive:  true,
	}
	if err := w.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "chain_name"}, {Name: "user"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"is_active",
		}),
	}).Create(&loan).Error; err != nil {
		return nil, fmt.Errorf("failed to upsert active loan: %w", err)
	}
	return &loan, nil
}

func (w *DBWrapper) UpdateActiveLoanLiquidationInfos(chainName string, liquidationInfos []*UpdateLiquidationInfo) error {
	// batch update
	for i := range liquidationInfos {
		loan := models.Loan{
			ChainName:       chainName,
			User:            liquidationInfos[i].User,
			HealthFactor:    liquidationInfos[i].HealthFactor,
			LiquidationInfo: liquidationInfos[i].LiquidationInfo,
			IsActive:        true,
		}
		if err := w.db.Model(&models.Loan{}).Where("chain_name = ? AND user = ?", chainName, liquidationInfos[i].User).Updates(loan).Error; err != nil {
			return fmt.Errorf("failed to update active loan health factor: %w", err)
		}
	}
	return nil
}

func (w *DBWrapper) DeactivateActiveLoan(chainName string, user []string) error {
	if err := w.db.Model(&models.Loan{}).Where("chain_name = ? AND user IN (?)", chainName, user).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate active loan: %w", err)
	}
	return nil
}

func (w *DBWrapper) UpdateActiveLoan(chainName, user string, loan *models.Loan) error {
	if err := w.db.Model(&models.Loan{}).Where(&models.Loan{ChainName: chainName, User: user}).Save(loan).Error; err != nil {
		return fmt.Errorf("failed to save active loan: %w", err)
	}

	return nil
}

func (w *DBWrapper) GetUserReserves(chainName string, user string) ([]*models.Reserve, error) {
	reserves := make([]*models.Reserve, 0)
	if err := w.db.Where(&models.Reserve{ChainName: chainName, User: user}).Find(&reserves).Error; err != nil {
		return nil, fmt.Errorf("failed to get user reserves: %w", err)
	}
	return reserves, nil
}

func (w *DBWrapper) GetLiquidationInfo(chainName string, user string) (*models.LiquidationInfo, error) {
	loan := &models.Loan{}
	if err := w.db.Where(&models.Loan{ChainName: chainName, User: user}).First(&loan).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loan: %w", err)
	}
	return loan.LiquidationInfo, nil
}

func (w *DBWrapper) AddUserReserves(chainName string, user string, reserves []*models.Reserve) error {
	// upsert
	for _, reserve := range reserves {
		if err := w.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "chain_name"}, {Name: "user"}, {Name: "reserve"}},
			DoUpdates: clause.AssignmentColumns([]string{
				`amount`, `amount_base`, `is_borrowing`, `is_using_as_collateral`,
			}),
		}).Create(&reserve).Error; err != nil {
			return fmt.Errorf("failed to upsert user reserve: %w", err)
		}
	}

	return nil
}
