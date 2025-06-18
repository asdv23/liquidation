package aavev3

import (
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"
	"liquidation-bot/internal/utils"
	"liquidation-bot/pkg/blockchain"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

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

func (s *Service) getReservesDecimalsAndSymbols() ([]*ReserveInfo, error) {
	now := time.Now()
	callOpts, cancel := s.getCallOpts()
	defer cancel()

	// token
	erc20Abi, err := bindings.ERC20MetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get erc20 abi: %w", err)
	}
	decimalsCalls, symbolsCalls := make([]bindings.Multicall3Call3, 0), make([]bindings.Multicall3Call3, 0)

	for _, reserve := range s.reservesList {
		symbolsCall, decimalsCall, err := getSymbolAndDecimalsMulticall3Call3(erc20Abi, reserve)
		if err != nil {
			return nil, fmt.Errorf("failed to get symbol and decimals: %w", err)
		}
		decimalsCalls = append(decimalsCalls, decimalsCall)
		symbolsCalls = append(symbolsCalls, symbolsCall)
	}

	var symbolsResults []bindings.Multicall3Result
	var decimalsResults []bindings.Multicall3Result
	// var results []bindings.Multicall3Result
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
	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to update reserves list and price: %w", err)
	}

	reserveInfos := make([]*ReserveInfo, 0)
	for i, reserve := range s.reservesList {
		reserveInfos = append(reserveInfos, &ReserveInfo{
			Address:  reserve.Hex(),
			Decimals: new(big.Int).SetBytes(decimalsResults[i].ReturnData),
			Symbol:   decodeSymbol(symbolsResults[i].ReturnData, erc20Abi),
		})
	}

	s.logger.Info("getReservesDecimalsAndSymbols", zap.Any("len", len(s.reservesList)), zap.Any("elapsed", time.Since(now)))
	return reserveInfos, nil
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
