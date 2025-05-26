export interface ChainConfig {
    name: string;
    rpcUrl: string;
    chainId: number;
    contracts: {
        lendingPool: string;
        [key: string]: string;
    };
    blockTime: number;
    confirmations: number;
}
