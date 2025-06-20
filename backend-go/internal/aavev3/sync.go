package aavev3

import (
	"context"
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	"liquidation-bot/internal/models"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func (s *Service) syncHealthFactorForUser(user string, loan *models.Loan) error {
	return s.processBatch([]string{user}, map[string]*models.Loan{user: loan})
}

const (
	batchSize = 100 // 每批处理的用户数量
)

// 每 5 分钟针对不符合条件的 loan 重新同步
// - the graph 导入的
// - 发现脏数据触发重置了清算信息
func (s *Service) resync(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Minute):
			// 1. 处理没有清算信息的贷款
			noLiquidationInfoLoans, err := s.dbWrapper.GetNoLiquidationInfoLoans(s.chain.Ctx, s.chain.ChainName)
			if err != nil {
				return fmt.Errorf("failed to get loans with no liquidation information: %w", err)
			}
			s.logger.Info("found loans with no liquidation information", zap.Int("count", len(noLiquidationInfoLoans)))

			if err := s.syncHealthFactorForLoans(noLiquidationInfoLoans); err != nil {
				return fmt.Errorf("failed to sync health factor for loans: %w", err)
			}
		}
	}

}

func (s *Service) syncHealthFactorForLoans(loans []*models.Loan) error {
	if len(loans) == 0 {
		s.logger.Info("no loans to check health factor")
		return nil
	}

	// 收集需要检查的用户
	loansMap := make(map[string]*models.Loan)
	var usersToCheck []string
	for _, loan := range loans {
		loansMap[loan.User] = loan
		usersToCheck = append(usersToCheck, loan.User)
	}

	// 分批处理
	for i := 0; i < len(usersToCheck); i += batchSize {
		end := i + batchSize
		if end > len(usersToCheck) {
			end = len(usersToCheck)
		}
		s.logger.Info("processing batch", zap.Int("i", i), zap.Int("total", len(usersToCheck)), zap.Int("batchSize", batchSize))

		batch := usersToCheck[i:end]
		if err := s.processBatch(batch, loansMap); err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}
	}
	return nil
}

