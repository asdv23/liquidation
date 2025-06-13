package blockchain

import (
	"liquidation-bot/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewClient(t *testing.T) {
	// 创建日志记录器
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	// 创建客户端
	chainConfigs := map[string]config.ChainConfig{
		"ethereum": {
			RPCURL: "ws://localhost:38546",
		},
	}
	privateKey := "0000000000000000000000000000000000000000000000000000000000000001"

	client, err := NewClient(logger, chainConfigs, privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}
