import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ethers } from 'ethers';
import { chainsConfig } from '../config/chains.config';
import { ChainConfig } from '../interfaces/chain-config.interface';

@Injectable()
export class ChainService implements OnModuleInit {
    private readonly logger = new Logger(ChainService.name);
    private providers: Map<string, ethers.JsonRpcProvider> = new Map();

    onModuleInit() {
        this.initializeProviders();
    }

    private initializeProviders() {
        Object.entries(chainsConfig).forEach(([chainName, config]) => {
            try {
                const provider = new ethers.JsonRpcProvider(config.rpcUrl);
                this.providers.set(chainName, provider);
                this.logger.log(`Initialized provider for ${chainName}`);
            } catch (error) {
                this.logger.error(`Failed to initialize provider for ${chainName}: ${error.message}`);
            }
        });
    }

    getProvider(chainName: string): ethers.JsonRpcProvider {
        const provider = this.providers.get(chainName);
        if (!provider) {
            throw new Error(`Provider for chain ${chainName} not found`);
        }
        return provider;
    }

    getChainConfig(chainName: string): ChainConfig {
        const config = chainsConfig[chainName];
        if (!config) {
            throw new Error(`Configuration for chain ${chainName} not found`);
        }
        return config;
    }

    getActiveChains(): string[] {
        return Object.keys(chainsConfig);
    }
} 