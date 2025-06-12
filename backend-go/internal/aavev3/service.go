package aavev3

import (
	"context"
	"fmt"
	"liquidation-bot/config"
	"liquidation-bot/pkg/blockchain"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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
	chainClient          *blockchain.Client
	chainName            string
	contracts            *blockchain.Contracts
	dbWrapper            *DBWrapper
	liquidationInfoCache map[string]*LiquidationInfo
	ctx                  context.Context
}

// NewService 创建新的借入发现服务
func NewService(
	logger *zap.Logger,
	chainClient *blockchain.Client,
	chainName string,
	cfg *config.Config,
	dbWrapper *DBWrapper,
) (*Service, error) {
	logger = logger.Named(chainName)
	contracts, err := chainClient.GetContracts(chainName)
	if err != nil {
		logger.Error("failed to get contracts", zap.Error(err))
		return nil, err
	}
	s := &Service{
		logger:               logger,
		chainClient:          chainClient,
		chainName:            chainName,
		contracts:            contracts,
		dbWrapper:            dbWrapper,
		liquidationInfoCache: make(map[string]*LiquidationInfo),
		ctx:                  context.Background(),
	}
	return s, nil
}

// Initialize 初始化服务
func (s *Service) Initialize() error {
	var eg errgroup.Group
	// 1. 启动事件处理
	eg.Go(s.handleEvents)
	// 2. 启动价格流
	eg.Go(s.startPriceStream)
	// 3.启动健康因子检查
	eg.Go(s.startHealthFactorChecker)
	err := eg.Wait()
	if err != nil {
		s.logger.Error("failed to initialize service", zap.Error(err))
		return fmt.Errorf("failed to initialize service: %w", err)
	}
	return nil
}

// startHealthFactorChecker 启动健康因子检查
func (s *Service) startHealthFactorChecker() error {
	for {
		s.checkHealthFactorsBatch()
		time.Sleep(time.Minute)
	}
}

func (s *Service) getCallOpts() (*bind.CallOpts, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), CallTimeout)
	return &bind.CallOpts{
		Context: ctx,
	}, cancel
}

func (s *Service) getWatchOpts() *bind.WatchOpts {
	// todo - read start from db
	return &bind.WatchOpts{
		Start:   nil,
		Context: s.ctx,
	}
}
