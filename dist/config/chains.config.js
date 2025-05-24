"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.chainsConfig = void 0;
exports.chainsConfig = {
    base: {
        name: 'Base',
        rpcUrl: process.env.BASE_RPC_URL || 'https://mainnet.base.org',
        chainId: 8453,
        contracts: {
            lendingPool: process.env.BASE_LENDING_POOL_ADDRESS || '',
        },
        blockTime: 2,
        confirmations: 1,
    },
    optimism: {
        name: 'Optimism',
        rpcUrl: process.env.OPTIMISM_RPC_URL || 'https://mainnet.optimism.io',
        chainId: 10,
        contracts: {
            lendingPool: process.env.OPTIMISM_LENDING_POOL_ADDRESS || '',
        },
        blockTime: 2,
        confirmations: 1,
    },
};
//# sourceMappingURL=chains.config.js.map