package aavev3

import (
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	"liquidation-bot/pkg/blockchain"

	"go.uber.org/zap"
)

func (s *Service) handleEvents() error {
	s.logger.Info("start to handle events", zap.String("aavev3_pool", s.chain.GetContracts().Addresses[blockchain.ContractTypeAaveV3Pool].Hex()))
	opts := s.getWatchOpts()

	borrowSink := make(chan *aavev3.PoolBorrow, 100)
	borrowSub, err := s.chain.GetContracts().AaveV3Pool.WatchBorrow(opts, borrowSink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch borrow events: %w", err)
	}
	defer borrowSub.Unsubscribe()

	repaySink := make(chan *aavev3.PoolRepay, 100)
	repaySub, err := s.chain.GetContracts().AaveV3Pool.WatchRepay(opts, repaySink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch repay events: %w", err)
	}
	defer repaySub.Unsubscribe()

	supplySink := make(chan *aavev3.PoolSupply, 100)
	supplySub, err := s.chain.GetContracts().AaveV3Pool.WatchSupply(opts, supplySink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch supply events: %w", err)
	}
	defer supplySub.Unsubscribe()

	withdrawSink := make(chan *aavev3.PoolWithdraw, 100)
	withdrawSub, err := s.chain.GetContracts().AaveV3Pool.WatchWithdraw(opts, withdrawSink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch withdraw events: %w", err)
	}
	defer withdrawSub.Unsubscribe()

	liquidationSink := make(chan *aavev3.PoolLiquidationCall, 100)
	liquidationSub, err := s.chain.GetContracts().AaveV3Pool.WatchLiquidationCall(opts, liquidationSink, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to watch liquidation events: %w", err)
	}
	defer liquidationSub.Unsubscribe()

	for {
		select {
		case <-s.chain.Ctx.Done():
			return fmt.Errorf("context done: %w", s.chain.Ctx.Err())
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
		}
	}
}

func (s *Service) handleBorrowEvent(event *aavev3.PoolBorrow) {
	s.logger.Info("borrow event 😄", zap.Any("user", event.User.Hex()))
	// create or update loan
	if err := s.dbWrapper.CreateOrUpdateActiveLoan(s.chain.ChainName, event.User.Hex()); err != nil {
		s.logger.Error("failed to create or update loan", zap.Error(err), zap.String("user", event.User.Hex()))
	}

	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleRepayEvent(event *aavev3.PoolRepay) {
	s.logger.Info("repay event 😢", zap.Any("user", event.User.Hex()))
	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleSupplyEvent(event *aavev3.PoolSupply) {
	s.logger.Info("supply event 👀", zap.Any("user", event.User.Hex()))

	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleWithdrawEvent(event *aavev3.PoolWithdraw) {
	s.logger.Info("withdraw event 🤨", zap.Any("user", event.User.Hex()))

	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleLiquidationEvent(event *aavev3.PoolLiquidationCall) {
	s.logger.Info("liquidation event 🤩", zap.Any("user", event.User.Hex()))

	// update health factor
	if err := s.updateHealthFactorViaEvent(event.User.Hex()); err != nil {
		s.logger.Error("failed to update health factor", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}
