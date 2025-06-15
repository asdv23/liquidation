package aavev3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	"liquidation-bot/internal/models"
	"liquidation-bot/pkg/blockchain"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

const (
	ODOS_QUOTE_URL    = "https://api.odos.xyz/sor/quote/v2"
	ODOS_ASSEMBLE_URL = "https://api.odos.xyz/sor/assemble"
)

func (s *Service) executeLiquidation(ctx context.Context, loan *models.Loan) error {
	if loan.LiquidationInfo == nil {
		userConfigs, err := s.getUserConfigurationForBatch([]string{loan.User})
		if err != nil {
			return fmt.Errorf("failed to get user configurations: %w", err)
		}
		loan.LiquidationInfo, err = s.findBestLiquidationInfo(loan.User, userConfigs[0])
		if err != nil {
			return fmt.Errorf("failed to find best liquidation info: %w", err)
		}
	}
	debtTokenInfo, err := s.dbWrapper.GetTokenInfo(s.chain.ChainName, loan.LiquidationInfo.DebtAsset)
	if err != nil {
		return fmt.Errorf("failed to get debt token info: %w", err)
	}

	// 全量清算 or 部分清算
	debtToCover := uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	debtToCoverUSD := loan.LiquidationInfo.DebtAmountBase
	if loan.LiquidationInfo.CollateralAmountBase < loan.LiquidationInfo.DebtAmountBase {
		// 部分清算
		debtToCoverUSD = loan.LiquidationInfo.CollateralAmountBase * (float64(999) / 1000)
		debtToCover = uint256.MustFromBig(USDToAmount(debtToCoverUSD, debtTokenInfo.Decimals.BigInt(), debtTokenInfo.Price.BigInt()))
	}

	s.logger.Info("💰 Executing flash loan liquidation with aggregator:", zap.String("user", loan.User), zap.Float64("healthFactor", loan.HealthFactor),
		zap.String("collateralAsset", loan.LiquidationInfo.CollateralAsset),
		zap.String("debtAsset", loan.LiquidationInfo.DebtAsset),
		zap.String("debtToCover", debtToCover.String()),
		zap.String("debtToCoverUSD", fmt.Sprintf("%f", debtToCoverUSD)),
	)

	go s.liquidateWithUniswapV3(ctx, loan, debtToCover)
	go s.liquidateWithOdos(ctx, loan, debtToCover, debtToCoverUSD)

	return nil
}

func (s *Service) liquidateWithUniswapV3(ctx context.Context, loan *models.Loan, debtToCover *uint256.Int) {
	logger := s.logger.Named("uniswap-v3").With()
	auth, err := s.chain.GetAuth()
	if err != nil {
		logger.Error("failed to get auth", zap.Error(err))
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			// use high gas tip
			gasTipCap, err := s.chain.GetClient().SuggestGasTipCap(ctx)
			if err != nil {
				logger.Error("failed to suggest gas price", zap.Error(err))
				continue
			}
			tip := gasTipCap.Mul(gasTipCap, big.NewInt(15)).Div(gasTipCap, big.NewInt(10))
			auth.GasTipCap = tip
			auth.GasFeeCap = tip.Add(tip, s.chain.GetBaseFee())

			// use pending state
			auth.Nonce = nil
			// auto estimate gas limit
			auth.GasLimit = 0

			logger.Info("prepared uniswap v3 gas params",
				zap.String("gasTipCap", gasTipCap.String()),
				zap.String("tip", tip.String()),
				zap.String("baseFee", s.chain.GetBaseFee().String()),
				zap.String("feeCap", auth.GasFeeCap.String()),
			)

			// send tx
			tx, err := s.chain.GetContracts().FlashLoanLiquidation.ExecuteLiquidation(auth,
				common.HexToAddress(loan.LiquidationInfo.CollateralAsset),
				common.HexToAddress(loan.LiquidationInfo.DebtAsset),
				common.HexToAddress(loan.User),
				debtToCover.ToBig(),
				[]byte{},
			)
			if err != nil {
				logger.Error("failed to execute liquidation", zap.Error(err))
				return
			}
			logger.Info("Liquidation with uniswap v3 transaction sent", zap.String("txHash", tx.Hash().Hex()))
		}
	}
}

