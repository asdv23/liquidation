package database

import (
	"fmt"

	"liquidation-bot/config"
	"liquidation-bot/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewConnection(cfg config.DatabaseConfig) (db *gorm.DB, err error) {
	switch cfg.DriverName {
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DataSourceName), &gorm.Config{})
	default:
		db, err = gorm.Open(sqlite.Open(cfg.DataSourceName), &gorm.Config{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移数据库结构
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Loan{},
		&models.Token{},
	)
}
