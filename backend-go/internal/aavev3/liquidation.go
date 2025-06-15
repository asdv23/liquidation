package aavev3

import (
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	"liquidation-bot/internal/models"
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

func (s *Service) findBestLiquidationInfos(liquidationInfos []*UpdateLiquidationInfo) error {
	users := make([]string, 0)
	for _, liquidationInfo := range liquidationInfos {
		users = append(users, liquidationInfo.User)
	}

	// TODO - 价格变化时，user config 不会变化
	userConfigs, err := s.getUserConfigurationForBatch(users)
	if err != nil {
		return fmt.Errorf("failed to get user configurations: %w", err)
	}
	if len(userConfigs) != len(liquidationInfos) {
		return fmt.Errorf("user configs length mismatch")
	}

	for i, userConfig := range userConfigs {
		info := liquidationInfos[i]
		liquidationInfo, err := s.findBestLiquidationInfo(info.User, userConfig) // 最好是链下计算
		if err != nil {
			return fmt.Errorf("failed to find best liquidation info: %w", err)
		}
		// userReserves不是和 total 同时查询，当存在数量和价格变化时，userReserves 和 total 就会不一致
		// if !checkUSDEqual(info.LiquidationInfo.TotalCollateralBase.BigInt(), liquidationInfo.TotalCollateralBase.BigInt()) {
		// 	s.logger.Info("calculate collateral base is not equal ❌❌", zap.String("user", info.User), zap.Any("info collateral base", info.LiquidationInfo.TotalCollateralBase.BigInt()), zap.Any("liquidationInfo collateral base", liquidationInfo.TotalCollateralBase.BigInt()))
		// }
		// if !checkUSDEqual(info.LiquidationInfo.TotalDebtBase.BigInt(), liquidationInfo.TotalDebtBase.BigInt()) {
		// 	s.logger.Info("calculate debt base is not equal ❌❌❌", zap.String("user", info.User), zap.Any("info debt base", info.LiquidationInfo.TotalDebtBase.BigInt()), zap.Any("liquidationInfo debt base", liquidationInfo.TotalDebtBase.BigInt()))
		// }
		liquidationInfo.TotalCollateralBase = models.NewBigInt(info.LiquidationInfo.TotalCollateralBase.BigInt())
		liquidationInfo.TotalDebtBase = models.NewBigInt(info.LiquidationInfo.TotalDebtBase.BigInt())
		liquidationInfo.LiquidationThreshold = models.NewBigInt(info.LiquidationInfo.LiquidationThreshold.BigInt())
		info.LiquidationInfo = liquidationInfo

		if info.HealthFactor < 1 {
			s.logger.Info("health factor below liquidation threshold 🌟🌟🌟🌟🌟🌟", zap.String("user", info.User), zap.Any("healthFactor", info.HealthFactor))
			// liquidationInfo, err := s.getLiquidationInfo(user, healthFactor)
			// if err != nil {
			// 	return fmt.Errorf("failed to get liquidation info: %w", err)
			// }

			// if liquidationInfo != nil {
			// 	// 执行清算
			// 	if err := s.executeLiquidation(user, healthFactor); err != nil {
			// 		return fmt.Errorf("failed to execute liquidation: %w", err)
			// 	}
			// }
		}
	}

	return nil
}

func (s *Service) findBestLiquidationInfo(user string, userConfig *aavev3.DataTypesUserConfigurationMap) (*models.LiquidationInfo, error) {
	userReserveDatas, err := s.getUserReserveDataBatch(user, userConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reserve data: %w", err)
	}

	liquidationInfo := models.LiquidationInfo{
		TotalCollateralBase:  models.NewBigInt(big.NewInt(0)),
		TotalDebtBase:        models.NewBigInt(big.NewInt(0)),
		LiquidationThreshold: models.NewBigInt(big.NewInt(0)),
		CollateralAmount:     models.NewBigInt(big.NewInt(0)),
		DebtAmount:           models.NewBigInt(big.NewInt(0)),
		CollateralAsset:      (common.Address{}).Hex(),
		DebtAsset:            (common.Address{}).Hex(),
	}
	userReserves := make([]*models.Reserve, 0)
	callIndex := 0
	for i, asset := range s.reservesList {
		if !isUsingAsCollateralOrBorrowing(userConfig, i) {
			continue
		}

		userReserveData := userReserveDatas[callIndex]
		callIndex++

		token, err := s.dbWrapper.GetTokenInfo(s.chain.ChainName, asset.Hex())
		if err != nil {
			return nil, fmt.Errorf("failed to get token info: %w", err)
		}
		if isBorrowing(userConfig, i) {
			debt := big.NewInt(0).Add(userReserveData.CurrentStableDebt, userReserveData.CurrentVariableDebt)
			base := amountToUSD(debt, token.Decimals.BigInt(), token.Price.BigInt())
			if base > liquidationInfo.DebtAmountBase {
				baseUSD := big.NewFloat(0).Mul(big.NewFloat(base), USD_DECIMALS)
				baseInt, _ := baseUSD.Int(nil)
				liquidationInfo.TotalDebtBase = models.NewBigInt(big.NewInt(0).Add(liquidationInfo.TotalDebtBase.BigInt(), baseInt))
				liquidationInfo.DebtAmountBase = base
				liquidationInfo.DebtAmount = (*models.BigInt)(debt)
				liquidationInfo.DebtAsset = asset.Hex()
			}
			userReserves = append(userReserves, &models.Reserve{
				ChainName:           s.chain.ChainName,
				User:                user,
				Reserve:             asset.Hex(),
				Amount:              (*models.BigInt)(debt),
				AmountBase:          base,
				IsBorrowing:         true,
				IsUsingAsCollateral: false,
			})
		}

		if isUsingAsCollateral(userConfig, i) {
			collateral := big.NewInt(0).Set(userReserveData.CurrentATokenBalance)
			base := amountToUSD(collateral, token.Decimals.BigInt(), token.Price.BigInt())
			if base > liquidationInfo.CollateralAmountBase {
				baseFloat := big.NewFloat(0).Mul(big.NewFloat(base), USD_DECIMALS)
				baseInt, _ := baseFloat.Int(nil)
				liquidationInfo.TotalCollateralBase = models.NewBigInt(big.NewInt(0).Add(liquidationInfo.TotalCollateralBase.BigInt(), baseInt))
				liquidationInfo.CollateralAmountBase = base
				liquidationInfo.CollateralAmount = (*models.BigInt)(collateral)
				liquidationInfo.CollateralAsset = asset.Hex()
			}
			userReserves = append(userReserves, &models.Reserve{
				ChainName:           s.chain.ChainName,
				User:                user,
				Reserve:             asset.Hex(),
				Amount:              (*models.BigInt)(collateral),
				AmountBase:          base,
				IsBorrowing:         false,
				IsUsingAsCollateral: true,
			})
		}
	}
	if err := s.dbWrapper.AddUserReserves(s.chain.ChainName, user, userReserves); err != nil {
		return nil, fmt.Errorf("failed to add user reserves: %w", err)
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
