package aavev3

import (
	"time"
)

func (s *Service) startPriceStream() error {
	for {
		// chainlink price stream
		s.logger.Info("TODO - chainlink price stream")
		time.Sleep(time.Minute)
	}
}
