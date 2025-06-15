package aavev3

import (
	"context"
	"fmt"
	"liquidation-bot/internal/models"
	"time"

	"go.uber.org/zap"
)

func (s *Service) startPriceStream(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done: %w", ctx.Err())
		case <-time.After(time.Second):
			callOpts, cancel := s.getCallOpts()
			defer cancel()

			prices, err := s.chain.GetContracts().PriceOracle.GetAssetsPrices(callOpts, s.reservesList)
			if err != nil {
				return fmt.Errorf("failed to get assets prices: %w", err)
			}
			tokenInfoMap, err := s.dbWrapper.GetTokenInfoMap(s.chain.ChainName)
			if err != nil {
				return fmt.Errorf("failed to get token infos: %w", err)
			}
			for i, price := range prices {
				tokenInfo := tokenInfoMap[s.reservesList[i].Hex()]
				if tokenInfo.Price.BigInt().Cmp(price) == 0 {
					continue
				}
				if err := s.dbWrapper.UpdateTokenPrice(s.chain.ChainName, tokenInfo.Address, models.NewBigInt(price)); err != nil {
					return fmt.Errorf("failed to update token price: %w", err)
				}
				s.logger.Info("update token price", zap.String("token", tokenInfo.Address), zap.String("newPrice", price.String()), zap.String("oldPrice", tokenInfo.Price.BigInt().String()))

				loans, err := s.dbWrapper.GetActiveLoansByToken(s.chain.ChainName, tokenInfo.Address)
				if err != nil {
					return fmt.Errorf("failed to get active loans by token: %w", err)
				}
				s.logger.Info("checking health factor for active loans by token", zap.Int("count", len(loans)))

				if err := s.checkHealthFactor(loans); err != nil {
					s.logger.Error("failed to check health factor", zap.Error(err))
					continue
				}
			}
		}
	}
}
