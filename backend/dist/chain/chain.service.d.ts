import { OnModuleInit } from '@nestjs/common';
import { ethers } from 'ethers';
import { ChainConfig } from '../interfaces/chain-config.interface';
export declare class ChainService implements OnModuleInit {
    private readonly logger;
    private providers;
    private initializationPromise;
    onModuleInit(): Promise<void>;
    private initializeProviders;
    private initializeProvider;
    getProvider(chainName: string): Promise<ethers.WebSocketProvider>;
    getChainConfig(chainName: string): ChainConfig;
    getActiveChains(): string[];
}
