package config

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config 配置
type Config struct {
	// 私钥
	PrivateKey string `json:"privateKey"`

	// 服务器配置
	Server ServerConfig `json:"server"`

	// 数据库配置
	Database DatabaseConfig `json:"database"`

	// 区块链配置
	Chains map[string]ChainConfig `json:"chains"`
}

// ChainConfig 链配置
type ChainConfig struct {
	RPCURL             string            `json:"rpcUrl"`
	PoolAddress        string            `json:"poolAddress"`
	GasLimit           uint64            `json:"gasLimit"`
	GasPriceMultiplier float64           `json:"gasPriceMultiplier"`
	ContractAddresses  map[string]string `json:"contractAddresses"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DriverName     string `json:"driverName"`
	DataSourceName string `json:"dataSourceName"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `json:"port"`
}

// NewConfig 创建新的配置
func NewConfig(logger *zap.Logger, path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 创建用于打印的配置副本，排除私钥
	printConfig := config
	printConfig.PrivateKey = "[REDACTED]"
	logger.Info("config", zap.Any("config", printConfig))

	// 验证配置
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// validate 验证配置
func (c *Config) validate() error {
	if c.PrivateKey == "" {
		return fmt.Errorf("private key not configured")
	}

	// 验证链配置
	if len(c.Chains) == 0 {
		return fmt.Errorf("no chains configured")
	}

	for chainName, chain := range c.Chains {
		if chain.RPCURL == "" {
			return fmt.Errorf("RPC URL not configured for chain %s", chainName)
		}
		if chain.GasLimit == 0 {
			return fmt.Errorf("gas limit not configured for chain %s", chainName)
		}
		if chain.GasPriceMultiplier <= 0 {
			return fmt.Errorf("gas price multiplier must be positive for chain %s", chainName)
		}
	}

	// 验证数据库配置
	if c.Database.DriverName == "" {
		return fmt.Errorf("database driver name not configured")
	}
	if c.Database.DataSourceName == "" {
		return fmt.Errorf("database data source name not configured")
	}

	return nil
}
