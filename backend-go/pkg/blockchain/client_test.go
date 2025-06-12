package blockchain

import (
	"liquidation-bot/config"
	"testing"

	"github.com/ethereum/go-ethereum/common"
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

func TestGetClient(t *testing.T) {
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

	// 获取客户端
	ethClient, err := client.GetClient("ethereum")
	assert.NoError(t, err)
	assert.NotNil(t, ethClient)
}

func TestGetAuth(t *testing.T) {
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

	// 获取认证
	auth, err := client.GetAuth("ethereum")
	assert.NoError(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, common.HexToAddress("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf"), auth.From)
}

// func TestMulticall(t *testing.T) {
// 	// 创建日志记录器
// 	logger, err := zap.NewDevelopment()
// 	assert.NoError(t, err)

// 	// 创建客户端
// 	chainConfigs := map[string]config.ChainConfig{
// 		"ethereum": {
// 			RPCURL: "ws://localhost:38546",
// 			ContractAddresses: map[string]string{
// 				"multicall3": "0xcA11bde05977b3631167028862bE2a173976CA11",
// 			},
// 		},
// 	}
// 	privateKey := "0000000000000000000000000000000000000000000000000000000000000001"

// 	client, err := NewClient(logger, chainConfigs, privateKey)
// 	assert.NoError(t, err)

// 	// 获取合约
// 	contracts, err := client.GetContracts("ethereum")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, contracts)
// 	assert.NotNil(t, contracts.Multicall3)

// 	// 获取认证
// 	auth, err := client.GetAuth("ethereum")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, auth)

// 	// 创建批量调用
// 	calls := []bindings.Multicall3Call{
// 		{
// 			Target:   common.HexToAddress("0x123"),
// 			CallData: []byte{1, 2, 3},
// 		},
// 	}

// 	// 执行批量调用
// 	results, err := contracts.Multicall3.Aggregate(auth, calls)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, results)
// }
