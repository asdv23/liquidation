package aavev3

import (
	"context"
	"fmt"
	"liquidation-bot/internal/models"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"
)

func (s *Service) updateReservesListAndPrice() error {
	now := time.Now()
	callOpts, cancel := s.getCallOpts()
	defer cancel()

	// reservesList
	reservesList, err := s.chain.GetContracts().AaveV3Pool.GetReservesList(callOpts)
	if err != nil {
		return fmt.Errorf("failed to get reserves list: %w", err)
	}
	s.reservesList = reservesList

	// prices
	prices, err := s.chain.GetContracts().PriceOracle.GetAssetsPrices(callOpts, reservesList)
	if err != nil {
		return fmt.Errorf("failed to get assets prices: %w", err)
	}

	// token
	reserveInfos, err := s.getReservesDecimalsAndSymbols()
	if err != nil {
		return fmt.Errorf("failed to get reserves decimals and symbols: %w", err)
	}

	// upsert token info
	for i, reserveInfo := range reserveInfos {
		if _, err := s.dbWrapper.UpsertTokenInfo(s.chain.ChainName, reserveInfo.Address, reserveInfo.Symbol, reserveInfo.Decimals, prices[i]); err != nil {
			return fmt.Errorf("failed to add token info: %w", err)
		}
	}

	s.logger.Info("updateReservesListAndPrice", zap.Any("len", len(reservesList)), zap.Any("elapsed", time.Since(now)))
	return nil
}

