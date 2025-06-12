package blockchain

import (
	"context"
	"fmt"
	"liquidation-bot/config"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Client 区块链客户端
type Client struct {
	sync.RWMutex

	logger       *zap.Logger
	clients      map[string]*ethclient.Client
	auths        map[string]*bind.TransactOpts
	chainConfigs map[string]config.ChainConfig
	privateKey   string
	reconnect    chan struct{}
	// 合约缓存
	contracts map[string]*Contracts
}

// NewClient 创建新的区块链客户端
func NewClient(logger *zap.Logger, chainConfigs map[string]config.ChainConfig, privateKey string) (*Client, error) {
	client := &Client{
		logger:       logger,
		clients:      make(map[string]*ethclient.Client),
		auths:        make(map[string]*bind.TransactOpts),
		chainConfigs: chainConfigs,
		privateKey:   privateKey,
		reconnect:    make(chan struct{}, 1),
		contracts:    make(map[string]*Contracts),
	}

	// 初始化所有链
	var errgroup errgroup.Group
	for chain := range chainConfigs {
		chain := chain
		errgroup.Go(func() error {
			return client.initChain(chain)
		})
	}
	if err := errgroup.Wait(); err != nil {
		return nil, fmt.Errorf("failed to init chains: %w", err)
	}

	// 启动连接监控
	go client.monitorConnections()

	return client, nil
}

// initChain 初始化链
func (c *Client) initChain(chain string) error {
	// 创建 WebSocket 客户端
	wsClient, err := ethclient.Dial(c.chainConfigs[chain].RPCURL)
	if err != nil {
		return fmt.Errorf("failed to dial ws: %w", err)
	}
	chainID, err := wsClient.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain id: %w", err)
	}

	// 创建认证
	key, err := crypto.HexToECDSA(c.privateKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	c.logger.Info("chain liquiditor", zap.String("chain", chain), zap.String("liquiditor", crypto.PubkeyToAddress(key.PublicKey).Hex()))

	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return fmt.Errorf("failed to create auth: %w", err)
	}

	contracts, err := NewContracts(wsClient, c.chainConfigs[chain].ContractAddresses)
	if err != nil {
		return fmt.Errorf("failed to create contracts: %w", err)
	}

	c.Lock()
	c.clients[chain] = wsClient
	c.auths[chain] = auth
	c.contracts[chain] = contracts
	c.Unlock()

	return nil
}

// monitorConnections 监控连接
func (c *Client) monitorConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.checkConnections(); err != nil {
				c.logger.Error("failed to check connections", zap.Error(err))
			}
		case <-c.reconnect:
			if err := c.checkConnections(); err != nil {
				c.logger.Error("failed to check connections", zap.Error(err))
			}
		}
	}
}

// checkConnections 检查连接
func (c *Client) checkConnections() error {
	c.RLock()
	defer c.RUnlock()

	for chain, client := range c.clients {
		// 检查连接
		_, err := client.BlockNumber(context.Background())
		if err != nil {
			c.logger.Error("connection lost", zap.String("chain", chain), zap.Error(err))
			go c.reconnectChain(chain)
		}
	}

	return nil
}

// reconnectChain 重连链
func (c *Client) reconnectChain(chain string) {
	c.Lock()
	defer c.Unlock()

	// 关闭旧连接
	if client, ok := c.clients[chain]; ok {
		client.Close()
	}

	// 重新初始化
	if err := c.initChain(chain); err != nil {
		c.logger.Error("failed to reconnect", zap.String("chain", chain), zap.Error(err))
	}
}

// GetClient 获取客户端
func (c *Client) GetClient(chain string) (*ethclient.Client, error) {
	c.RLock()
	defer c.RUnlock()

	client, ok := c.clients[chain]
	if !ok {
		return nil, fmt.Errorf("chain %s not found", chain)
	}

	return client, nil
}

// GetAuth 获取认证
func (c *Client) GetAuth(chain string) (*bind.TransactOpts, error) {
	c.RLock()
	defer c.RUnlock()

	auth, ok := c.auths[chain]
	if !ok {
		return nil, fmt.Errorf("chain %s not found", chain)
	}

	return auth, nil
}

func (c *Client) GetContracts(chain string) (*Contracts, error) {
	c.RLock()
	defer c.RUnlock()

	contracts, ok := c.contracts[chain]
	if !ok {
		return nil, fmt.Errorf("chain %s not found", chain)
	}

	return contracts, nil
}
