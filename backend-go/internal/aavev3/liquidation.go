package aavev3

import (
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"
	"liquidation-bot/internal/models"
	"liquidation-bot/internal/utils"
	"liquidation-bot/pkg/blockchain"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

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

func (s *Service) updateLiquidationInfo(user string) error {
	oldLiquidationInfo, err := s.dbWrapper.GetLiquidationInfo(s.chain.ChainName, user)
	if err != nil {
		return fmt.Errorf("failed to get liquidation info: %w", err)
	}
	liquidationInfo, err := s.findBestLiquidationInfo(user)
	if err != nil {
		return fmt.Errorf("failed to find best liquidation info: %w", err)
	}
	if !liquidationInfo.Cmp(oldLiquidationInfo) {
		s.logger.Info("liquidation info changed", zap.String("user", user), zap.Any("old", oldLiquidationInfo), zap.Any("new", oldLiquidationInfo))
	}

	if err := s.dbWrapper.UpdateActiveLoanLiquidationInfo(s.chain.ChainName, user, liquidationInfo); err != nil {
		return fmt.Errorf("failed to update loan liquidation info: %w", err)
	}

	return nil
}

func (s *Service) findBestLiquidationInfo(user string) (*models.LiquidationInfo, error) {
	userConfig, err := s.getUserConfiguration(user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user configuration: %w", err)
	}
	abi, err := aavev3.AaveProtocolDataProviderMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get abi: %w", err)
	}
	target := s.chain.GetContracts().Addresses[blockchain.ContractTypeDataProvider]

	// 遍历所有可能的清算对
	getUserReserveDataCalls := make([]bindings.Multicall3Call3, 0)
	for reserveIndex, reserve := range s.reservesList {
		if isBorrowing(userConfig, reserveIndex) || isUsingAsCollateral(userConfig, reserveIndex) {
			callData, err := abi.Pack("getUserReserveData", reserve, common.HexToAddress(user))
			if err != nil {
				return nil, fmt.Errorf("failed to pack getUserReserveData call: %w", err)
			}
			getUserReserveDataCalls = append(getUserReserveDataCalls, bindings.Multicall3Call3{
				Target:   target,
				CallData: callData,
			})
		}
	}

	// 执行模拟调用
	callOpts, cancel := s.getCallOpts()
	defer cancel()
	results, err := utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, getUserReserveDataCalls)
	if err != nil {
		return nil, fmt.Errorf("failed to parse aggregate3 result: %w", err)
	}

	var liquidationInfo models.LiquidationInfo
	for i, result := range results {
		if !result.Success {
			continue
		}
		var userReserveData UserReserveData
		if err := abi.UnpackIntoInterface(&userReserveData, "getUserReserveData", result.ReturnData); err != nil {
			return nil, fmt.Errorf("failed to unpack user reserve data: %w", err)
		}

		asset := s.reservesList[i]
		token, err := s.dbWrapper.GetTokenInfo(s.chain.ChainName, asset.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get token info: %w", err)
		}
		if isBorrowing(userConfig, i) {
			debt := big.NewInt(0).Add(userReserveData.CurrentStableDebt, userReserveData.CurrentVariableDebt)
			base := amountToUSD(debt, token.Decimals, (*big.Int)(token.Price))
			if base > liquidationInfo.DebtAmountBase {
				liquidationInfo.DebtAmountBase = base
				liquidationInfo.DebtAmount = (*models.BigInt)(debt)
				liquidationInfo.DebtAsset = asset.Hex()
			}
		}

		if isUsingAsCollateral(userConfig, i) {
			collateral := big.NewInt(0).Set(userReserveData.CurrentATokenBalance)
			base := amountToUSD(collateral, token.Decimals, (*big.Int)(token.Price))
			if base > liquidationInfo.CollateralAmountBase {
				liquidationInfo.CollateralAmountBase = base
				liquidationInfo.CollateralAmount = (*models.BigInt)(collateral)
				liquidationInfo.CollateralAsset = asset.Hex()
			}
		}
	}

	return &liquidationInfo, nil
}

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
