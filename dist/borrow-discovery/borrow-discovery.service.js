"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
var BorrowDiscoveryService_1;
Object.defineProperty(exports, "__esModule", { value: true });
exports.BorrowDiscoveryService = void 0;
const common_1 = require("@nestjs/common");
const ethers_1 = require("ethers");
const chain_service_1 = require("../chain/chain.service");
const fs = require("fs");
const path = require("path");
let BorrowDiscoveryService = BorrowDiscoveryService_1 = class BorrowDiscoveryService {
    constructor(chainService) {
        this.chainService = chainService;
        this.logger = new common_1.Logger(BorrowDiscoveryService_1.name);
        this.unsafeLoans = new Map();
        this.activeLoans = new Map();
        this.liquidationTimes = new Map();
        this.aaveV3PoolABI = JSON.parse(fs.readFileSync(path.join(process.cwd(), 'abis/AAVE_V3_POOL.json'), 'utf8'));
        this.pollingInterval = parseInt(process.env.POLLING_INTERVAL || '300000', 10);
    }
    async onModuleInit() {
        await this.startListening();
        this.startHealthFactorPolling();
    }
    async startListening() {
        const chains = this.chainService.getActiveChains();
        for (const chain of chains) {
            const provider = this.chainService.getProvider(chain);
            const config = this.chainService.getChainConfig(chain);
            const contractAddress = config.contracts.lendingPool;
            const contract = new ethers_1.ethers.Contract(contractAddress, this.aaveV3PoolABI, provider);
            contract.on('Borrow', async (reserve, user, onBehalfOf, amount, interestRateMode, borrowRate, referral, event) => {
                this.logger.log(`[${chain}] New borrow detected: user=${onBehalfOf}, amount=${ethers_1.ethers.formatEther(amount)}`);
                if (!this.activeLoans.has(chain)) {
                    this.activeLoans.set(chain, new Set());
                }
                const activeLoansSet = this.activeLoans.get(chain);
                if (activeLoansSet) {
                    activeLoansSet.add(onBehalfOf);
                }
                await this.checkHealthFactor(chain, onBehalfOf, contract);
            });
            contract.on('Repay', async (reserve, user, repayer, amount, useATokens, event) => {
                this.logger.log(`[${chain}] Repay detected: user=${user}, amount=${ethers_1.ethers.formatEther(amount)}`);
                const activeLoansSet = this.activeLoans.get(chain);
                if (activeLoansSet) {
                    activeLoansSet.delete(user);
                }
            });
            contract.on('LiquidationCall', async (collateralAsset, debtAsset, user, debtToCover, liquidatedCollateralAmount, liquidator, receiveAToken, event) => {
                this.logger.log(`[${chain}] Liquidation detected: user=${user}, liquidator=${liquidator}`);
                const activeLoansSet = this.activeLoans.get(chain);
                if (activeLoansSet) {
                    activeLoansSet.delete(user);
                }
                this.recordLiquidationTime(chain, user);
            });
            this.logger.log(`Started listening for events on ${chain} at ${contractAddress}`);
        }
    }
    async checkHealthFactor(chain, userAddress, contract) {
        try {
            const userData = await contract.getUserAccountData(userAddress);
            const healthFactor = ethers_1.ethers.formatUnits(userData.healthFactor, 18);
            this.logger.log(`[${chain}] User ${userAddress} health factor: ${healthFactor}`);
            if (Number(healthFactor) < 1) {
                this.logger.warn(`[${chain}] Unsafe loan detected for user ${userAddress} with health factor ${healthFactor}`);
                if (!this.unsafeLoans.has(chain)) {
                    this.unsafeLoans.set(chain, []);
                }
                const unsafeLoansList = this.unsafeLoans.get(chain);
                if (unsafeLoansList) {
                    unsafeLoansList.push({ user: userAddress, healthFactor, timestamp: new Date().toISOString() });
                }
            }
        }
        catch (error) {
            this.logger.error(`[${chain}] Error checking health factor for user ${userAddress}: ${error.message}`);
        }
    }
    startHealthFactorPolling() {
        setInterval(async () => {
            const chains = this.chainService.getActiveChains();
            for (const chain of chains) {
                const provider = this.chainService.getProvider(chain);
                const config = this.chainService.getChainConfig(chain);
                const contractAddress = config.contracts.lendingPool;
                const contract = new ethers_1.ethers.Contract(contractAddress, this.aaveV3PoolABI, provider);
                const activeLoansSet = this.activeLoans.get(chain);
                if (activeLoansSet) {
                    for (const user of activeLoansSet) {
                        await this.checkHealthFactor(chain, user, contract);
                    }
                }
            }
        }, this.pollingInterval);
    }
    recordLiquidationTime(chain, user) {
        if (!this.liquidationTimes.has(chain)) {
            this.liquidationTimes.set(chain, []);
        }
        const liquidationTimesList = this.liquidationTimes.get(chain);
        if (liquidationTimesList) {
            liquidationTimesList.push({ user, timestamp: new Date().toISOString() });
        }
    }
    getUnsafeLoans(chain) {
        return this.unsafeLoans.get(chain) || [];
    }
    getLiquidationTimes(chain) {
        return this.liquidationTimes.get(chain) || [];
    }
};
exports.BorrowDiscoveryService = BorrowDiscoveryService;
exports.BorrowDiscoveryService = BorrowDiscoveryService = BorrowDiscoveryService_1 = __decorate([
    (0, common_1.Injectable)(),
    __metadata("design:paramtypes", [chain_service_1.ChainService])
], BorrowDiscoveryService);
//# sourceMappingURL=borrow-discovery.service.js.map