func (s *Service) processBatch(batchUsers []string, activeLoans map[string]*models.Loan) error {
	// 获取用户账户数据
	accountDataMap, err := s.getUserAccountDataBatch(batchUsers)
	if err != nil {
		return fmt.Errorf("failed to get user account data for batch: %w", err)
	}

	// 处理每个用户
	deactivateUsers := make([]string, 0)
	updateInfoUsers := make([]*UpdateLiquidationInfo, 0)
	for _, user := range batchUsers {
		loan := activeLoans[user]
		accountData, ok := accountDataMap[user]
		if !ok {
			s.logger.Error("account data is nil", zap.String("chain", s.chain.ChainName), zap.String("user", user))
			continue
		}

		//	deactivate user if debt is less than MIN_DEBT_BASE
		if debtBase := accountData.TotalDebtBase; debtBase.Cmp(MIN_DEBT_BASE) < 0 {
			s.logger.Info("total debt base is less than MIN_DEBT_BASE", zap.String("user", user), zap.Any("debtBase", debtBase), zap.Any("minDebtBase", MIN_DEBT_BASE))
			deactivateUsers = append(deactivateUsers, user)
			continue
		}

		// 计算健康因子
		healthFactor := formatHealthFactor(accountData.HealthFactor)
		if healthFactor == 0 {
			continue
		}
		if calcHealthFactor, ok := accountData.checkCalcHealthFactor(healthFactor); !ok {
			s.logger.Error("health factor mismatch ❌", zap.String("user", user), zap.Float64("healthFactor", healthFactor), zap.Float64("calcHealthFactor", calcHealthFactor))
		}

		// 检查并更新贷款信息
		if lastHealthFactor := loan.HealthFactor; lastHealthFactor == healthFactor {
			s.logger.Info("health factor not changed, skip update liquidation info", zap.String("user", user), zap.Float64("healthFactor", healthFactor))
			continue
		}
		s.logger.Info("health factor changed", zap.String("user", user),
			zap.Float64("lastHealthFactor", loan.HealthFactor), zap.Float64("healthFactor", healthFactor), zap.Any("acctHealthFactor", accountData.HealthFactor),
			zap.Any("totalCollateralBase", accountData.TotalCollateralBase),
			zap.Any("totalDebtBase", accountData.TotalDebtBase),
			zap.Any("currentLiquidationThreshold", accountData.CurrentLiquidationThreshold),
		)

		info := &UpdateLiquidationInfo{
			User:         user,
			HealthFactor: healthFactor,
			LiquidationInfo: &models.LiquidationInfo{
				TotalCollateralBase:  models.NewBigInt(accountData.TotalCollateralBase),
				TotalDebtBase:        models.NewBigInt(accountData.TotalDebtBase),
				LiquidationThreshold: models.NewBigInt(accountData.CurrentLiquidationThreshold),
			},
		}

		updateInfoUsers = append(updateInfoUsers, info)
	}

	if len(deactivateUsers) > 0 {
		s.logger.Info("deactivate users", zap.Any("users", len(deactivateUsers)))
		if err := s.dbWrapper.DeactivateActiveLoan(s.chain.ChainName, deactivateUsers); err != nil {
			return fmt.Errorf("failed to deactivate active loan: %w", err)
		}
	}
	if len(updateInfoUsers) > 0 {
		s.logger.Info("finding best liquidation infos", zap.Any("users", len(updateInfoUsers)))
		if err := s.findBestLiquidationInfos(updateInfoUsers); err != nil {
			return fmt.Errorf("failed to find best liquidation info: %w", err)
		}

		// 更新 liquidation info 到数据库
		s.logger.Info("updating loan liquidation infos in database", zap.Any("users", len(updateInfoUsers)))
		if err := s.dbWrapper.BatchUpdateLoanLiquidationInfos(s.chain.ChainName, updateInfoUsers); err != nil {
			return fmt.Errorf("failed to update loan liquidation infos in database: %w", err)
		}

		// 如果健康因子小于 1，则加入到待清算队列
		for _, info := range updateInfoUsers {
			if info.HealthFactor < 1 {
				s.logger.Info("health factor below liquidation threshold 🌟🌟🌟🌟🌟🌟", zap.String("user", info.User), zap.Any("healthFactor", info.HealthFactor))
				s.toBeLiquidatedChan <- info.User
			}
		}
	}

	return nil
}

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
		liquidationInfo.TotalCollateralBase = info.LiquidationInfo.TotalCollateralBase
		liquidationInfo.TotalDebtBase = info.LiquidationInfo.TotalDebtBase
		liquidationInfo.LiquidationThreshold = info.LiquidationInfo.LiquidationThreshold
		info.LiquidationInfo = liquidationInfo
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
		CollateralAsset:      (common.Address{}).Hex(),
		CollateralAmount:     models.NewBigInt(big.NewInt(0)),
		CollateralAmountBase: models.NewBigInt(big.NewInt(0)),
		DebtAsset:            (common.Address{}).Hex(),
		DebtAmount:           models.NewBigInt(big.NewInt(0)),
		DebtAmountBase:       models.NewBigInt(big.NewInt(0)),
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
		reserve := &models.Reserve{ChainName: s.chain.ChainName, User: user, Reserve: asset.Hex()}
		if isBorrowing(userConfig, i) {
			debt := big.NewInt(0).Add(userReserveData.CurrentStableDebt, userReserveData.CurrentVariableDebt)
			base := amountToBase(debt, token.Decimals.BigInt(), token.Price.BigInt())
			baseBigInt := models.NewBigInt(base)
			if baseBigInt.BigInt().Cmp(liquidationInfo.DebtAmountBase.BigInt()) > 0 {
				liquidationInfo.TotalDebtBase = liquidationInfo.TotalDebtBase.Add(baseBigInt)
				liquidationInfo.DebtAmountBase = baseBigInt
				liquidationInfo.DebtAmount = models.NewBigInt(debt)
				liquidationInfo.DebtAsset = asset.Hex()
			}
			reserve.BorrowedAmount = models.NewBigInt(debt)
			reserve.BorrowedAmountBase = baseBigInt
			reserve.IsBorrowing = true
			s.logger.Info("borrowing", zap.String("user", user), zap.Any("reserve", asset.Hex()), zap.Any("amount", debt), zap.Any("base", base))
		}

		if isUsingAsCollateral(userConfig, i) {
			collateral := big.NewInt(0).Set(userReserveData.CurrentATokenBalance)
			base := amountToBase(collateral, token.Decimals.BigInt(), token.Price.BigInt())
			baseBigInt := models.NewBigInt(base)
			if baseBigInt.BigInt().Cmp(liquidationInfo.CollateralAmountBase.BigInt()) > 0 {
				liquidationInfo.TotalCollateralBase = liquidationInfo.TotalCollateralBase.Add(baseBigInt)
				liquidationInfo.CollateralAmountBase = baseBigInt
				liquidationInfo.CollateralAmount = models.NewBigInt(collateral)
				liquidationInfo.CollateralAsset = asset.Hex()
			}
			reserve.CollateralAmount = models.NewBigInt(collateral)
			reserve.CollateralAmountBase = baseBigInt
			reserve.IsUsingAsCollateral = true
			s.logger.Info("collateral", zap.String("user", user), zap.Any("reserve", asset.Hex()), zap.Any("amount", collateral), zap.Any("base", base))
		}
		userReserves = append(userReserves, reserve)
	}
	if liquidationInfo.DebtAmount.BigInt().Cmp(big.NewInt(0)) == 0 || liquidationInfo.CollateralAmount.BigInt().Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("found no debt or collateral to liquidate, user: %s", user)
	}

	if err := s.dbWrapper.AddUserReserves(s.chain.ChainName, user, userReserves); err != nil {
		return nil, fmt.Errorf("failed to add user reserves: %w", err)
	}

	s.logger.Info("found best liquidation info", zap.String("user", user), zap.Any("liquidationInfo", liquidationInfo.String()))
	return &liquidationInfo, nil
}
