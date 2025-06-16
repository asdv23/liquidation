package core

import (
	"liquidation-bot/config"
	"liquidation-bot/internal/aavev3"
	"liquidation-bot/pkg/blockchain"
	"os"
	"time"

	"github.com/avast/retry-go/v4"
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
	var chainClient *blockchain.Client
	if err := retry.Do(func() error {
		client, err := blockchain.NewClient(logger, cfg.Chains, cfg.PrivateKey)
		if err != nil {
			logger.Error("failed to create chain client", zap.Error(err))
			return err
		}
		chainClient = client
		return nil
	}, retry.MaxDelay(60*time.Second), retry.Delay(1*time.Second), retry.DelayType(retry.BackOffDelay)); err != nil {
		logger.Error("failed to create chain client", zap.Error(err))
		os.Exit(1)
	}

	dbWrapper, err := aavev3.NewDBWrapper(db)
	if err != nil {
		logger.Error("failed to create db wrapper", zap.Error(err))
		os.Exit(1)
	}

	aavev3Services := make(map[string]*aavev3.Service)
	for chainName := range cfg.Chains {
		chain, err := chainClient.GetChain(chainName)
		if err != nil {
			logger.Error("failed to get chain", zap.Error(err))
			os.Exit(1)
		}

		aavev3Service, err := aavev3.NewService(chain, dbWrapper)
		if err != nil {
			logger.Error("failed to create aavev3 service", zap.Error(err))
			os.Exit(1)
		}
		aavev3Services[chainName] = aavev3Service
	}

	return &Services{
		DB:     db,
		Config: cfg,
		Logger: logger,
		Chain:  chainClient,
		Aavev3: aavev3Services,
	}
}
