package aavev3

import (
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"

	"go.uber.org/zap"
)

func (s *Service) handleEvents() error {
	s.logger.Info("start to handle events")
	opts := s.getWatchOpts()

	borrowSink := make(chan *aavev3.PoolBorrow)
	borrowSub, err := s.contracts.AaveV3Pool.WatchBorrow(opts, borrowSink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch borrow events: %w", err)
	}
	defer borrowSub.Unsubscribe()

	repaySink := make(chan *aavev3.PoolRepay)
	repaySub, err := s.contracts.AaveV3Pool.WatchRepay(opts, repaySink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch repay events: %w", err)
	}
	defer repaySub.Unsubscribe()

	supplySink := make(chan *aavev3.PoolSupply)
	supplySub, err := s.contracts.AaveV3Pool.WatchSupply(opts, supplySink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch supply events: %w", err)
	}
	defer supplySub.Unsubscribe()

	withdrawSink := make(chan *aavev3.PoolWithdraw)
	withdrawSub, err := s.contracts.AaveV3Pool.WatchWithdraw(opts, withdrawSink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch withdraw events: %w", err)
	}
	defer withdrawSub.Unsubscribe()

	liquidationSink := make(chan *aavev3.PoolLiquidationCall)
	liquidationSub, err := s.contracts.AaveV3Pool.WatchLiquidationCall(opts, liquidationSink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch liquidation events: %w", err)
	}
	defer liquidationSub.Unsubscribe()

	for {
		select {
		case borrowEvent := <-borrowSink:
			s.handleBorrowEvent(borrowEvent)
		case repayEvent := <-repaySink:
			s.handleRepayEvent(repayEvent)
		case supplyEvent := <-supplySink:
			s.handleSupplyEvent(supplyEvent)
		case withdrawEvent := <-withdrawSink:
			s.handleWithdrawEvent(withdrawEvent)
		case liquidationEvent := <-liquidationSink:
			s.handleLiquidationEvent(liquidationEvent)
		case <-s.ctx.Done():
			return fmt.Errorf("context done: %w", s.ctx.Err())
		case err := <-borrowSub.Err():
			return fmt.Errorf("failed to watch borrow events: %w", err)
		case err := <-repaySub.Err():
			return fmt.Errorf("failed to watch repay events: %w", err)
		case err := <-supplySub.Err():
			return fmt.Errorf("failed to watch supply events: %w", err)
		case err := <-withdrawSub.Err():
			return fmt.Errorf("failed to watch withdraw events: %w", err)
		case err := <-liquidationSub.Err():
			return fmt.Errorf("failed to watch liquidation events: %w", err)
		}
	}
}

func (s *Service) handleBorrowEvent(event *aavev3.PoolBorrow) {
	s.logger.Info("borrow event", zap.Any("event", event))
	// create or update loan
	if err := s.dbWrapper.CreateOrUpdateActiveLoan(s.chainName, event.User.Hex()); err != nil {
		s.logger.Error("failed to create or update loan", zap.Error(err), zap.String("user", event.User.Hex()))
	}

	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleRepayEvent(event *aavev3.PoolRepay) {
	s.logger.Info("repay event", zap.Any("event", event))
	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleSupplyEvent(event *aavev3.PoolSupply) {
	s.logger.Info("supply event", zap.Any("event", event))

	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleWithdrawEvent(event *aavev3.PoolWithdraw) {
	s.logger.Info("withdraw event", zap.Any("event", event))

	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleLiquidationEvent(event *aavev3.PoolLiquidationCall) {
	s.logger.Info("liquidation event", zap.Any("event", event))

	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}
