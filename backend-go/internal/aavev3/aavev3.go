package aavev3

import (
	"fmt"

	aavev3 "liquidation-bot/bindings/aavev3"

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
