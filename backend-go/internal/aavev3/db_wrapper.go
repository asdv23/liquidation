package aavev3

import (
	"fmt"
	"liquidation-bot/internal/models"
	"sync"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type DBWrapper struct {
	sync.RWMutex

	db          *gorm.DB
	tokenCache  map[string]map[string]*models.Token
	activeLoans map[string]map[string]*models.Loan
}

func NewDBWrapper(db *gorm.DB) (*DBWrapper, error) {
	w := &DBWrapper{db: db, tokenCache: make(map[string]map[string]*models.Token), activeLoans: make(map[string]map[string]*models.Loan)}
	var eg errgroup.Group
	eg.Go(w.loadTokenCache)
	eg.Go(w.loadActiveLoans)
	return w, eg.Wait()
}

func (w *DBWrapper) loadTokenCache() error {
	var tokens []models.Token
	if err := w.db.Find(&tokens).Error; err != nil {
		return fmt.Errorf("failed to load tokens: %w", err)
	}

	for _, token := range tokens {
		if w.tokenCache[token.ChainName] == nil {
			w.tokenCache[token.ChainName] = make(map[string]*models.Token)
		}
		w.tokenCache[token.ChainName][token.Address] = &token
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

func (w *DBWrapper) GetDB() *gorm.DB {
	return w.db
}

func (w *DBWrapper) GetTokenInfo(chainName string, address string) (*models.Token, error) {
	w.RLock()
	defer w.RUnlock()

	token, ok := w.tokenCache[chainName][address]
	if !ok {
		return nil, fmt.Errorf("token not found")
	}
	return token, nil
}

func (w *DBWrapper) AddTokenInfo(chainName string, address string, symbol string, decimals int) (*models.Token, error) {
	w.Lock()
	defer w.Unlock()

	token := &models.Token{
		ChainName: chainName,
		Address:   address,
		Symbol:    symbol,
		Decimals:  decimals,
	}
	if err := w.db.Create(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to add token info: %w", err)
	}
	if w.tokenCache[chainName] == nil {
		w.tokenCache[chainName] = make(map[string]*models.Token)
	}
	w.tokenCache[chainName][address] = token

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
