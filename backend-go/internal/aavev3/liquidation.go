package aavev3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"liquidation-bot/internal/models"
	"liquidation-bot/pkg/blockchain"
	"math/big"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
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
	if loan.HealthFactor == 0 || loan.LiquidationInfo == nil || loan.LiquidationInfo.DebtAsset == "" || loan.LiquidationInfo.CollateralAsset == "" {
		return fmt.Errorf("liquidation info is nil, user: %s", loan.User)
	}
	//	deactivate user if debt is less than MIN_DEBT_BASE
	if loan.LiquidationInfo.TotalDebtBase.BigInt().Cmp(MIN_DEBT_BASE) < 0 {
		s.logger.Info("total debt base is less than MIN_DEBT_BASE", zap.String("user", loan.User), zap.Any("debtBase", loan.LiquidationInfo.TotalDebtBase.BigInt()), zap.Any("minDebtBase", MIN_DEBT_BASE))
		return nil
	}

	debtTokenInfo, err := s.dbWrapper.GetTokenInfo(s.chain.ChainName, loan.LiquidationInfo.DebtAsset)
	if err != nil {
		return fmt.Errorf("failed to get debt token info: %w", err)
	}

	// 全量清算 or 部分清算
	debtToCover := uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	debtToCoverBase := amountToBase(loan.LiquidationInfo.DebtAmountBase.BigInt(), debtTokenInfo.Decimals.BigInt(), debtTokenInfo.Price.BigInt())
	if loan.LiquidationInfo.CollateralAmountBase.BigInt().Cmp(loan.LiquidationInfo.DebtAmountBase.BigInt()) < 0 {
		// 部分清算
		debtToCoverBase.Mul(loan.LiquidationInfo.CollateralAmountBase.BigInt(), big.NewInt(999))
		debtToCoverBase.Div(debtToCoverBase, big.NewInt(1000))
		debtToCover = uint256.MustFromBig(baseToAmount(debtToCoverBase, debtTokenInfo.Decimals.BigInt(), debtTokenInfo.Price.BigInt()))
		s.logger.Info("partial liquidation", zap.String("user", loan.User),
			zap.String("debtToCover", debtToCover.String()),
			zap.String("debtToCoverBase", debtToCoverBase.String()),
		)
	}

	s.logger.Info("💰 Executing flash loan liquidation with aggregator:", zap.String("user", loan.User), zap.Float64("healthFactor", loan.HealthFactor),
		zap.String("collateralAsset", loan.LiquidationInfo.CollateralAsset),
		zap.String("debtAsset", loan.LiquidationInfo.DebtAsset),
		zap.String("debtToCover", debtToCover.String()),
		zap.String("debtToCoverBase", debtToCoverBase.String()),
	)

	go func() {
		if err := retry.Do(func() error {
			return s.liquidateWithUniswapV3(ctx, loan, debtToCover)
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.LastErrorOnly(true)); err != nil {
			s.logger.Error("failed to liquidate with uniswap v3", zap.Error(err))
		}
	}()
	go func() {
		if err := retry.Do(func() error {
			return s.liquidateWithOdos(ctx, loan, debtToCover, debtToCoverBase)
		}, retry.Attempts(3), retry.Delay(100*time.Millisecond), retry.LastErrorOnly(true)); err != nil {
			s.logger.Error("failed to liquidate with odos", zap.Error(err))
		}
	}()

	return nil
}

func (s *Service) liquidateWithUniswapV3(ctx context.Context, loan *models.Loan, debtToCover *uint256.Int) error {
	logger := s.logger.Named("uniswap-v3").With(zap.String("user", loan.User))
	auth, err := s.chain.GetAuth()
	if err != nil {
		return fmt.Errorf("failed to get auth: %w", err)
	}
	// use high gas tip
	gasTipCap, err := s.chain.GetClient().SuggestGasTipCap(ctx)
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %w, user: %s", err, loan.User)
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
		return fmt.Errorf("failed to send execute liquidation tx: %w", err)
	}
	logger.Info("Liquidation with uniswap v3 transaction sent", zap.String("txHash", tx.Hash().Hex()))
	return nil
}

func (s *Service) liquidateWithOdos(ctx context.Context, loan *models.Loan, debtToCover *uint256.Int, debtToCoverBase *big.Int) error {
	logger := s.logger.Named("odos").With(zap.String("user", loan.User))
	auth, err := s.chain.GetAuth()
	if err != nil {
		return fmt.Errorf("failed to get auth: %w", err)
	}

	// 获取 aggregator data
	pathData, err := s.getAggregatorData(logger, loan, debtToCoverBase)
	if err != nil {
		return fmt.Errorf("failed to get aggregator data: %w, user: %s", err, loan.User)
	}
	// use high gas tip
	gasTipCap, err := s.chain.GetClient().SuggestGasTipCap(ctx)
	if err != nil {
		return fmt.Errorf("failed to suggest gas price: %w, user: %s", err, loan.User)
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
		return fmt.Errorf("failed to execute liquidation: %w", err)
	}
	logger.Info("Liquidation with odos transaction sent", zap.String("txHash", tx.Hash().Hex()))
	return nil
}

// getAggregatorData 从 Odos API 获取聚合器数据
func (s *Service) getAggregatorData(logger *zap.Logger, loan *models.Loan, debtToCoverBase *big.Int) ([]byte, error) {
	collateralTokenInfo, err := s.dbWrapper.GetTokenInfo(s.chain.ChainName, loan.LiquidationInfo.CollateralAsset)
	if err != nil {
		return nil, fmt.Errorf("failed to get collateral token info: %w", err)
	}
	collateralAmount := baseToAmount(debtToCoverBase, collateralTokenInfo.Decimals.BigInt(), collateralTokenInfo.Price.BigInt())

	// quote request
	inputAmount := big.NewInt(0)
	outputTokens := make([]OutputToken, 0)
	usdc := s.chain.GetContracts().Addresses[blockchain.ContractTypeUSDC].Hex()
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
	if inputAmount.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("input amount is 0")
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
	resp, err := http.Post(ODOS_QUOTE_URL, "application/json", &buf)
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
		return nil, fmt.Errorf("no path id, errCode: %d, detail: %s", result.ErrCode, result.Detail)
	}

	// 获取路径数据
	return s.getPathData(logger, result.PathID)
}

// getPathData 从 Odos API 获取路径数据
func (s *Service) getPathData(logger *zap.Logger, pathID string) ([]byte, error) {
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
	resp, err := http.Post(ODOS_ASSEMBLE_URL, "application/json", &buf)
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
