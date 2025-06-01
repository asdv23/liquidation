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
            const aavev3Pool = process.env[`${chainKey.toUpperCase()}_AAVE_V3_POOL`];
            const flashLoanLiquidation = process.env[`${chainKey.toUpperCase()}_FLASH_LOAN_LIQUIDATION`];
            const usdc = process.env[`${chainKey.toUpperCase()}_USDC`];
            const blockTime = process.env[`${chainKey.toUpperCase()}_BLOCK_TIME`];
            const nativePrice = process.env[`${chainKey.toUpperCase()}_NATIVE_PRICE`] || 3000;

            if (!rpcUrl || !aavev3Pool || !flashLoanLiquidation || !blockTime || !usdc) {
                logger.warn(`Missing configuration for chain ${chainKey}: RPC_URL=${rpcUrl}, AAVE_V3_POOL=${aavev3Pool}, FLASH_LOAN_LIQUIDATION=${flashLoanLiquidation}, BLOCK_TIME=${blockTime}`);
                continue;
            }

            const blockTimeMs = parseInt(blockTime, 10);
            if (isNaN(blockTimeMs)) {
                logger.warn(`Invalid BLOCK_TIME for chain ${chainKey}: ${blockTime}`);
                continue;
            }

            chains[chainKey] = {
                chainId: 0,
                name: chainKey,
                rpcUrl: rpcUrl,
                contracts: {
                    aavev3Pool,
                    flashLoanLiquidation,
                    usdc,
                },
                blockTime: blockTimeMs,
                minWaitTime: Math.floor(blockTimeMs / 2),
                nativePrice: Number(nativePrice),
                minDebtUSD: 2,
            };
            logger.log(`Loaded chain config for ${chainKey}: RPC=${rpcUrl}, AAVE_V3_POOL=${aavev3Pool}, FlashLoanLiquidation=${flashLoanLiquidation}, BlockTime=${blockTimeMs}ms, MinWaitTime=${Math.floor(blockTimeMs / 2)}ms, NativePrice=${nativePrice}`);
        }
    }

    if (Object.keys(chains).length === 0) {
        logger.error('No chain configurations found in environment variables!');
        ``
    } else {
        logger.log(`Successfully loaded ${Object.keys(chains).length} chain configurations: ${JSON.stringify(chains, null, 2)}`);
    }

    return chains;
}

export const chainsConfig = getChainsConfigFromEnv(); 