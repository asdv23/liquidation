package aavev3

import (
	"fmt"
	bindings "liquidation-bot/bindings/common"
	"liquidation-bot/internal/models"
	"liquidation-bot/internal/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func (s *Service) createTokenInfoFromChain(asset string) (*models.Token, error) {
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

func getSymbolAndDecimalsMulticall3Call3(abi *abi.ABI, asset common.Address) (bindings.Multicall3Call3, bindings.Multicall3Call3, error) {
	symbolCallData, err := abi.Pack("symbol")
	if err != nil {
		return bindings.Multicall3Call3{}, bindings.Multicall3Call3{}, err
	}
	decimalsCallData, err := abi.Pack("decimals")
	if err != nil {
		return bindings.Multicall3Call3{}, bindings.Multicall3Call3{}, err
	}
	return bindings.Multicall3Call3{
			Target:   asset,
			CallData: symbolCallData,
		}, bindings.Multicall3Call3{
			Target:   asset,
			CallData: decimalsCallData,
		}, nil
}
