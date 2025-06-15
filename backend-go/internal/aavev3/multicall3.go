package aavev3

import (
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"
	"liquidation-bot/internal/models"
	"liquidation-bot/internal/utils"
	"liquidation-bot/pkg/blockchain"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// uint256 totalCollateralBase,
// uint256 totalDebtBase,
// uint256 availableBorrowsBase,
// uint256 currentLiquidationThreshold,
// uint256 ltv,
// uint256 healthFactor
//
// UserAccountData 用户账户数据
type UserAccountData struct {
	TotalCollateralBase         *big.Int
	TotalDebtBase               *big.Int
	AvailableBorrowsBase        *big.Int
	CurrentLiquidationThreshold *big.Int
	Ltv                         *big.Int
	HealthFactor                *big.Int
}

// (vars.totalCollateralInBaseCurrency.percentMul(vars.avgLiquidationThreshold)).wadDiv(
//
//	    vars.totalDebtInBaseCurrencyvars.healthFactor = (vars.totalDebtInBaseCurrency == 0)
//		? type(uint256).max
//		: (vars.totalCollateralInBaseCurrency.percentMul(vars.avgLiquidationThreshold)).wadDiv(
//		  vars.totalDebtInBaseCurrency
//		);
//
// 计算手算的健康因子和合约里是否一致
func (uad *UserAccountData) checkCalcHealthFactor(healthFactor float64) (float64, bool) {
	x := new(big.Int)
	calcHealthFactor := formatHealthFactor(x.Lsh(big.NewInt(1), 256).Sub(x, big.NewInt(1)))
	if uad.TotalDebtBase.Sign() != 0 {
		y := new(big.Int)
		y = y.Mul(uad.TotalCollateralBase, uad.CurrentLiquidationThreshold).Mul(y, big.NewInt(1e14)).Div(y, uad.TotalDebtBase)
		calcHealthFactor = formatHealthFactor(y)
	}
	if fmt.Sprintf("%0.2f", calcHealthFactor) != fmt.Sprintf("%0.2f", healthFactor) {
		fmt.Println("calcHealthFactor", fmt.Sprintf("%0.2f", calcHealthFactor), "healthFactor", fmt.Sprintf("%0.2f", healthFactor))
		return calcHealthFactor, false
	}
	return calcHealthFactor, true
}

// uint256 currentATokenBalance,
// uint256 currentStableDebt,
// uint256 currentVariableDebt,
// uint256 principalStableDebt,
// uint256 scaledVariableDebt,
// uint256 stableBorrowRate,
// uint256 liquidityRate,
// uint40 stableRateLastUpdated,
// bool usageAsCollateralEnabled
//
// UserReserveData 用户储备数据
type UserReserveData struct {
	CurrentATokenBalance     *big.Int
	CurrentVariableDebt      *big.Int
	CurrentStableDebt        *big.Int
	PrincipalStableDebt      *big.Int
	ScaledVariableDebt       *big.Int
	StableBorrowRate         *big.Int
	LiquidityRate            *big.Int
	StableRateLastUpdated    *big.Int
	UsageAsCollateralEnabled bool
}

func (s *Service) getUserConfigurationForBatch(users []string) ([]*aavev3.DataTypesUserConfigurationMap, error) {
	now := time.Now()
	target := s.chain.GetContracts().Addresses[blockchain.ContractTypeAaveV3Pool]
	abi, err := aavev3.PoolMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get abi: %w", err)
	}

	calls := make([]bindings.Multicall3Call3, 0)
	for _, user := range users {
		callData, err := abi.Pack("getUserConfiguration", common.HexToAddress(user))
		if err != nil {
			return nil, fmt.Errorf("failed to pack call data: %w", err)
		}

		call := bindings.Multicall3Call3{
			Target:   target,
			CallData: callData,
		}

		calls = append(calls, call)
	}

	// 准备调用参数
	opts, cancel := s.getCallOpts()
	defer cancel()
	results, err := utils.Aggregate3(opts, s.chain.GetContracts().Multicall3, calls)
	if err != nil {
		return nil, fmt.Errorf("failed to call aggregate: %w", err)
	}

	userConfigs := make([]*aavev3.DataTypesUserConfigurationMap, 0)
	for i, result := range results {
		if !result.Success {
			s.logger.Info("getUserConfiguration call failed", zap.String("user", users[i]))
			continue
		}
		var userConfig struct {
			Data aavev3.DataTypesUserConfigurationMap
		}
		if err := abi.UnpackIntoInterface(&userConfig, "getUserConfiguration", result.ReturnData); err != nil {
			return nil, fmt.Errorf("failed to unpack user configuration: %w", err)
		}

		userConfigs = append(userConfigs, &aavev3.DataTypesUserConfigurationMap{
			Data: userConfig.Data.Data,
		})
	}
	s.logger.Info("getUserConfiguration", zap.Any("len", len(results)), zap.Any("elapsed", time.Since(now)))

	return userConfigs, nil
}

func (s *Service) getUserAccountDataBatch(users []string) (map[string]*UserAccountData, error) {
	now := time.Now()
	// 准备批量调用数据
	aaveAbi, err := aavev3.PoolMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get abi: %w", err)
	}

	var calls []bindings.Multicall3Call3
	for _, user := range users {
		// 编码 getUserAccountData 调用
		callData, err := aaveAbi.Pack("getUserAccountData", common.HexToAddress(user))
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

	s.logger.Info("getUserAccountData", zap.Any("len", len(results)), zap.Any("elapsed", time.Since(now)))
	return accountDataMap, nil
}

