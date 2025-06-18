package aavev3

import (
	"context"
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	"liquidation-bot/pkg/blockchain"
	"math/big"

	"go.uber.org/zap"
)

func (s *Service) handleEvents(ctx context.Context) error {
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
		case <-ctx.Done():
			return fmt.Errorf("context done: %w", ctx.Err())
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
	s.logger.Info(" - ", zap.Any("Reserve", event.Reserve.Hex()))
	s.logger.Info(" - ", zap.Any("User", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("OnBehalfOf", event.OnBehalfOf.Hex()))
	s.infoAmount("Amount", event.Reserve.Hex(), event.Amount)
	s.logger.Info(" - ", zap.Any("InterestRateMode", event.InterestRateMode))
	s.logger.Info(" - ", zap.Any("BorrowRate", event.BorrowRate))
	s.logger.Info(" - ", zap.Any("ReferralCode", event.ReferralCode))

	if err := s.updateLoan(event.User.Hex()); err != nil {
		s.logger.Error("failed to update loan", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleRepayEvent(event *aavev3.PoolRepay) {
	s.logger.Info("repay event 😢", zap.Any("user", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("Reserve", event.Reserve.Hex()))
	s.logger.Info(" - ", zap.Any("User", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("Repayer", event.Repayer.Hex()))
	s.infoAmount("Amount", event.Reserve.Hex(), event.Amount)
	s.logger.Info(" - ", zap.Any("UseATokens", event.UseATokens))

	if err := s.updateLoan(event.User.Hex()); err != nil {
		s.logger.Error("failed to update loan", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleSupplyEvent(event *aavev3.PoolSupply) {
	s.logger.Info("supply event 👀", zap.Any("user", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("Reserve", event.Reserve.Hex()))
	s.logger.Info(" - ", zap.Any("User", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("OnBehalfOf", event.OnBehalfOf.Hex()))
	s.infoAmount("Amount", event.Reserve.Hex(), event.Amount)
	s.logger.Info(" - ", zap.Any("ReferralCode", event.ReferralCode))

	if err := s.updateLoan(event.User.Hex()); err != nil {
		s.logger.Error("failed to update loan", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleWithdrawEvent(event *aavev3.PoolWithdraw) {
	s.logger.Info("withdraw event 🤨", zap.Any("user", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("Reserve", event.Reserve.Hex()))
	s.logger.Info(" - ", zap.Any("User", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("To", event.To.Hex()))
	s.infoAmount("Amount", event.Reserve.Hex(), event.Amount)

	if err := s.updateLoan(event.User.Hex()); err != nil {
		s.logger.Error("failed to update loan", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleLiquidationEvent(event *aavev3.PoolLiquidationCall) {
	s.logger.Info("liquidation event 🤩", zap.Any("user", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("CollateralAsset", event.CollateralAsset.Hex()))
	s.logger.Info(" - ", zap.Any("DebtAsset", event.DebtAsset.Hex()))
	s.logger.Info(" - ", zap.Any("User", event.User.Hex()))
	s.infoAmount("DebtToCover", event.DebtAsset.Hex(), event.DebtToCover)
	s.infoAmount("LiquidatedCollateralAmount", event.CollateralAsset.Hex(), event.LiquidatedCollateralAmount)
	s.logger.Info(" - ", zap.Any("Liquidator", event.Liquidator.Hex()))
	s.logger.Info(" - ", zap.Any("ReceiveAToken", event.ReceiveAToken))

	if err := s.updateLoan(event.User.Hex()); err != nil {
		s.logger.Error("failed to update loan", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) updateLoan(user string) error {
	loan, err := s.dbWrapper.CreateOrUpdateActiveLoan(s.chain.ChainName, user)
	if err != nil {
		return fmt.Errorf("failed to create or update loan: %w", err)
	}

	// sync health factor
	if err := s.syncHealthFactorForUser(user, loan); err != nil {
		s.logger.Error("failed to sync health factor for user", zap.Error(err), zap.String("user", user))
	}

	return nil
}

func (s *Service) infoAmount(msg, reserve string, amount *big.Int) {
	if tokenInfo, err := s.dbWrapper.GetTokenInfo(s.chain.ChainName, reserve); err != nil {
		s.logger.Info(" - ", zap.Any(msg, amount.String()), zap.Error(err))
	} else {
		s.logger.Info(" - ", zap.Any(msg, formatAmount(amount, tokenInfo.Decimals.BigInt())+" "+tokenInfo.Symbol))
		s.logger.Info(" - ", zap.Any(msg+"USD", amountToUSD(amount, tokenInfo.Decimals.BigInt(), tokenInfo.Price.BigInt())))
		s.logger.Info(" - ", zap.Any("Price", big.NewFloat(0).Quo(big.NewFloat(0).SetInt((*big.Int)(tokenInfo.Price)), USD_DECIMALS)))
	}
}
