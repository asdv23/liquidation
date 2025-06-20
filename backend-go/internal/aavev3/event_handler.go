package aavev3

import (
	"context"
	"fmt"
	aavev3 "liquidation-bot/bindings/aavev3"
	"liquidation-bot/pkg/blockchain"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/event"
	"go.uber.org/zap"
)

func watchWithReconnect[T any](
	ctx context.Context,
	logger *zap.Logger,
	name string,
	watchFunc func(chan T) (event.Subscription, error),
	handler func(T),
) {
	logger = logger.Named(name)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		sink := make(chan T, 100)
		sub, err := watchFunc(sink)
		if err != nil {
			logger.Error("Watch failed", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Info("Subscribed to event")

	handleLoop:
		for {
			select {
			case <-ctx.Done():
				sub.Unsubscribe()
				return
			case err := <-sub.Err():
				logger.Warn("Subscription dropped", zap.Error(err))
				sub.Unsubscribe()
				time.Sleep(5 * time.Second)
				break handleLoop
			case evt := <-sink:
				handler(evt)
			}
		}
	}
}

func (s *Service) handleEvents(ctx context.Context) error {
	s.logger.Info("start to handle events", zap.String("aavev3_pool", s.chain.GetContracts().Addresses[blockchain.ContractTypeAaveV3Pool].Hex()))
	opts := s.getWatchOpts()

	// borrow
	go watchWithReconnect(ctx, s.logger, "borrow", func(sink chan *aavev3.PoolBorrow) (event.Subscription, error) {
		return s.chain.GetContracts().AaveV3Pool.WatchBorrow(opts, sink, nil, nil, nil)
	}, s.handleBorrowEvent)

	// repay
	go watchWithReconnect(ctx, s.logger, "repay", func(sink chan *aavev3.PoolRepay) (event.Subscription, error) {
		return s.chain.GetContracts().AaveV3Pool.WatchRepay(opts, sink, nil, nil, nil)
	}, s.handleRepayEvent)

	// supply
	go watchWithReconnect(ctx, s.logger, "supply", func(sink chan *aavev3.PoolSupply) (event.Subscription, error) {
		return s.chain.GetContracts().AaveV3Pool.WatchSupply(opts, sink, nil, nil, nil)
	}, s.handleSupplyEvent)

	// withdraw
	go watchWithReconnect(ctx, s.logger, "withdraw", func(sink chan *aavev3.PoolWithdraw) (event.Subscription, error) {
		return s.chain.GetContracts().AaveV3Pool.WatchWithdraw(opts, sink, nil, nil, nil)
	}, s.handleWithdrawEvent)

	// liquidation
	go watchWithReconnect(ctx, s.logger, "liquidation", func(sink chan *aavev3.PoolLiquidationCall) (event.Subscription, error) {
		return s.chain.GetContracts().AaveV3Pool.WatchLiquidationCall(opts, sink, nil, nil, nil)
	}, s.handleLiquidationEvent)

	return nil
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

	if err := s.resyncLoan(event.User.Hex()); err != nil {
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

	if err := s.resyncLoan(event.User.Hex()); err != nil {
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

	if err := s.resyncLoan(event.User.Hex()); err != nil {
		s.logger.Error("failed to update loan", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) handleWithdrawEvent(event *aavev3.PoolWithdraw) {
	s.logger.Info("withdraw event 🤨", zap.Any("user", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("Reserve", event.Reserve.Hex()))
	s.logger.Info(" - ", zap.Any("User", event.User.Hex()))
	s.logger.Info(" - ", zap.Any("To", event.To.Hex()))
	s.infoAmount("Amount", event.Reserve.Hex(), event.Amount)

	if err := s.resyncLoan(event.User.Hex()); err != nil {
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

	if err := s.resyncLoan(event.User.Hex()); err != nil {
		s.logger.Error("failed to update loan", zap.Error(err), zap.String("user", event.User.Hex()))
	}
}

func (s *Service) resyncLoan(user string) error {
	// resync set to true, wait 5min to sync for all
	_, err := s.dbWrapper.CreateOrUpdateActiveLoan(s.chain.ChainName, user)
	if err != nil {
		return fmt.Errorf("failed to create or update loan: %w", err)
	}

	// sync health factor
	// if err := s.syncHealthFactorForUser(user, loan); err != nil {
	// 	s.logger.Error("failed to sync health factor for user", zap.Error(err), zap.String("user", user))
	// }

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
