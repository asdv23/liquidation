package aavev3

import (
	"liquidation-bot/internal/models"
	"math/big"
	"time"
)

// LiquidationInfo 清算信息
type LiquidationInfo struct {
	User             string
	HealthFactor     float64
	LastUpdated      *time.Time
	CollateralAssets []string
	DebtAssets       []string
	CollateralPrices map[string]*big.Int
	DebtPrices       map[string]*big.Int
}

type UpdateLiquidationInfo struct {
	User            string
	HealthFactor    float64
	LiquidationInfo *models.LiquidationInfo
}
