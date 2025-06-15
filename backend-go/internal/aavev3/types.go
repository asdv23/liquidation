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

type InputToken struct {
	TokenAddress string `json:"tokenAddress"`
	Amount       string `json:"amount"`
}

type OutputToken struct {
	TokenAddress string `json:"tokenAddress"`
	Proportion   string `json:"proportion"`
}

type QuotePayload struct {
	ChainID              string        `json:"chainId"`
	InputTokens          []InputToken  `json:"inputTokens"`
	OutputTokens         []OutputToken `json:"outputTokens"`
	UserAddr             string        `json:"userAddr"`
	SlippageLimitPercent string        `json:"slippageLimitPercent"`
	PathViz              string        `json:"pathViz"`
	PathVizImage         string        `json:"pathVizImage"`
}

type QuoteResponse struct {
	PathID string `json:"pathId"`
}

type AssemblePayload struct {
	UserAddr string `json:"userAddr"`
	PathID   string `json:"pathId"`
	Simulate bool   `json:"simulate"`
	Receiver string `json:"receiver"`
}

type AssembleResponse struct {
	Transaction *transaction `json:"transaction"`
}
type transaction struct {
	Value string `json:"value"`
	To    string `json:"to"`
	Data  string `json:"data"`
}
