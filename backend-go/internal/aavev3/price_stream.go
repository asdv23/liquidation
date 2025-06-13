package aavev3

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) startPriceStream(ctx context.Context) error {
	for {
		// chainlink price stream
		// s.logger.Info("TODO - chainlink price stream")
		select {
		case <-ctx.Done():
			return fmt.Errorf("context done: %w", ctx.Err())
		case <-time.After(time.Minute):
		}
	}
}
