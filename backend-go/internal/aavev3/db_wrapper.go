package aavev3

import (
	"context"
	"fmt"
	"liquidation-bot/internal/models"
	"math/big"
	"sync"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type DBWrapper struct {
	sync.RWMutex

	db           *gorm.DB
	tokens       map[string]map[string]*models.Token
	activeLoans  map[string]map[string]*models.Loan
	userReserves map[string]map[string][]*models.Reserve
}

func NewDBWrapper(db *gorm.DB) (*DBWrapper, error) {
	w := &DBWrapper{db: db, tokens: make(map[string]map[string]*models.Token), activeLoans: make(map[string]map[string]*models.Loan)}
	eg, _ := errgroup.WithContext(context.Background())
	eg.Go(w.loadTokens)
	eg.Go(w.loadActiveLoans)
	eg.Go(w.loadUserReserves)
	return w, eg.Wait()
}

func (w *DBWrapper) loadTokens() error {
	var tokens []models.Token
	if err := w.db.Find(&tokens).Error; err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	for _, token := range tokens {
		if w.tokens[token.ChainName] == nil {
			w.tokens[token.ChainName] = make(map[string]*models.Token)
		}
		w.tokens[token.ChainName][token.Address] = &token
	}
	return nil
}

func (w *DBWrapper) loadActiveLoans() error {
	var loans []models.Loan
	if err := w.db.Where("is_active = ?", true).Find(&loans).Error; err != nil {
		return fmt.Errorf("failed to load active loans: %w", err)
	}

	for _, loan := range loans {
		if w.activeLoans[loan.ChainName] == nil {
			w.activeLoans[loan.ChainName] = make(map[string]*models.Loan)
		}
		w.activeLoans[loan.ChainName][loan.User] = &loan
	}
	return nil
}

func (w *DBWrapper) loadUserReserves() error {
	var reserves []models.Reserve
	if err := w.db.Find(&reserves).Error; err != nil {
		return fmt.Errorf("failed to load user reserves: %w", err)
	}

	for _, reserve := range reserves {
		if w.userReserves[reserve.ChainName] == nil {
			w.userReserves[reserve.ChainName] = make(map[string][]*models.Reserve)
		}
		w.userReserves[reserve.ChainName][reserve.User] = append(w.userReserves[reserve.ChainName][reserve.User], &reserve)
	}
	return nil
}

func (w *DBWrapper) GetDB() *gorm.DB {
	return w.db
}

func (w *DBWrapper) GetTokenInfo(chainName string, address string) (*models.Token, error) {
	w.RLock()
	defer w.RUnlock()

	token, ok := w.tokens[chainName][address]
	if !ok {
		return nil, fmt.Errorf("token not found")
	}
	return token, nil
}

func (w *DBWrapper) AddTokenInfo(chainName string, address string, symbol string, decimals, price *big.Int) (*models.Token, error) {
	w.Lock()
	defer w.Unlock()

	token := &models.Token{
		ChainName: chainName,
		Address:   address,
		Symbol:    symbol,
		Decimals:  int(decimals.Int64()),
		Price:     (*models.BigInt)(price),
	}
	if err := w.db.Where(&models.Token{ChainName: chainName, Address: address}).
		Assign(token).
		FirstOrCreate(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to add token info: %w", err)
	}
	if w.tokens[chainName] == nil {
		w.tokens[chainName] = make(map[string]*models.Token)
	}
	w.tokens[chainName][address] = token

	return token, nil
}

func (w *DBWrapper) GetActiveLoans(chainName string) (map[string]*models.Loan, bool) {
	w.RLock()
	defer w.RUnlock()

	loans, ok := w.activeLoans[chainName]
	return loans, ok
}

func (w *DBWrapper) CreateOrUpdateActiveLoan(chainName string, user string) error {
	w.Lock()
	defer w.Unlock()

	var loan models.Loan
	if err := w.db.Where(&models.Loan{ChainName: chainName, User: user}).
		Assign(models.Loan{IsActive: true}).
		FirstOrCreate(&loan).Error; err != nil {
		return fmt.Errorf("failed to upsert active loan: %w", err)
	}
	if w.activeLoans[chainName] == nil {
		w.activeLoans[chainName] = make(map[string]*models.Loan)
	}
	w.activeLoans[chainName][user] = &loan

	return nil
}

func (w *DBWrapper) UpdateActiveLoanHealthFactor(chainName string, user string, healthFactor float64) error {
	w.Lock()
	defer w.Unlock()

	if err := w.db.Model(&models.Loan{}).Where(&models.Loan{ChainName: chainName, User: user}).Update("health_factor", healthFactor).Error; err != nil {
		return fmt.Errorf("failed to update active loan: %w", err)
	}
	w.activeLoans[chainName][user].HealthFactor = healthFactor

	return nil
}

func (w *DBWrapper) UpdateActiveLoanLiquidationInfo(chainName, user string, liquidationInfo *models.LiquidationInfo) error {
	w.Lock()
	defer w.Unlock()

	loan, ok := w.activeLoans[chainName][user]
	if !ok {
		return fmt.Errorf("active loan not found")
	}
	loan.LiquidationInfo = liquidationInfo

	if err := w.db.Save(loan).Error; err != nil {
		return fmt.Errorf("failed to save active loan: %w", err)
	}

	return nil
}

func (w *DBWrapper) GetUserReserves(chainName string, user string) ([]*models.Reserve, bool) {
	w.RLock()
	defer w.RUnlock()

	reserves, ok := w.userReserves[chainName][user]
	return reserves, ok
}

func (w *DBWrapper) GetLiquidationInfo(chainName string, user string) (*models.LiquidationInfo, error) {
	w.RLock()
	defer w.RUnlock()

	loan, ok := w.activeLoans[chainName][user]
	if !ok {
		return &models.LiquidationInfo{}, nil
	}
	return loan.LiquidationInfo, nil
}
