package aavev3

import (
	"fmt"
	"liquidation-bot/internal/models"
	"math/big"

	"go.uber.org/zap"
)

func (s *Service) checkHealthFactorForUser(user string, loan *models.Loan) error {
	return s.processBatch([]string{user}, map[string]*models.Loan{user: loan}, true)
}

func (s *Service) updateHealthFactorViaPrice(token string) error {
	return nil
}

const (
	batchSize = 100 // 每批处理的用户数量
)

func (s *Service) checkHealthFactor(loans []*models.Loan) error {
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
		if err := s.processBatch(batch, loansMap, false); err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}
	}
	return nil
}

func (s *Service) processBatch(batchUsers []string, activeLoans map[string]*models.Loan, findBestLiquidationInfos bool) error {
	// 获取用户账户数据
	accountDataMap, err := s.getUserAccountDataBatch(batchUsers)
	if err != nil {
		return fmt.Errorf("failed to get user account data for batch: %w", err)
	}

	// 处理每个用户
	deactivateUsers := make([]string, 0)
	updateHfUsers := make([]*UpdateLiquidationInfo, 0)
	updateInfoUsers := make([]*UpdateLiquidationInfo, 0)
	for _, user := range batchUsers {
		loan := activeLoans[user]
		accountData, ok := accountDataMap[user]
		if !ok {
			s.logger.Error("account data is nil", zap.String("chain", s.chain.ChainName), zap.String("user", user))
			continue
		}

		//	deactivate user if debt is less than MIN_DEBT_USD
		if debtUSD := big.NewFloat(0).Quo(big.NewFloat(0).SetInt(accountData.TotalDebtBase), USD_DECIMALS); debtUSD.Cmp(MIN_DEBT_USD) < 0 {
			s.logger.Info("total debt base is less than MIN_DEBT_USD", zap.String("user", user), zap.Any("debtUSD", debtUSD), zap.Any("minDebtUSD", MIN_DEBT_USD))
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
		updateHfUsers = append(updateHfUsers, info)

		// 需要重新计算，则更新 liquidationInfo
		if findBestLiquidationInfos {
			updateInfoUsers = append(updateInfoUsers, info)
		}
	}

	if len(deactivateUsers) > 0 {
		s.logger.Info("deactivate users", zap.Any("users", len(deactivateUsers)))
		if err := s.dbWrapper.DeactivateActiveLoan(s.chain.ChainName, deactivateUsers); err != nil {
			return fmt.Errorf("failed to deactivate active loan: %w", err)
		}
	}
	if len(updateHfUsers) > 0 {
		s.logger.Info("update health factor users", zap.Any("users", len(updateHfUsers)))
		// 更新 health factor 到数据库
		if err := s.dbWrapper.UpdateActiveLoanLiquidationInfos(s.chain.ChainName, updateHfUsers); err != nil {
			return fmt.Errorf("failed to update loan liquidation infos in database: %w", err)
		}
	}
	if len(updateInfoUsers) > 0 {
		s.logger.Info("find best liquidation infos", zap.Any("users", len(updateInfoUsers)))
		if err := s.findBestLiquidationInfos(updateInfoUsers); err != nil {
			return fmt.Errorf("failed to find best liquidation info: %w", err)
		}
	}

	return nil
}
