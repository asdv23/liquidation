package blockchain

import (
	"context"
	"fmt"
	"liquidation-bot/config"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Client 区块链客户端
type Client struct {
	sync.RWMutex

	ctx        context.Context
	logger     *zap.Logger
	privateKey string
	// 链
	chains map[string]*Chain
}

// NewClient 创建新的区块链客户端
func NewClient(logger *zap.Logger, chainConfigs map[string]config.ChainConfig, privateKey string) (*Client, error) {
	client := &Client{
		ctx:        context.Background(),
		logger:     logger,
		privateKey: privateKey,
		chains:     make(map[string]*Chain),
	}

	var eg errgroup.Group
	for chainName, chainConfig := range chainConfigs {
		chainName, chainConfig := chainName, chainConfig
		eg.Go(func() error {
			return client.newChain(chainName, chainConfig)
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("failed to init chains: %w", err)
	}

	return client, nil
}

// initChain 初始化链
func (c *Client) newChain(chainName string, chainConfig config.ChainConfig) error {
	chain, err := NewChain(c.ctx, c.logger, chainName, c.privateKey, chainConfig)
	if err != nil {
		return fmt.Errorf("failed to create chain: %w", err)
	}

	c.Lock()
	c.chains[chainName] = chain
	c.Unlock()

	return nil
}

// GetClient 获取客户端
func (c *Client) GetChain(chainName string) (*Chain, error) {
	c.RLock()
	defer c.RUnlock()

	chain, ok := c.chains[chainName]
	if !ok {
		return nil, fmt.Errorf("chain %s not found", chainName)
	}

	return chain, nil
}