// 如果 ReserveList 中的 token 价格变化：
// 1. 更新对应的 Token 价格
// 2. 获取用户 Reserve 中有此 token 的 用户的所有用户 Reserve
// 3. 计算并更新用户 Loan 的 totalDebt和totalCollateral，以及健康因子
// 4. 如果健康因子低于阈值，则进行清算
func (s *Service) startSyncPricesForReserveList(ctx context.Context) error {
	for {
		if err := s.syncPricesForReserveList(ctx); err != nil {
			s.logger.Error("failed to sync prices for reserve list", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context done: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}
}

func (s *Service) syncPricesForReserveList(ctx context.Context) error {
	now := time.Now()
	defer func() {
		s.logger.Info("syncPricesForReserveList", zap.Any("elapsed", time.Since(now)))
	}()

	callOpts, cancel := s.getCallOpts()
	defer cancel()

	prices, err := s.chain.GetContracts().PriceOracle.GetAssetsPrices(callOpts, s.reservesList)
	if err != nil {
		return fmt.Errorf("failed to get assets prices: %w", err)
	}
	tokenInfoMap, err := s.dbWrapper.GetTokenInfoMap(s.chain.ChainName)
	if err != nil {
		return fmt.Errorf("failed to get token infos: %w", err)
	}
	updatedReserves := make([]string, 0)
	for i, price := range prices {
		tokenInfo := tokenInfoMap[s.reservesList[i].Hex()]
		if tokenInfo.Price.BigInt().Cmp(price) == 0 {
			continue
		}
		// update token price
		if err := s.dbWrapper.UpdateTokenPrice(s.chain.ChainName, tokenInfo.Address, models.NewBigInt(price)); err != nil {
			return fmt.Errorf("failed to update token price: %w", err)
		}
		s.logger.Info("update token price", zap.String("token", tokenInfo.Address), zap.String("newPrice", price.String()), zap.String("oldPrice", tokenInfo.Price.BigInt().String()))

		// update tokenInfoMap
		tokenInfo.Price = models.NewBigInt(price)
		updatedReserves = append(updatedReserves, tokenInfo.Address)
	}

	if len(updatedReserves) == 0 {
		return nil
	}

	// find all user reserves with updatedReserves
	userLoans, userReserves, err := s.dbWrapper.GetUserLoansAndReservesByReserves(s.chain.ChainName, updatedReserves)
	if err != nil {
		return fmt.Errorf("failed to get user reserves: %w", err)
	}

	// calc base for each user reserve
	// calc max debt and max collateral for each user
	userLiquidationInfoMap := make(map[string]*models.LiquidationInfo)
	// calc total debt and total collateral for each user
	for _, userReserve := range userReserves {
		tokenInfo := tokenInfoMap[userReserve.Reserve]
		amountBase := amountToBase(userReserve.Amount.BigInt(), tokenInfo.Decimals.BigInt(), tokenInfo.Price.BigInt())
		amountBaseBigInt := models.NewBigInt(amountBase)
		if _, ok := userLiquidationInfoMap[userReserve.User]; !ok {
			defaultBigInt := models.NewBigInt(big.NewInt(0))
			if userReserve.IsBorrowing {
				userLiquidationInfoMap[userReserve.User] = &models.LiquidationInfo{
					TotalDebtBase:        amountBaseBigInt,
					DebtAsset:            userReserve.Reserve,
					DebtAmount:           userReserve.Amount,
					DebtAmountBase:       amountBaseBigInt,
					TotalCollateralBase:  defaultBigInt,
					CollateralAsset:      "",
					CollateralAmount:     defaultBigInt,
					CollateralAmountBase: defaultBigInt,
				}
			}
			if userReserve.IsUsingAsCollateral {
				userLiquidationInfoMap[userReserve.User] = &models.LiquidationInfo{
					TotalCollateralBase:  amountBaseBigInt,
					CollateralAsset:      userReserve.Reserve,
					CollateralAmount:     userReserve.Amount,
					CollateralAmountBase: amountBaseBigInt,
					TotalDebtBase:        defaultBigInt,
					DebtAsset:            "",
					DebtAmount:           defaultBigInt,
					DebtAmountBase:       defaultBigInt,
				}
			}
			continue
		}
		liquidationInfo := userLiquidationInfoMap[userReserve.User]

		// update total debt and max debt amount
		if userReserve.IsBorrowing {
			liquidationInfo.TotalDebtBase = liquidationInfo.TotalDebtBase.Add(amountBaseBigInt)
			if liquidationInfo.DebtAsset == "" || amountBaseBigInt.BigInt().Cmp(liquidationInfo.DebtAmountBase.BigInt()) > 0 {
				liquidationInfo.DebtAsset = userReserve.Reserve
				liquidationInfo.DebtAmount = userReserve.Amount
				liquidationInfo.DebtAmountBase = amountBaseBigInt
			}
		}

		// update total collateral and max collateral amount
		if userReserve.IsUsingAsCollateral {
			liquidationInfo.TotalCollateralBase = liquidationInfo.TotalCollateralBase.Add(amountBaseBigInt)
			if liquidationInfo.CollateralAsset == "" || amountBaseBigInt.BigInt().Cmp(liquidationInfo.CollateralAmountBase.BigInt()) > 0 {
				liquidationInfo.CollateralAsset = userReserve.Reserve
				liquidationInfo.CollateralAmount = userReserve.Amount
				liquidationInfo.CollateralAmountBase = amountBaseBigInt
			}
		}
	}

	// calc user health factor(need LiquidationThreshold)
	var wg sync.WaitGroup
	batchSize := 1000
	for i := 0; i < len(userLoans); i += batchSize {
		end := i + batchSize
		if end > len(userLoans) {
			end = len(userLoans)
		}
		batch := userLoans[i:end]
		s.logger.Info("sync prices for loan list", zap.Int("i", i), zap.Int("total", len(userLoans)), zap.Int("batchSize", batchSize))
		wg.Add(1)
		go func(batch []*models.Loan) {
			defer wg.Done()

			liquidationInfos := make([]*UpdateLiquidationInfo, 0)
			toBeLiquidated := make([]string, 0)
			for _, loan := range batch {
				liquidationInfo := userLiquidationInfoMap[loan.User]
				if liquidationInfo.DebtAsset == "" || liquidationInfo.CollateralAsset == "" {
					s.logger.Error("debt or collateral is empty, user: %s", zap.String("user", loan.User), zap.String("liquidationInfo", liquidationInfo.String()))
					continue
				}
				healthFactor := calcHealthFactor(liquidationInfo.TotalCollateralBase.BigInt(), liquidationInfo.TotalDebtBase.BigInt(), loan.LiquidationInfo.LiquidationThreshold.BigInt())
				s.logger.Info("health factor changed", zap.String("user", loan.User), zap.Float64("lastHealthFactor", loan.HealthFactor), zap.Float64("healthFactor", healthFactor))

				liquidationInfos = append(liquidationInfos, &UpdateLiquidationInfo{
					User:            loan.User,
					HealthFactor:    healthFactor,
					LiquidationInfo: liquidationInfo,
				})

				if healthFactor < 1 {
					s.logger.Info("user health factor is below threshold", zap.String("user", loan.User), zap.Float64("healthFactor", healthFactor))
					toBeLiquidated = append(toBeLiquidated, loan.User)
				}
			}
			if err := s.dbWrapper.BatchUpdateLoanLiquidationInfos(s.chain.ChainName, liquidationInfos); err != nil {
				s.logger.Error("failed to batch update loan liquidation infos", zap.Error(err))
			}

			for _, user := range toBeLiquidated {
				s.toBeLiquidatedChan <- user
			}
		}(batch)
	}
	wg.Wait()
	return nil
}
