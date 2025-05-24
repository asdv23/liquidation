import { ChainConfig } from '../interfaces/chain-config.interface';

export const chainsConfig: Record<string, ChainConfig> = {
    base: {
        name: 'Base',
        rpcUrl: process.env.BASE_RPC_URL || 'https://mainnet.base.org',
        chainId: 8453,
        contracts: {
            lendingPool: process.env.BASE_LENDING_POOL_ADDRESS || '',
            // 其他合约地址
        },
        blockTime: 2, // 秒
        confirmations: 1,
    },
    optimism: {
        name: 'Optimism',
        rpcUrl: process.env.OPTIMISM_RPC_URL || 'https://mainnet.optimism.io',
        chainId: 10,
        contracts: {
            lendingPool: process.env.OPTIMISM_LENDING_POOL_ADDRESS || '',
            // 其他合约地址
        },
        blockTime: 2,
        confirmations: 1,
    },
}; 