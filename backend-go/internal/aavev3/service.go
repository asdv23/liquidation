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
// 4. 持续清算待清算用户，直到健康因子大于 1

type Service struct {
	sync.RWMutex

	logger             *zap.Logger
	chain              *blockchain.Chain
	dbWrapper          *DBWrapper
	toBeLiquidatedChan chan string
	reservesList       []common.Address
	liquidatingUsers   map[string]struct{}
}

// NewService 创建新的借入发现服务
func NewService(
	chain *blockchain.Chain,
	dbWrapper *DBWrapper,
) (*Service, error) {
	s := &Service{
		logger:             chain.Logger.Named("aavev3"),
		chain:              chain,
		dbWrapper:          dbWrapper,
		toBeLiquidatedChan: make(chan string, 100),
		liquidatingUsers:   make(map[string]struct{}),
		reservesList:       make([]common.Address, 0),
	}
	go s.Initialize()
	return s, nil
}

// Initialize 初始化服务
func (s *Service) Initialize() {
	err := retry.Do(func() error {
		// 1. 更新 ReserveList 和价格
		if err := s.updateReservesListAndPrice(); err != nil {
			return fmt.Errorf("failed to update reserves list and price: %w", err)
		}

		// 2. 处理没有清算信息的贷款
		noLiquidationInfoLoans, err := s.dbWrapper.GetNoLiquidationInfoLoans(s.chain.Ctx, s.chain.ChainName)
		if err != nil {
			return fmt.Errorf("failed to get loans with no liquidation information: %w", err)
		}

		s.logger.Info("found loans with no liquidation information", zap.Int("count", len(noLiquidationInfoLoans)))
		if err := s.syncHealthFactorForLoans(noLiquidationInfoLoans); err != nil {
			return fmt.Errorf("failed to sync health factor for loans: %w", err)
		}

		// 3. 启动后台任务
		eg, ctx := errgroup.WithContext(s.chain.Ctx)
		// 1. 启动事件处理
		eg.Go(func() error {
			return s.handleEvents(ctx)
		})
		// 2. 启动价格同步
		eg.Go(func() error {
			return s.startSyncPricesForReserveList(ctx)
		})
		// 3.启动健康因子同步 - useless
		eg.Go(func() error {
			return s.startSyncHealthFactorForAllActiveLoans(ctx)
		})
		// 4. 启动清算检查
		eg.Go(func() error {
			return s.startLiquidation(ctx)
		})
		// 5. 启动重同步
		eg.Go(func() error {
			return s.resync(ctx)
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

// startSyncHealthFactorForAllActiveLoans 启动健康因子同步
func (s *Service) startSyncHealthFactorForAllActiveLoans(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(12 * time.Hour):
			activeLoans, err := s.dbWrapper.ChainActiveLoans(s.chain.ChainName)
			if err != nil {
				s.logger.Error("failed to get active loans", zap.Error(err))
				continue
			}

			if err := s.updateReservesListAndPrice(); err != nil {
				s.logger.Error("failed to update reserves list and price", zap.Error(err))
				continue
			}

			s.logger.Info("sync health factor for all active loans", zap.Int("count", len(activeLoans)))
			if err := s.syncHealthFactorForLoans(activeLoans); err != nil {
				s.logger.Error("failed to sync health factor", zap.Error(err))
				continue
			}
		}
	}
}

func (s *Service) startLiquidation(ctx context.Context) error {
	go func() {
		// check if there are any liquidation loans
		loans, err := s.dbWrapper.GetLiquidationLoans(ctx, s.chain.ChainName)
		if err != nil {
			s.logger.Error("failed to get liquidation infos", zap.Error(err))
			return
		}
		s.logger.Info("found liquidation loans", zap.Int("count", len(loans)))
		for _, loan := range loans {
			s.toBeLiquidatedChan <- loan.User
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case user := <-s.toBeLiquidatedChan:
			s.logger.Info("liquidating loan", zap.String("user", user))
			if _, ok := s.liquidatingUsers[user]; ok {
				s.logger.Info("already liquidating", zap.String("user", user))
				continue
			}
			s.liquidatingUsers[user] = struct{}{}

			go func(user string) {
				liquidationCtx, cancel := context.WithCancel(ctx)
				defer cancel()

				loan, err := s.dbWrapper.GetLoan(ctx, s.chain.ChainName, user)
				if err != nil {
					s.logger.Error("failed to get loan", zap.Error(err))
					return
				}

				s.logger.Info("executing liquidation", zap.String("user", user), zap.Float64("healthFactor", loan.HealthFactor))
				if err := s.executeLiquidation(liquidationCtx, loan); err != nil {
					s.logger.Error("failed to execute liquidation", zap.Error(err))
					return
				}

				for {
					select {
					case <-ctx.Done():
						return
					case <-time.After(400 * time.Millisecond):
						loan, err := s.dbWrapper.GetLoan(ctx, s.chain.ChainName, user)
						if err != nil {
							s.logger.Error("failed to get loan", zap.Error(err))
							continue
						}
						if loan.HealthFactor >= 1 {
							delete(s.liquidatingUsers, user)
							s.logger.Info("health factor above liquidation threshold, exit liquidation loop", zap.String("user", user))
							return
						}
					}
				}
			}(user)
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