func (s *Service) updateReservesListAndPrice() error {
	now := time.Now()
	callOpts, cancel := s.getCallOpts()
	defer cancel()

	reservesList, err := s.chain.GetContracts().AaveV3Pool.GetReservesList(callOpts)
	if err != nil {
		return fmt.Errorf("failed to get reserves list: %w", err)
	}
	s.reservesList = reservesList

	// token
	erc20Abi, err := bindings.ERC20MetaData.GetAbi()
	if err != nil {
		return fmt.Errorf("failed to get erc20 abi: %w", err)
	}
	decimalsCalls, symbolsCalls := make([]bindings.Multicall3Call3, 0), make([]bindings.Multicall3Call3, 0)

	// price
	abi, err := aavev3.AaveOracleMetaData.GetAbi()
	if err != nil {
		return fmt.Errorf("failed to get aave oracle abi: %w", err)
	}
	target := s.chain.GetContracts().Addresses[blockchain.ContractTypePriceOracle]

	getReservesPriceCalls := make([]bindings.Multicall3Call3, 0)
	for _, reserve := range reservesList {
		symbolsCall, decimalsCall, err := getSymbolAndDecimalsMulticall3Call3(erc20Abi, reserve)
		if err != nil {
			return fmt.Errorf("failed to get symbol and decimals: %w", err)
		}
		decimalsCalls = append(decimalsCalls, decimalsCall)
		symbolsCalls = append(symbolsCalls, symbolsCall)

		callData, err := abi.Pack("getAssetPrice", reserve)
		if err != nil {
			return fmt.Errorf("failed to pack get asset price call: %w", err)
		}

		getReservesPriceCalls = append(getReservesPriceCalls, bindings.Multicall3Call3{
			Target:   target,
			CallData: callData,
		})
	}

	var symbolsResults []bindings.Multicall3Result
	var decimalsResults []bindings.Multicall3Result
	var results []bindings.Multicall3Result
	var eg errgroup.Group
	eg.Go(func() error {
		symbolsResults, err = utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, symbolsCalls)
		if err != nil {
			return fmt.Errorf("failed to get symbols: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		decimalsResults, err = utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, decimalsCalls)
		if err != nil {
			return fmt.Errorf("failed to get decimals: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		results, err = utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, getReservesPriceCalls)
		if err != nil {
			return fmt.Errorf("failed to get reserves price: %w", err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to update reserves list and price: %w", err)
	}

	for i, result := range results {
		price := new(big.Int).SetBytes(result.ReturnData)
		decimals := new(big.Int).SetBytes(decimalsResults[i].ReturnData)
		symbol := decodeSymbol(symbolsResults[i].ReturnData, erc20Abi)
		if _, err := s.dbWrapper.AddTokenInfo(s.chain.ChainName, reservesList[i].Hex(), symbol, decimals, price); err != nil {
			return fmt.Errorf("failed to add token info: %w", err)
		}
	}

	s.logger.Info("updateReservesListAndPrice", zap.Any("len", len(reservesList)), zap.Any("elapsed", time.Since(now)))
	return nil
}

func (s *Service) createTokenInfoFromChain(asset string) (*models.Token, error) {
	s.logger.Warn("find new token", zap.String("asset", asset))
	erc20Abi, err := bindings.ERC20MetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get erc20 abi: %w", err)
	}

	symbolCall, decimalsCall, err := getSymbolAndDecimalsMulticall3Call3(erc20Abi, common.HexToAddress(asset))
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol and decimals call data: %w", err)
	}

	callOpts, cancel := s.getCallOpts()
	defer cancel()

	symbolsResults, err := utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, []bindings.Multicall3Call3{symbolCall})
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}
	decimalsResults, err := utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, []bindings.Multicall3Call3{decimalsCall})
	if err != nil {
		return nil, fmt.Errorf("failed to get decimals: %w", err)
	}

	symbol := decodeSymbol(symbolsResults[0].ReturnData, erc20Abi)
	decimals := new(big.Int).SetBytes(decimalsResults[0].ReturnData)

	token, err := s.dbWrapper.AddTokenInfo(s.chain.ChainName, asset, symbol, decimals, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to add token info: %w", err)
	}

	return token, nil
}

func (s *Service) getUserReserveDataBatch(user string, userConfig *aavev3.DataTypesUserConfigurationMap) ([]*UserReserveData, error) {
	now := time.Now()
	abi, err := aavev3.AaveProtocolDataProviderMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get abi: %w", err)
	}
	target := s.chain.GetContracts().Addresses[blockchain.ContractTypeDataProvider]

	// 遍历所有可能的清算对
	getUserReserveDataCalls := make([]bindings.Multicall3Call3, 0)
	for reserveIndex, reserve := range s.reservesList {
		if isUsingAsCollateralOrBorrowing(userConfig, reserveIndex) {
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

	userReserves := make([]*UserReserveData, 0)
	for _, result := range results {
		var userReserveData UserReserveData
		if err := abi.UnpackIntoInterface(&userReserveData, "getUserReserveData", result.ReturnData); err != nil {
			return nil, fmt.Errorf("failed to unpack user reserve data: %w", err)
		}

		userReserves = append(userReserves, &userReserveData)
	}
	s.logger.Info("getUserReserveData", zap.Any("user", user), zap.Any("len", len(results)), zap.Any("elapsed", time.Since(now)))
	return userReserves, nil
}
