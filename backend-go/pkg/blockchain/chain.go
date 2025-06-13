package blockchain

import (
	"context"
	"fmt"
	"liquidation-bot/config"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type Chain struct {
	sync.RWMutex

	Ctx       context.Context
	Logger    *zap.Logger
	Cfg       config.ChainConfig
	ChainName string

	reconnectCh chan struct{}
	privateKey  string

	client    *ethclient.Client
	auth      *bind.TransactOpts
	contracts *Contracts
}

func NewChain(ctx context.Context, logger *zap.Logger, chainName string, privateKey string, cfg config.ChainConfig) (*Chain, error) {
	c := &Chain{
		Ctx:         ctx,
		Logger:      logger.Named(chainName),
		ChainName:   chainName,
		Cfg:         cfg,
		reconnectCh: make(chan struct{}, 1),
		privateKey:  privateKey,
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

	go c.reconnect()
	return c, nil
}

func (c *Chain) connect() error {
	c.Lock()
	defer c.Unlock()

	if !strings.HasPrefix(c.Cfg.RPCURL, "ws") {
		return fmt.Errorf("chain %s rpc url is not a websocket url", c.ChainName)
	}

	// 创建 WebSocket 客户端
	wsClient, err := ethclient.Dial(c.Cfg.RPCURL)
	if err != nil {
		return fmt.Errorf("failed to dial ws: %w", err)
	}
	chainID, err := wsClient.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get chain id: %w", err)
	}
	c.client = wsClient

	// 创建认证
	key, err := crypto.HexToECDSA(c.privateKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return fmt.Errorf("failed to create auth: %w", err)
	}
	c.auth = auth

	contracts, err := NewContracts(wsClient, c.Cfg.ContractAddresses)
	if err != nil {
		return fmt.Errorf("failed to create contracts: %w", err)
	}
	c.contracts = contracts

	// 连接成功后订阅事件
	go c.subscribe()
	return nil
}

func (c *Chain) subscribe() {
	c.RLock()
	client := c.client
	c.RUnlock()

	defer client.Close()

	headers := make(chan *types.Header, 100)
	sub, err := client.SubscribeNewHead(c.Ctx, headers)
	if err != nil {
		c.Logger.Error("Subscription failed", zap.Error(err))
		c.reconnectCh <- struct{}{} // 触发重连
		return
	}
	defer sub.Unsubscribe()

	for {
		select {
		case <-c.Ctx.Done():
			return
		case err := <-sub.Err():
			c.Logger.Error("Subscription error", zap.Error(err))
			c.reconnectCh <- struct{}{} // 触发重连
			return
		case header := <-headers:
			c.Logger.Debug("New block", zap.Uint64("blockNumber", header.Number.Uint64()))
		}
	}
}

func (c *Chain) reconnect() {
	for {
		select {
		case <-c.Ctx.Done():
			return
		case <-c.reconnectCh:
			c.Logger.Info("WebSocket disconnected, attempting to reconnect...")
			maxAttempts := 5
			for attempt := 1; attempt <= maxAttempts; attempt++ {
				if err := c.connect(); err != nil {
					c.Logger.Error("Reconnect attempt failed", zap.Int("attempt", attempt), zap.Error(err))
					time.Sleep(time.Duration(attempt) * time.Second) // 简单指数退避
					continue
				}
				c.Logger.Info("Reconnected successfully")
				break
			}
			c.Logger.Error("Failed to reconnect after max attempts, retrying later...", zap.Int("maxAttempts", maxAttempts))
			time.Sleep(time.Minute) // 等待更长时间后重试
		}
	}
}

func (c *Chain) GetClient() *ethclient.Client {
	c.RLock()
	defer c.RUnlock()
	return c.client
}

func (c *Chain) GetAuth() *bind.TransactOpts {
	c.RLock()
	defer c.RUnlock()
	return c.auth
}

func (c *Chain) GetContracts() *Contracts {
	c.RLock()
	defer c.RUnlock()
	return c.contracts
}
