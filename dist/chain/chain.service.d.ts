import { OnModuleInit } from '@nestjs/common';
import { ethers } from 'ethers';
import { ChainConfig } from '../interfaces/chain-config.interface';
export declare class ChainService implements OnModuleInit {
    private readonly logger;
    private providers;
    onModuleInit(): void;
    private initializeProviders;
    getProvider(chainName: string): ethers.JsonRpcProvider;
    getChainConfig(chainName: string): ChainConfig;
    getActiveChains(): string[];
}
