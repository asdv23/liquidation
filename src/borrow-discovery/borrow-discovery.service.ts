import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ethers } from 'ethers';
import { ChainService } from '../chain/chain.service';
import * as fs from 'fs';
import * as path from 'path';

@Injectable()
export class BorrowDiscoveryService implements OnModuleInit {
    private readonly logger = new Logger(BorrowDiscoveryService.name);
    private readonly unsafeLoans: Map<string, any[]> = new Map();
    private readonly aaveV3PoolABI = JSON.parse(fs.readFileSync(path.join(process.cwd(), 'abis/AAVE_V3_POOL.json'), 'utf8'));

    constructor(private readonly chainService: ChainService) { }

    async onModuleInit() {
        await this.startListening();
    }

    private async startListening() {
        const chains = this.chainService.getActiveChains();
        for (const chain of chains) {
            const provider = this.chainService.getProvider(chain);
            const config = this.chainService.getChainConfig(chain);
            const contractAddress = config.contracts.lendingPool;
            const contract = new ethers.Contract(contractAddress, this.aaveV3PoolABI, provider);

            // 监听 Borrow 事件
            contract.on('Borrow', async (reserve, user, onBehalfOf, amount, interestRateMode, borrowRate, referral, event) => {
                this.logger.log(`[${chain}] New borrow detected: user=${onBehalfOf}, amount=${ethers.formatEther(amount)}`);
                await this.checkHealthFactor(chain, onBehalfOf, contract);
            });

            this.logger.log(`Started listening for Borrow events on ${chain} at ${contractAddress}`);
        }
    }

    private async checkHealthFactor(chain: string, userAddress: string, contract: ethers.Contract) {
        try {
            const userData = await contract.getUserAccountData(userAddress);
            const healthFactor = ethers.formatUnits(userData.healthFactor, 18);
            this.logger.log(`[${chain}] User ${userAddress} health factor: ${healthFactor}`);

            if (Number(healthFactor) < 1) {
                this.logger.warn(`[${chain}] Unsafe loan detected for user ${userAddress} with health factor ${healthFactor}`);
                if (!this.unsafeLoans.has(chain)) {
                    this.unsafeLoans.set(chain, []);
                }
                this.unsafeLoans.get(chain).push({ user: userAddress, healthFactor, timestamp: new Date().toISOString() });
            }
        } catch (error) {
            this.logger.error(`[${chain}] Error checking health factor for user ${userAddress}: ${error.message}`);
        }
    }

    getUnsafeLoans(chain: string): any[] {
        return this.unsafeLoans.get(chain) || [];
    }
} 