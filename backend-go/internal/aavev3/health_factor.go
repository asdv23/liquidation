package aavev3

import (
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"
	"liquidation-bot/internal/models"
	"liquidation-bot/internal/utils"
	"liquidation-bot/pkg/blockchain"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func (s *Service) updateHealthFactorViaEvent(user string, loan *models.Loan) error {
	return s.processBatch([]string{user}, map[string]*models.Loan{user: loan})
}

func (s *Service) updateHealthFactorViaPrice(token string) error {
	return nil
}

const (
	batchSize = 100 // 每批处理的用户数量
)

func (s *Service) checkHealthFactorsBatch() {
	activeLoans, err := s.dbWrapper.ChainActiveLoans(s.chain.ChainName)
	if err != nil {
		s.logger.Error("failed to get active loans", zap.Error(err))
		return
	}
	if len(activeLoans) == 0 {
		s.logger.Info("no active loans")
		return
	}

	// 收集需要检查的用户
	var usersToCheck []string
	for user := range activeLoans {
		usersToCheck = append(usersToCheck, user)
	}
	if len(usersToCheck) == 0 {
		return
	}

	// 分批处理
	for i := 0; i < len(usersToCheck); i += batchSize {
		end := i + batchSize
		if end > len(usersToCheck) {
			end = len(usersToCheck)
		}
		s.logger.Info("processing batch", zap.Int("i", i), zap.Int("total", len(usersToCheck)), zap.Int("batchSize", batchSize))

		batch := usersToCheck[i:end]
		if err := s.processBatch(batch, activeLoans); err != nil {
			s.logger.Error("Failed to process batch",
				zap.String("chain", s.chain.ChainName),
				zap.Error(err))
		}
	}
}

func (s *Service) processBatch(batchUsers []string, activeLoans map[string]*models.Loan) error {
	// 获取用户账户数据
	accountDataMap, err := s.getUserAccountDataBatch(batchUsers)
	if err != nil {
		return fmt.Errorf("failed to get user account data for batch: %w", err)
	}

	// 处理每个用户
	deactivateUsers := make([]string, 0)
	liquidationInfos := make([]*UpdateLiquidationInfo, 0)
	for _, user := range batchUsers {
		loan := activeLoans[user]
		accountData, ok := accountDataMap[user]
		if !ok {
			s.logger.Error("account data is nil", zap.String("chain", s.chain.ChainName), zap.String("user", user))
			continue
		}

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
		s.logger.Info("health factor changed", zap.String("user", user), zap.Float64("lastHealthFactor", loan.HealthFactor), zap.Float64("healthFactor", healthFactor))

		liquidationInfos = append(liquidationInfos, &UpdateLiquidationInfo{
			User:         user,
			HealthFactor: healthFactor,
			LiquidationInfo: &models.LiquidationInfo{
				TotalCollateralBase:  models.NewBigInt(accountData.TotalCollateralBase),
				TotalDebtBase:        models.NewBigInt(accountData.TotalDebtBase),
				LiquidationThreshold: models.NewBigInt(accountData.CurrentLiquidationThreshold),
			},
		})
	}

	if len(deactivateUsers) > 0 {
		s.logger.Info("deactivate users", zap.Any("users", len(deactivateUsers)))
		if err := s.dbWrapper.DeactivateActiveLoan(s.chain.ChainName, deactivateUsers); err != nil {
			return fmt.Errorf("failed to deactivate active loan: %w", err)
		}
	}
	if len(liquidationInfos) > 0 {
		s.logger.Info("update health factor users", zap.Any("users", len(liquidationInfos)))
		// 查找最佳清算信息
		if err := s.findBestLiquidationInfos(liquidationInfos); err != nil {
			return fmt.Errorf("failed to find best liquidation info: %w", err)
		}
		// 更新 health factor 到数据库
		if err := s.dbWrapper.UpdateActiveLoanLiquidationInfos(s.chain.ChainName, liquidationInfos); err != nil {
			return fmt.Errorf("failed to update loan liquidation infos in database: %w", err)
		}
	}

	return nil
}

func (s *Service) getUserAccountDataBatch(users []string) (map[string]*UserAccountData, error) {
	// 准备批量调用数据
	var calls []bindings.Multicall3Call3
	var aaveAbi *abi.ABI

	for _, user := range users {
		abi, err := aavev3.PoolMetaData.GetAbi()
		if err != nil {
			return nil, fmt.Errorf("failed to get abi: %w", err)
		}
		aaveAbi = abi
		// 编码 getUserAccountData 调用
		callData, err := abi.Pack("getUserAccountData", common.HexToAddress(user))
		if err != nil {
			return nil, fmt.Errorf("failed to pack getUserAccountData call: %w", err)
		}

		calls = append(calls, bindings.Multicall3Call3{
			Target:       s.chain.GetContracts().Addresses[blockchain.ContractTypeAaveV3Pool],
			AllowFailure: false,
			CallData:     callData,
		})
	}

	// 执行模拟调用
	callOpts, cancel := s.getCallOpts()
	defer cancel()
	results, err := utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, calls)
	if err != nil {
		return nil, fmt.Errorf("failed to parse aggregate3 result: %w", err)
	}

	// 解析每个用户的数据
	accountDataMap := make(map[string]*UserAccountData)
	for i, result := range results {
		if !result.Success {
			continue
		}

		var data UserAccountData
		if err := aaveAbi.UnpackIntoInterface(&data, "getUserAccountData", result.ReturnData); err != nil {
			continue
		}

		accountDataMap[users[i]] = &data
	}

	return accountDataMap, nil
}
