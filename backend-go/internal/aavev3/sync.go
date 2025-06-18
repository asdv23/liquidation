package aavev3

import (
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	"liquidation-bot/internal/models"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

func (s *Service) syncHealthFactorForUser(user string, loan *models.Loan) error {
	return s.processBatch([]string{user}, map[string]*models.Loan{user: loan}, true)
}

const (
	batchSize = 100 // 每批处理的用户数量
)

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
			s.toBeLiquidatedChan <- info.User
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
			base := amountToBase(debt, token.Decimals.BigInt(), token.Price.BigInt())
			baseBigInt := models.NewBigInt(base)
			if baseBigInt.BigInt().Cmp(liquidationInfo.DebtAmountBase.BigInt()) > 0 {
				liquidationInfo.TotalDebtBase = liquidationInfo.TotalDebtBase.Add(baseBigInt)
				liquidationInfo.DebtAmountBase = baseBigInt
				liquidationInfo.DebtAmount = models.NewBigInt(debt)
				liquidationInfo.DebtAsset = asset.Hex()
			}
			userReserves = append(userReserves, &models.Reserve{
				ChainName:           s.chain.ChainName,
				User:                user,
				Reserve:             asset.Hex(),
				Amount:              (*models.BigInt)(debt),
				IsBorrowing:         true,
				IsUsingAsCollateral: false,
			})
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
			userReserves = append(userReserves, &models.Reserve{
				ChainName:           s.chain.ChainName,
				User:                user,
				Reserve:             asset.Hex(),
				Amount:              (*models.BigInt)(collateral),
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
