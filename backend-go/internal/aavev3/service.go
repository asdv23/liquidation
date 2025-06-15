package aavev3

import (
	"context"
	"fmt"
	"liquidation-bot/pkg/blockchain"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
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
		s.syncHealthFactorsForAll()

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(5 * time.Minute):
		}
	}
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
