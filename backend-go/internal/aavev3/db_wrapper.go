package aavev3

import (
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

// chainName -> address -> token
func (w *DBWrapper) loadTokens() (map[string]map[string]*models.Token, error) {
	var tokens []models.Token
	if err := w.db.Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to load tokens: %w", err)
	}

	tokensMap := make(map[string]map[string]*models.Token)
	for _, token := range tokens {
		if tokensMap[token.ChainName] == nil {
			tokensMap[token.ChainName] = make(map[string]*models.Token)
		}
		tokensMap[token.ChainName][token.Address] = &token
	}
	return tokensMap, nil
}

// chainName -> user -> loan
func (w *DBWrapper) loadActiveLoans() (map[string]map[string]*models.Loan, error) {
	var loans []models.Loan
	if err := w.db.Where("is_active = ?", true).Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to load active loans: %w", err)
	}

	activeLoansMap := make(map[string]map[string]*models.Loan)
	for _, loan := range loans {
		if activeLoansMap[loan.ChainName] == nil {
			activeLoansMap[loan.ChainName] = make(map[string]*models.Loan)
		}
		activeLoansMap[loan.ChainName][loan.User] = &loan
	}
	return activeLoansMap, nil
}

// chainName -> user -> reserves
func (w *DBWrapper) loadUserReserves() (map[string]map[string][]*models.Reserve, error) {
	var reserves []models.Reserve
	if err := w.db.Find(&reserves).Error; err != nil {
		return nil, fmt.Errorf("failed to load user reserves: %w", err)
	}

	userReservesMap := make(map[string]map[string][]*models.Reserve)
	for _, reserve := range reserves {
		if userReservesMap[reserve.ChainName] == nil {
			userReservesMap[reserve.ChainName] = make(map[string][]*models.Reserve)
		}
		userReservesMap[reserve.ChainName][reserve.User] = append(userReservesMap[reserve.ChainName][reserve.User], &reserve)
	}
	return userReservesMap, nil
}

func (w *DBWrapper) GetDB() *gorm.DB {
	return w.db
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

// user -> loan
func (w *DBWrapper) ChainActiveLoans(chainName string) (map[string]*models.Loan, error) {
	loans := make([]*models.Loan, 0)
	if err := w.db.Where(&models.Loan{ChainName: chainName}).Find(&loans).Error; err != nil {
		return nil, fmt.Errorf("failed to get active loans: %w", err)
	}
	activeLoansMap := make(map[string]*models.Loan)
	for _, loan := range loans {
		activeLoansMap[loan.User] = loan
	}
	return activeLoansMap, nil
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

func (w *DBWrapper) UpdateActiveLoanHealthFactor(chainName string, user string, healthFactor float64) error {
	if err := w.db.Model(&models.Loan{}).Where(&models.Loan{ChainName: chainName, User: user}).Update("health_factor", healthFactor).Error; err != nil {
		return fmt.Errorf("failed to update active loan: %w", err)
	}
	return nil
}

func (w *DBWrapper) DeactivateActiveLoan(chainName string, user string) error {
	if err := w.db.Model(&models.Loan{}).Where(&models.Loan{ChainName: chainName, User: user}).Update("is_active", false).Error; err != nil {
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
