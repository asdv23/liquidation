package aavev3

// import (
// 	"fmt"
// 	"math/big"
// 	"time"

// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/core/types"
// 	"go.uber.org/zap"
// )

// func (s *Service) getLiquidationInfo(user string, healthFactor float64) (*LiquidationInfo, error) {
// 	// 检查缓存
// 	if info, ok := s.liquidationInfoCache[user]; ok {
// 		return info, nil
// 	}

// 	// 获取用户账户数据
// 	// accountData, err := s.getUserAccountData(user)
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("failed to get user account data: %w", err)
// 	// }

// 	// 获取用户资产列表
// 	assets, err := s.getUserAssets(user)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get user assets: %w", err)
// 	}

// 	// 计算清算信息
// 	info := &LiquidationInfo{
// 		User:             user,
// 		HealthFactor:     healthFactor,
// 		LastUpdated:      &time.Time{},
// 		CollateralAssets: make([]string, 0),
// 		DebtAssets:       make([]string, 0),
// 		CollateralPrices: make(map[string]*big.Int),
// 		DebtPrices:       make(map[string]*big.Int),
// 	}

// 	// 处理抵押品
// 	for _, asset := range assets {
// 		if asset.IsCollateral {
// 			info.CollateralAssets = append(info.CollateralAssets, asset.Address)
// 			info.CollateralPrices[asset.Address] = asset.Price
// 		} else {
// 			info.DebtAssets = append(info.DebtAssets, asset.Address)
// 			info.DebtPrices[asset.Address] = asset.Price
// 		}
// 	}

// 	// 更新缓存
// 	s.liquidationInfoCache[user] = info
// 	return info, nil
// }

// func (s *Service) executeLiquidation(user string, healthFactor float64) error {
// 	// 获取清算信息
// 	info, err := s.getLiquidationInfo(user, healthFactor)
// 	if err != nil {
// 		return fmt.Errorf("failed to get liquidation info: %w", err)
// 	}

// 	// 检查是否有可清算的资产
// 	if len(info.CollateralAssets) == 0 || len(info.DebtAssets) == 0 {
// 		return nil
// 	}

// 	// 选择最优的清算对
// 	bestPair, err := s.findBestLiquidationPair(info)
// 	if err != nil {
// 		return fmt.Errorf("failed to find best liquidation pair: %w", err)
// 	}

// 	if bestPair == nil {
// 		return nil
// 	}

// 	// 执行清算交易
// 	tx, err := s.executeLiquidationTx(bestPair)
// 	if err != nil {
// 		return fmt.Errorf("failed to execute liquidation transaction: %w", err)
// 	}
// 	s.logger.Info("Liquidation transaction sent", zap.String("txHash", tx.Hash().Hex()))

// 	return nil
// }

// func (s *Service) findBestLiquidationPair(info *LiquidationInfo) (*LiquidationPair, error) {
// 	var bestPair *LiquidationPair
// 	var maxProfit *big.Int

// 	// 遍历所有可能的清算对
// 	for _, collateral := range info.CollateralAssets {
// 		for _, debt := range info.DebtAssets {
// 			// 计算清算收益
// 			profit, err := s.calculateLiquidationProfit(info, collateral, debt)
// 			if err != nil {
// 				continue
// 			}

// 			// 更新最优对
// 			if bestPair == nil || profit.Cmp(maxProfit) > 0 {
// 				bestPair = &LiquidationPair{
// 					CollateralAsset: collateral,
// 					DebtAsset:       debt,
// 					Profit:          profit,
// 				}
// 				maxProfit = profit
// 			}
// 		}
// 	}

// 	return bestPair, nil
// }

// func (s *Service) calculateLiquidationProfit(
// 	info *LiquidationInfo,
// 	collateral string,
// 	debt string,
// ) (*big.Int, error) {
// 	// 获取清算参数
// 	params, err := s.getLiquidationParams(info, collateral, debt)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get liquidation params: %w", err)
// 	}

// 	// 计算清算收益
// 	profit := new(big.Int).Sub(
// 		params.CollateralAmount,
// 		params.DebtAmount,
// 	)

// 	return profit, nil
// }

// func (s *Service) getLiquidationParams(info *LiquidationInfo, collateral string, debt string) (*LiquidationParams, error) {
// 	return nil, nil
// }

// func (s *Service) executeLiquidationTx(pair *LiquidationPair) (*types.Transaction, error) {
// 	// 准备交易参数
// 	auth, err := s.chainClient.GetAuth(s.chainName)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get auth: %w", err)
// 	}

// 	// 执行清算
// 	tx, err := s.contracts.FlashLoanLiquidation.ExecuteLiquidation(auth,
// 		common.HexToAddress(pair.CollateralAsset),
// 		common.HexToAddress(pair.DebtAsset),
// 		common.HexToAddress(pair.User),
// 		big.NewInt(-1),
// 		[]byte{}, // data
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to execute liquidation: %w", err)
// 	}

// 	return tx, nil
// }
