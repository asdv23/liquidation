export interface ChainConfig {
    chainId: number;
    name: string;
    rpcUrl: string;
    contracts: {
        [key: string]: string;
    };
    blockTime: number;
    minWaitTime: number;
    nativePrice: number;
    minDebtUSD: number;
} 