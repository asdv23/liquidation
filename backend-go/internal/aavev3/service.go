package aavev3

import (
	"bytes"
	"context"
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	bindings "liquidation-bot/bindings/common"
	"liquidation-bot/internal/utils"
	"liquidation-bot/pkg/blockchain"
	"math/big"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	CallTimeout = 5 * time.Second
)

// Service aavev3服务
//
// 1. 监听所有事件
// Borrow: 创建活跃贷款或更新健康因子
// Repay: 更新健康因子
// Supply: 更新健康因子
// Withdraw: 更新健康因子
// Liquidation: 记录清算信息并更新健康因子
// 2. 根据 ChianLink 价格流，对于受影响的活跃贷款，更新健康因子
// 3. 每隔一分钟检查所有活跃贷款健康因子，如果发生变化，应该告警

type Service struct {
	sync.RWMutex

	logger               *zap.Logger
	chain                *blockchain.Chain
	dbWrapper            *DBWrapper
	liquidationInfoCache map[string]*LiquidationInfo
	reservesList         []common.Address
}

// NewService 创建新的借入发现服务
func NewService(
	chain *blockchain.Chain,
	dbWrapper *DBWrapper,
) (*Service, error) {
	s := &Service{
		logger:               chain.Logger.Named("aavev3"),
		chain:                chain,
		dbWrapper:            dbWrapper,
		liquidationInfoCache: make(map[string]*LiquidationInfo),
		reservesList:         make([]common.Address, 0),
	}
	chain.Register(s.Initialize)
	go s.Initialize()
	return s, nil
}

// Initialize 初始化服务
func (s *Service) Initialize() {
	err := retry.Do(func() error {
		eg, ctx := errgroup.WithContext(s.chain.Ctx)
		// 1. 启动事件处理
		eg.Go(func() error {
			return s.handleEvents(ctx)
		})
		// 2. 启动价格流
		eg.Go(func() error {
			return s.startPriceStream(ctx)
		})
		// 3.启动健康因子检查
		eg.Go(func() error {
			return s.startHealthFactorChecker(ctx)
		})
		// 4. 启动储备检查
		eg.Go(func() error {
			return s.startReservesChecker(ctx)
		})
		if err := eg.Wait(); err != nil {
			s.logger.Error("failed to initialize service", zap.Error(err))
			return fmt.Errorf("failed to initialize service: %w", err)
		}
		return nil
	}, retry.Attempts(3), retry.Delay(10*time.Second), retry.MaxDelay(1*time.Minute))
	if err != nil {
		s.logger.Fatal("failed to initialize service", zap.Error(err))
	}
}

// startHealthFactorChecker 启动健康因子检查
func (s *Service) startHealthFactorChecker(ctx context.Context) error {
	for {
		s.checkHealthFactorsBatch()

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Minute):
		}
	}
}

// startReservesChecker 启动储备检查
func (s *Service) startReservesChecker(ctx context.Context) error {
	for {
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
			decimalsCallData, err := erc20Abi.Pack("decimals")
			if err != nil {
				return fmt.Errorf("failed to pack decimals call: %w", err)
			}
			symbolsCallData, err := erc20Abi.Pack("symbol")
			if err != nil {
				return fmt.Errorf("failed to pack symbol call: %w", err)
			}

			callData, err := abi.Pack("getAssetPrice", reserve)
			if err != nil {
				return fmt.Errorf("failed to pack get asset price call: %w", err)
			}

			decimalsCalls = append(decimalsCalls, bindings.Multicall3Call3{
				Target:   reserve,
				CallData: decimalsCallData,
			})

			symbolsCalls = append(symbolsCalls, bindings.Multicall3Call3{
				Target:   reserve,
				CallData: symbolsCallData,
			})

			getReservesPriceCalls = append(getReservesPriceCalls, bindings.Multicall3Call3{
				Target:   target,
				CallData: callData,
			})
		}

		symbolsResults, err := utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, symbolsCalls)
		if err != nil {
			return fmt.Errorf("failed to get symbols: %w", err)
		}
		decimalsResults, err := utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, decimalsCalls)
		if err != nil {
			return fmt.Errorf("failed to get decimals: %w", err)
		}

		results, err := utils.Aggregate3(callOpts, s.chain.GetContracts().Multicall3, getReservesPriceCalls)
		if err != nil {
			return fmt.Errorf("failed to get reserves price: %w", err)
		}
		for i, result := range results {
			price := new(big.Int).SetBytes(result.ReturnData)
			decimals := new(big.Int).SetBytes(decimalsResults[i].ReturnData)
			symbol := decodeSymbol(symbolsResults[i].ReturnData, erc20Abi)
			s.logger.Info("reserves price", zap.String("reserve", reservesList[i].Hex()), zap.String("price", price.String()), zap.String("decimals", decimals.String()), zap.String("symbol", symbol))
			if _, err := s.dbWrapper.AddTokenInfo(s.chain.ChainName, reservesList[i].Hex(), symbol, decimals, price); err != nil {
				return fmt.Errorf("failed to add token info: %w", err)
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Minute):
		}
	}
}

func decodeSymbol(returnData []byte, erc20Abi *abi.ABI) string {
	var symbol string
	err := erc20Abi.UnpackIntoInterface(&symbol, "symbol", returnData)
	if err != nil {
		return string(bytes.TrimRight(returnData, "\x00"))
	}

	return symbol
}

func (s *Service) getCallOpts() (*bind.CallOpts, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(s.chain.Ctx, CallTimeout)
	return &bind.CallOpts{
		Context: ctx,
	}, cancel
}

func (s *Service) getWatchOpts() *bind.WatchOpts {
	// todo - read start from db
	return &bind.WatchOpts{
		Start:   nil,
		Context: s.chain.Ctx,
	}
}
