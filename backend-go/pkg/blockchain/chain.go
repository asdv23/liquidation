package blockchain

import (
	"context"
	"fmt"
	"liquidation-bot/config"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type ReconnectFn func()

type Chain struct {
	sync.RWMutex

	Ctx       context.Context
	Logger    *zap.Logger
	Cfg       config.ChainConfig
	ChainName string
	ChainID   *big.Int

	privateKey string

	client    *ethclient.Client
	auth      func() (*bind.TransactOpts, error)
	contracts *Contracts
}

func NewChain(ctx context.Context, logger *zap.Logger, chainName string, privateKey string, cfg config.ChainConfig) (*Chain, error) {
	c := &Chain{
		Ctx:        ctx,
		Logger:     logger.Named(chainName),
		ChainName:  chainName,
		Cfg:        cfg,
		privateKey: privateKey,
	}

	if err := c.connect(); err != nil {
		return nil, err
	}

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
	chainID, err := wsClient.ChainID(c.Ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain id: %w", err)
	}
	c.client = wsClient
	c.ChainID = chainID

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
	c.Logger.Info("contract addresses", zap.Any("contracts", c.contracts.Addresses))

	return nil
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
