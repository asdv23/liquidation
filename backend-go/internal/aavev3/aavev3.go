package aavev3

import (
	"fmt"

	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"
	"liquidation-bot/internal/utils"
	"liquidation-bot/pkg/blockchain"

	"github.com/ethereum/go-ethereum/common"
)

func (s *Service) getUserAccountData(user string) (*UserAccountData, error) {
	// 准备调用参数
	opts, cancel := s.getCallOpts()
	defer cancel()

	// 调用合约
	result, err := s.chain.GetContracts().AaveV3Pool.GetUserAccountData(opts, common.HexToAddress(user))
	if err != nil {
		return nil, fmt.Errorf("failed to call getUserAccountData: %w", err)
	}

	// 解析结果
	data := &UserAccountData{
		TotalCollateralBase:         result.TotalCollateralBase,
		TotalDebtBase:               result.TotalDebtBase,
		AvailableBorrowsBase:        result.AvailableBorrowsBase,
		CurrentLiquidationThreshold: result.CurrentLiquidationThreshold,
		Ltv:                         result.Ltv,
		HealthFactor:                result.HealthFactor,
	}

	return data, nil
}

func (s *Service) getUserReserveData(asset string, user string) (*UserReserveData, error) {
	// 准备调用参数
	opts, cancel := s.getCallOpts()
	defer cancel()

	// 调用合约
	result, err := s.chain.GetContracts().DataProvider.GetUserReserveData(opts, common.HexToAddress(asset), common.HexToAddress(user))
	if err != nil {
		return nil, fmt.Errorf("failed to call getUserReserveData: %w", err)
	}

	data := &UserReserveData{
		CurrentATokenBalance:     result.CurrentATokenBalance,
		CurrentStableDebt:        result.CurrentStableDebt,
		CurrentVariableDebt:      result.CurrentVariableDebt,
		PrincipalStableDebt:      result.PrincipalStableDebt,
		ScaledVariableDebt:       result.ScaledVariableDebt,
		StableBorrowRate:         result.StableBorrowRate,
		LiquidityRate:            result.LiquidityRate,
		StableRateLastUpdated:    result.StableRateLastUpdated,
		UsageAsCollateralEnabled: result.UsageAsCollateralEnabled,
	}

	return data, nil
}

func (s *Service) getReservesList() ([]common.Address, error) {
	// 准备调用参数
	opts, cancel := s.getCallOpts()
	defer cancel()

	// 调用合约
	result, err := s.chain.GetContracts().AaveV3Pool.GetReservesList(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to call getReservesList: %w", err)
	}

	return result, nil
}

func (s *Service) getUserConfiguration(user string) (*aavev3.DataTypesUserConfigurationMap, error) {
	// 准备调用参数
	opts, cancel := s.getCallOpts()
	defer cancel()

	// 调用合约
	result, err := s.chain.GetContracts().AaveV3Pool.GetUserConfiguration(opts, common.HexToAddress(user))
	if err != nil {
		return nil, fmt.Errorf("failed to call getUserConfiguration: %w", err)
	}

	return &result, nil
}

func (s *Service) getUserConfigurationForBatch(users []string) ([]*aavev3.DataTypesUserConfigurationMap, error) {
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
	for _, result := range results {
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

	return userConfigs, nil
}
