package blockchain

import (
	"context"
	"fmt"
	"liquidation-bot/config"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
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
	ChainID   *big.Int

	reconnectCh  chan struct{}
	reconnectFns []ReconnectFn
	privateKey   string

	client    *ethclient.Client
	auth      func() (*bind.TransactOpts, error)
	contracts *Contracts
	baseFee   *big.Int
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

type ReconnectFn func()

func (c *Chain) Register(reconnectFn ReconnectFn) {
	c.Lock()
	defer c.Unlock()

	c.reconnectFns = append(c.reconnectFns, reconnectFn)
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
	chainID, err := wsClient.ChainID(c.Ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain id: %w", err)
	}
	header, err := wsClient.HeaderByNumber(c.Ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get header: %w", err)
	}
	c.ChainID = chainID
	c.client = wsClient
	c.baseFee = header.BaseFee

	// 创建认证
	key, err := crypto.HexToECDSA(strings.TrimPrefix(c.privateKey, "0x"))
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}
	c.Logger.Info("liquidator", zap.String("address", crypto.PubkeyToAddress(key.PublicKey).Hex()))

	c.auth = func() (*bind.TransactOpts, error) {
		auth, err := bind.NewKeyedTransactorWithChainID(key, c.ChainID)
		if err != nil {
			return nil, fmt.Errorf("failed to create auth: %w", err)
		}
		return auth, nil
	}

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
			if c.baseFee.Cmp(header.BaseFee) != 0 {
				c.baseFee = header.BaseFee
				c.Logger.Info("New base fee", zap.String("baseFee", header.BaseFee.String()))
			}
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
			err := retry.Do(func() error {
				if err := c.connect(); err != nil {
					c.Logger.Error("Reconnect attempt failed", zap.Error(err))
					return fmt.Errorf("failed to reconnect: %w", err)
				}
				c.Logger.Info("Reconnected successfully")
				for _, fn := range c.reconnectFns {
					fn := fn
					c.Logger.Info("Reconnecting", zap.String("fn", fmt.Sprintf("%T", fn)))
					go fn()
				}
				return nil
			}, retry.Attempts(5), retry.Delay(time.Second), retry.MaxDelay(time.Minute))
			if err != nil {
				c.Logger.Fatal("Failed to reconnect", zap.Error(err))
			}
		}
	}
}

func (c *Chain) GetClient() *ethclient.Client {
	c.RLock()
	defer c.RUnlock()
	return c.client
}

func (c *Chain) GetAuth() (*bind.TransactOpts, error) {
	c.RLock()
	defer c.RUnlock()
	return c.auth()
}

func (c *Chain) GetContracts() *Contracts {
	c.RLock()
	defer c.RUnlock()
	return c.contracts
}

func (c *Chain) GetBaseFee() *big.Int {
	c.RLock()
	defer c.RUnlock()
	return c.baseFee
}
