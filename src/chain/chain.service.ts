import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ethers } from 'ethers';
import { chainsConfig } from '../config/chains.config';
import { ChainConfig } from '../interfaces/chain-config.interface';
import WebSocket from 'ws';

@Injectable()
export class ChainService implements OnModuleInit {
    private readonly logger = new Logger(ChainService.name);
    private providers: Map<string, ethers.WebSocketProvider> = new Map();
    private initializationPromise: Promise<void> | null = null;

    async onModuleInit() {
        this.initializationPromise = this.initializeProviders();
        await this.initializationPromise;
    }

    private async initializeProviders() {
        this.logger.log('Initializing providers...');
        const initPromises = Object.entries(chainsConfig).map(([chainName, config]) =>
            this.initializeProvider(chainName, config)
        );
        await Promise.all(initPromises);
        this.logger.log('Providers initialized.');
    }

    private async initializeProvider(chainName: string, config: ChainConfig) {
        try {
            // 确保 URL 是 wss:// 格式
            const wsUrl = config.rpcUrl.replace('https://', 'wss://');
            const provider = new ethers.WebSocketProvider(wsUrl);

            // 等待网络检测
            await provider.getNetwork();
            this.logger.log(`Network detected for ${chainName}`);

            // 添加错误处理
            const ws = provider.websocket as WebSocket;
            ws.on('error', (error: Error) => {
                this.logger.error(`WebSocket error for ${chainName}: ${error.message}`);
            });

            // 添加重连逻辑
            ws.on('close', () => {
                this.logger.warn(`WebSocket connection closed for ${chainName}, attempting to reconnect...`);
                setTimeout(() => {
                    this.initializeProvider(chainName, config);
                }, 5000); // 5秒后重试
            });

            this.providers.set(chainName, provider);
            this.logger.log(`Initialized provider for ${chainName} at ${wsUrl}`);
        } catch (error) {
            this.logger.error(`Failed to initialize provider for ${chainName}: ${error.message}`);
            // 如果初始化失败，5秒后重试
            setTimeout(() => {
                this.initializeProvider(chainName, config);
            }, 5000);
        }
    }

    async getProvider(chainName: string): Promise<ethers.WebSocketProvider> {
        // 如果初始化还在进行中，等待它完成
        if (this.initializationPromise) {
            await this.initializationPromise;
        }

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