func (s *Service) liquidateWithOdos(ctx context.Context, loan *models.Loan, debtToCover *uint256.Int, debtToCoverUSD float64) {
	logger := s.logger.Named("odos").With(zap.String("user", loan.User))
	auth, err := s.chain.GetAuth()
	if err != nil {
		logger.Error("failed to get auth", zap.Error(err))
		return
	}

	// 获取 aggregator data
	pathData, err := s.getAggregatorData(ctx, logger, loan, debtToCoverUSD)
	if err != nil {
		logger.Error("failed to get aggregator data", zap.Error(err))
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(60 * time.Second):
			pathData, err = s.getAggregatorData(ctx, logger, loan, debtToCoverUSD)
			if err != nil {
				logger.Error("failed to get aggregator data", zap.Error(err))
				continue
			}
		case <-time.After(100 * time.Millisecond):
			// use high gas tip
			gasTipCap, err := s.chain.GetClient().SuggestGasTipCap(ctx)
			if err != nil {
				logger.Error("failed to suggest gas price", zap.Error(err))
				continue
			}
			tip := gasTipCap.Mul(gasTipCap, big.NewInt(15)).Div(gasTipCap, big.NewInt(10))
			auth.GasTipCap = tip
			auth.GasFeeCap = tip.Add(tip, s.chain.GetBaseFee())

			// use pending state
			auth.Nonce = nil
			// auto estimate gas limit
			auth.GasLimit = 0

			logger.Info("prepared odos gas params",
				zap.String("gasTipCap", gasTipCap.String()),
				zap.String("tip", tip.String()),
				zap.String("baseFee", s.chain.GetBaseFee().String()),
				zap.String("feeCap", auth.GasFeeCap.String()),
				zap.String("pathData", common.Bytes2Hex(pathData)),
			)

			// send tx
			tx, err := s.chain.GetContracts().FlashLoanLiquidation.ExecuteLiquidation(auth,
				common.HexToAddress(loan.LiquidationInfo.CollateralAsset),
				common.HexToAddress(loan.LiquidationInfo.DebtAsset),
				common.HexToAddress(loan.User),
				debtToCover.ToBig(),
				pathData,
			)
			if err != nil {
				logger.Error("failed to execute liquidation", zap.Error(err))
				return
			}
			logger.Info("Liquidation with odos transaction sent", zap.String("txHash", tx.Hash().Hex()))
		}
	}
}

