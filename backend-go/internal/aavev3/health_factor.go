package aavev3

import (
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"
	"liquidation-bot/internal/models"
	"liquidation-bot/pkg/blockchain"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (s *Service) updateHealthFactorViaEvent(user string) error {
	activeLoans, ok := s.dbWrapper.GetActiveLoans(s.chain.ChainName)
	if !ok || len(activeLoans) == 0 {
		s.logger.Info("no active loans", zap.String("chain", s.chain.ChainName))
		return nil
	}

	return s.processBatch([]string{user}, activeLoans)
}

func (s *Service) updateHealthFactorViaPrice(token string) error {
	return nil
}

const (
	batchSize = 100 // 每批处理的用户数量
)

func (s *Service) checkHealthFactorsBatch() {
	activeLoans, ok := s.dbWrapper.GetActiveLoans(s.chain.ChainName)
	if !ok || len(activeLoans) == 0 {
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
	var eg errgroup.Group
	for _, user := range batchUsers {
		user := user
		accountData, ok := accountDataMap[user]
		if !ok {
			s.logger.Error("account data is nil", zap.String("chain", s.chain.ChainName), zap.String("user", user))
			continue
		}

		eg.Go(func() error {
			return s.processUser(user, accountData, activeLoans[user])
		})
	}

	return eg.Wait()
}

func (s *Service) processUser(user string, accountData *UserAccountData, loan *models.Loan) error {
	if loan == nil {
		return s.dbWrapper.CreateOrUpdateActiveLoan(s.chain.ChainName, user)
	}

	// 计算健康因子
	healthFactor := formatHealthFactor(accountData.HealthFactor)
	if healthFactor == 0 {
		return nil
	}
	if calcHealthFactor, ok := accountData.checkCalcHealthFactor(healthFactor); !ok {
		s.logger.Error("health factor mismatch ❌", zap.String("user", user), zap.Float64("healthFactor", healthFactor), zap.Float64("calcHealthFactor", calcHealthFactor))
	}

	// 检查并更新贷款信息
	if lastHealthFactor := loan.HealthFactor; lastHealthFactor != healthFactor {
		s.logger.Info("health factor changed", zap.String("user", user), zap.Float64("lastHealthFactor", lastHealthFactor), zap.Float64("healthFactor", healthFactor))

		// 更新 health factor 到数据库
		if err := s.dbWrapper.UpdateActiveLoanHealthFactor(s.chain.ChainName, user, healthFactor); err != nil {
			return fmt.Errorf("failed to update loan health factor in database: %w", err)
		}
	}

	// 检查是否需要清算
	if healthFactor <= 1 {
		s.logger.Info("health factor below liquidation threshold", zap.String("user", user), zap.Float64("healthFactor", healthFactor))
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
	var aggregate3Result []any
	raw := &bindings.Multicall3Raw{Contract: s.chain.GetContracts().Multicall3}
	if err := raw.Call(callOpts, &aggregate3Result, "aggregate3", &calls); err != nil {
		return nil, fmt.Errorf("failed to execute multicall: %w", err)
	}
	if len(aggregate3Result) == 0 {
		return nil, fmt.Errorf("failed to get aggregate3 result")
	}

	// 解析 aggregate3 结果
	aggregate3Results, ok := aggregate3Result[0].([]struct {
		Success    bool    "json:\"success\""
		ReturnData []uint8 "json:\"returnData\""
	})
	if !ok {
		return nil, fmt.Errorf("failed to parse aggregate3 result: %v", reflect.TypeOf(aggregate3Result[0]))
	}

	// 转换为 Multicall3Result 类型
	results := make([]bindings.Multicall3Result, len(aggregate3Results))
	for i, v := range aggregate3Results {
		results[i] = bindings.Multicall3Result{
			Success:    v.Success,
			ReturnData: v.ReturnData,
		}
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
