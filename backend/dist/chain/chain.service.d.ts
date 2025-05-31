import { OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import { ethers } from 'ethers';
import { ChainConfig } from '../interfaces/chain-config.interface';
import { ConfigService } from '@nestjs/config';
export declare class ChainService implements OnModuleInit, OnModuleDestroy {
    private readonly configService;
    private readonly logger;
    private providers;
    private initializationPromise;
    private readonly PRIVATE_KEY;
    constructor(configService: ConfigService);
    onModuleInit(): Promise<void>;
    onModuleDestroy(): Promise<void>;
    private initializeProviders;
    private initializeProvider;
    getProvider(chainName: string): Promise<ethers.WebSocketProvider>;
    getSigner(chainName: string): Promise<ethers.Signer>;
    getChainConfig(chainName: string): ChainConfig;
    getActiveChains(): string[];
}
