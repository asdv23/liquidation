export interface ChainConfig {
    name: string;
    rpcUrl: string;
    contracts: {
        [key: string]: string;
    };
    blockTime: number;
} 