// getAggregatorData 从 Odos API 获取聚合器数据
func (s *Service) getAggregatorData(ctx context.Context, logger *zap.Logger, loan *models.Loan, debtToCoverUSD float64) ([]byte, error) {
	collateralTokenInfo, err := s.dbWrapper.GetTokenInfo(s.chain.ChainName, loan.LiquidationInfo.CollateralAsset)
	if err != nil {
		return nil, fmt.Errorf("failed to get collateral token info: %w", err)
	}
	collateralAmount := USDToAmount(debtToCoverUSD, collateralTokenInfo.Decimals.BigInt(), collateralTokenInfo.Price.BigInt())

	usdc := s.chain.GetContracts().Addresses[blockchain.ContractTypeUSDC].Hex()

	inputAmount := big.NewInt(0)
	//nolint:staticcheck // SA4006: this variable is used in QuotePayload
	outputTokens := make([]OutputToken, 0)
	if loan.LiquidationInfo.CollateralAsset == usdc {
		inputAmount = collateralAmount.Mul(collateralAmount, big.NewInt(958)).Div(collateralAmount, big.NewInt(1000))
		outputTokens = []OutputToken{
			{
				TokenAddress: loan.LiquidationInfo.DebtAsset,
				Proportion:   "1",
			},
		}
	} else if loan.LiquidationInfo.DebtAsset == usdc {
		inputAmount = collateralAmount.Mul(collateralAmount, big.NewInt(992)).Div(collateralAmount, big.NewInt(1000))
		outputTokens = []OutputToken{
			{
				TokenAddress: loan.LiquidationInfo.DebtAsset,
				Proportion:   "1",
			},
		}
	} else {
		inputAmount = collateralAmount
		outputTokens = []OutputToken{
			{
				TokenAddress: loan.LiquidationInfo.DebtAsset,
				Proportion:   "0.95",
			},
			{
				TokenAddress: usdc,
				Proportion:   "0.05",
			},
		}
	}

	// 构建请求数据
	payload := QuotePayload{
		ChainID: s.chain.ChainID.String(),
		InputTokens: []InputToken{
			{
				TokenAddress: loan.LiquidationInfo.CollateralAsset,
				Amount:       inputAmount.String(),
			},
		},
		OutputTokens:         outputTokens,
		UserAddr:             s.chain.GetContracts().Addresses[blockchain.ContractTypeFlashLoanLiquidation].Hex(),
		SlippageLimitPercent: "3",
		PathViz:              "false",
		PathVizImage:         "false",
	}
	logger.Info("quote payload", zap.Any("postData", payload))

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, fmt.Errorf("encode payload: %w", err)
	}
	// 发送请求
	resp, err := http.Post(
		"https://api.odos.xyz/sor/quote/v2",
		"application/json",
		&buf,
	)
	if err != nil {
		return nil, fmt.Errorf("post to odos api: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result QuoteResponse
	if err := jsoniter.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// 如果没有 pathId，返回空字符串
	if result.PathID == "" {
		return nil, fmt.Errorf("no path id")
	}

	// 获取路径数据
	return s.getPathData(ctx, result.PathID)
}

// getPathData 从 Odos API 获取路径数据
func (s *Service) getPathData(_ context.Context, pathID string) ([]byte, error) {
	usdc := s.chain.GetContracts().Addresses[blockchain.ContractTypeUSDC].Hex()
	receiver := s.chain.GetContracts().Addresses[blockchain.ContractTypeFlashLoanLiquidation].Hex()

	// 构建请求数据
	payload := AssemblePayload{
		UserAddr: receiver,
		PathID:   pathID,
		Simulate: false,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, fmt.Errorf("encode payload: %w", err)
	}

	// 发送请求
	resp, err := http.Post(
		ODOS_ASSEMBLE_URL,
		"application/json",
		&buf,
	)
	if err != nil {
		return nil, fmt.Errorf("post to odos api: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result AssembleResponse
	if err := jsoniter.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Transaction == nil {
		return nil, fmt.Errorf("no transaction")
	}

	// abi.encode([address, address, bytes], [usdc, result.Transaction.To, result.Transaction.Data])
	return encodeData(usdc, receiver, result.Transaction.Data)
}

// 假设 usdc, to 是 common.Address，data 是 []byte
func encodeData(usdc, to, data string) ([]byte, error) {
	// 定义参数类型
	args := abi.Arguments{
		{Type: abi.Type{Elem: nil, Size: 0, T: abi.AddressTy}}, // address
		{Type: abi.Type{Elem: nil, Size: 0, T: abi.AddressTy}}, // address
		{Type: abi.Type{Elem: nil, Size: 0, T: abi.BytesTy}},   // bytes
	}

	// 打包参数
	encoded, err := args.Pack(common.HexToAddress(usdc), common.HexToAddress(to), common.Hex2Bytes(data))
	if err != nil {
		return nil, fmt.Errorf("pack data: %w", err)
	}

	return encoded, nil
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
		if !checkUSDEqual(info.LiquidationInfo.TotalCollateralBase.BigInt(), liquidationInfo.TotalCollateralBase.BigInt()) {
			s.logger.Info("calculate collateral base is not equal ❌❌", zap.String("user", info.User), zap.Any("info collateral base", info.LiquidationInfo.TotalCollateralBase.BigInt()), zap.Any("liquidationInfo collateral base", liquidationInfo.TotalCollateralBase.BigInt()))
		}
		if !checkUSDEqual(info.LiquidationInfo.TotalDebtBase.BigInt(), liquidationInfo.TotalDebtBase.BigInt()) {
			s.logger.Info("calculate debt base is not equal ❌❌❌", zap.String("user", info.User), zap.Any("info debt base", info.LiquidationInfo.TotalDebtBase.BigInt()), zap.Any("liquidationInfo debt base", liquidationInfo.TotalDebtBase.BigInt()))
		}
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
