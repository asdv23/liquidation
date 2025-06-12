package core

import (
	"liquidation-bot/config"
	"liquidation-bot/internal/aavev3"
	"liquidation-bot/pkg/blockchain"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Services struct {
	DB     *gorm.DB
	Config *config.Config
	Logger *zap.Logger
	Chain  *blockchain.Client
	Aavev3 map[string]*aavev3.Service
}

func NewServices(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *Services {
	chainClient, err := blockchain.NewClient(logger, cfg.Chains, cfg.PrivateKey)
	if err != nil {
		logger.Error("failed to create chain client", zap.Error(err))
		os.Exit(1)
	}

	dbWrapper, err := aavev3.NewDBWrapper(db)
	if err != nil {
		logger.Error("failed to create db wrapper", zap.Error(err))
		os.Exit(1)
	}

	aavev3Services := make(map[string]*aavev3.Service)
	for chain := range cfg.Chains {
		aavev3Service, err := aavev3.NewService(logger, chainClient, chain, cfg, dbWrapper)
		if err != nil {
			logger.Error("failed to create aavev3 service", zap.Error(err))
			os.Exit(1)
		}

		go func() {
			retries := 0
			for {
				if err := aavev3Service.Initialize(); err == nil {
					break
				}
				retries++
				if retries > 3 {
					logger.Error("failed to initialize aavev3 service", zap.Error(err))
					os.Exit(1)
				}
				time.Sleep(5 * time.Second)
			}
		}()
		aavev3Services[chain] = aavev3Service
	}

	return &Services{
		DB:     db,
		Config: cfg,
		Logger: logger,
		Chain:  chainClient,
		Aavev3: aavev3Services,
	}
}
