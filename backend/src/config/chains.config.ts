import { ChainConfig } from '../interfaces/chain-config.interface';
import { Logger } from '@nestjs/common';
import * as dotenv from 'dotenv';

const logger = new Logger('ChainsConfig');

// 加载 .env 文件
dotenv.config();

function getChainsConfigFromEnv(): Record<string, ChainConfig> {
    logger.log('Loading chain configurations from environment variables...');
    logger.log('Available environment variables:', Object.keys(process.env).filter(key => key.includes('_RPC_URL')));

    const chains: Record<string, ChainConfig> = {};
    for (const key of Object.keys(process.env)) {
        const match = key.match(/^([A-Z0-9_]+)_RPC_URL$/);
        if (match) {
            const chainKey = match[1].toLowerCase();
            const rpcUrl = process.env[`${chainKey.toUpperCase()}_RPC_URL`];
            const contractAddress = process.env[`${chainKey.toUpperCase()}_AAVE_V3_POOL`];

            if (!rpcUrl || !contractAddress) {
                logger.warn(`Missing configuration for chain ${chainKey}: RPC_URL=${rpcUrl}, CONTRACT=${contractAddress}`);
                continue;
            }

            chains[chainKey] = {
                name: chainKey,
                rpcUrl: rpcUrl,
                chainId: 0,
                contracts: {
                    lendingPool: contractAddress,
                },
                blockTime: 2,
                confirmations: 1,
            };
            logger.log(`Loaded chain config for ${chainKey}: RPC=${rpcUrl}, Contract=${contractAddress}`);
        }
    }

    if (Object.keys(chains).length === 0) {
        logger.error('No chain configurations found in environment variables!');
    } else {
        logger.log(`Successfully loaded ${Object.keys(chains).length} chain configurations: ${Object.keys(chains).join(', ')}`);
    }

    return chains;
}

export const chainsConfig = getChainsConfigFromEnv(); 