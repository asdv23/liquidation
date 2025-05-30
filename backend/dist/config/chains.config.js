"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.chainsConfig = void 0;
const common_1 = require("@nestjs/common");
const dotenv = require("dotenv");
const logger = new common_1.Logger('ChainsConfig');
dotenv.config();
function getChainsConfigFromEnv() {
    logger.log('Loading chain configurations from environment variables...');
    logger.log('Available environment variables:', Object.keys(process.env).filter(key => key.includes('_RPC_URL')));
    const chains = {};
    for (const key of Object.keys(process.env)) {
        const match = key.match(/^([A-Z0-9_]+)_RPC_URL$/);
        if (match) {
            const chainKey = match[1].toLowerCase();
            const rpcUrl = process.env[`${chainKey.toUpperCase()}_RPC_URL`];
            const aavev3Pool = process.env[`${chainKey.toUpperCase()}_AAVE_V3_POOL`];
            const flashLoanLiquidation = process.env[`${chainKey.toUpperCase()}_FLASH_LOAN_LIQUIDATION`];
            const blockTime = process.env[`${chainKey.toUpperCase()}_BLOCK_TIME`];
            if (!rpcUrl || !aavev3Pool || !flashLoanLiquidation || !blockTime) {
                logger.warn(`Missing configuration for chain ${chainKey}: RPC_URL=${rpcUrl}, AAVE_V3_POOL=${aavev3Pool}, FLASH_LOAN_LIQUIDATION=${flashLoanLiquidation}, BLOCK_TIME=${blockTime}`);
                continue;
            }
            const blockTimeMs = parseInt(blockTime, 10);
            if (isNaN(blockTimeMs)) {
                logger.warn(`Invalid BLOCK_TIME for chain ${chainKey}: ${blockTime}`);
                continue;
            }
            chains[chainKey] = {
                name: chainKey,
                rpcUrl: rpcUrl,
                contracts: {
                    aavev3Pool,
                    flashLoanLiquidation,
                },
                blockTime: blockTimeMs,
                minWaitTime: Math.floor(blockTimeMs / 2),
            };
            logger.log(`Loaded chain config for ${chainKey}: RPC=${rpcUrl}, AAVE_V3_POOL=${aavev3Pool}, FlashLoanLiquidation=${flashLoanLiquidation}, BlockTime=${blockTimeMs}ms, MinWaitTime=${Math.floor(blockTimeMs / 2)}ms`);
        }
    }
    if (Object.keys(chains).length === 0) {
        logger.error('No chain configurations found in environment variables!');
        ``;
    }
    else {
        logger.log(`Successfully loaded ${Object.keys(chains).length} chain configurations: ${JSON.stringify(chains, null, 2)}`);
    }
    return chains;
}
exports.chainsConfig = getChainsConfigFromEnv();
//# sourceMappingURL=chains.config.js.map