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
const chain_service_1 = require("../chain/chain.service");
const ethers_1 = require("ethers");
const config_1 = require("@nestjs/config");
const database_service_1 = require("../database/database.service");
const fs = require("fs");
const path = require("path");
let BorrowDiscoveryService = BorrowDiscoveryService_1 = class BorrowDiscoveryService {
    constructor(chainService, configService, databaseService) {
        this.chainService = chainService;
        this.configService = configService;
        this.databaseService = databaseService;
        this.logger = new common_1.Logger(BorrowDiscoveryService_1.name);
        this.activeLoans = new Map();
        this.liquidationTimes = new Map();
        this.tokenCache = new Map();
        this.LIQUIDATION_THRESHOLD = 1.05;
        this.CRITICAL_THRESHOLD = 1.1;
        this.HEALTH_FACTOR_THRESHOLD = 1.2;
        this.MIN_WAIT_TIME = this.configService.get('MIN_CHECK_INTERVAL', 200);
        this.MAX_WAIT_TIME = this.configService.get('MAX_CHECK_INTERVAL', 30 * 60 * 1000);
    }
    async onModuleInit() {
        this.logger.log('BorrowDiscoveryService initializing...');
        await this.loadTokenCache();
        await this.loadActiveLoans();
        await this.startListening();
        this.startHealthFactorChecker();
        this.startHeartbeat();
    }
    async loadTokenCache() {
        try {
            const tokens = await this.databaseService.getAllTokens();
            for (const token of tokens) {
                if (!this.tokenCache.has(token.chainName)) {
                    this.tokenCache.set(token.chainName, new Map());
                }
                const chainTokens = this.tokenCache.get(token.chainName);
                chainTokens.set(token.address.toLowerCase(), {
                    symbol: token.symbol,
                    decimals: token.decimals,
                });
            }
            this.logger.log(`Loaded ${tokens.length} tokens into cache`);
        }
        catch (error) {
            this.logger.error(`Error loading token cache: ${error.message}`);
        }
    }
    async getTokenInfo(chainName, address, provider) {
        const normalizedAddress = address.toLowerCase();
        const chainTokens = this.tokenCache.get(chainName);
        if (chainTokens === null || chainTokens === void 0 ? void 0 : chainTokens.has(normalizedAddress)) {
            return chainTokens.get(normalizedAddress);
        }
        const dbToken = await this.databaseService.getToken(chainName, normalizedAddress);
        if (dbToken) {
            if (!this.tokenCache.has(chainName)) {
                this.tokenCache.set(chainName, new Map());
            }
            const tokenInfo = {
                symbol: dbToken.symbol,
                decimals: dbToken.decimals,
            };
            this.tokenCache.get(chainName).set(normalizedAddress, tokenInfo);
            return tokenInfo;
        }
        try {
            const erc20Abi = [
                'function symbol() view returns (string)',
                'function decimals() view returns (uint8)',
            ];
            const contract = new ethers_1.ethers.Contract(normalizedAddress, erc20Abi, provider);
            const [symbol, decimals] = await Promise.all([
                contract.symbol(),
                contract.decimals(),
            ]);
            await this.databaseService.saveToken(chainName, normalizedAddress, symbol, Number(decimals));
            if (!this.tokenCache.has(chainName)) {
                this.tokenCache.set(chainName, new Map());
            }
            const tokenInfo = { symbol, decimals: Number(decimals) };
            this.tokenCache.get(chainName).set(normalizedAddress, tokenInfo);
            return tokenInfo;
        }
        catch (error) {
            this.logger.error(`Error getting token info for ${normalizedAddress} on ${chainName}: ${error.message}`);
            return { symbol: 'UNKNOWN', decimals: 18 };
        }
    }
    formatAmount(amount, decimals) {
        return Number(ethers_1.ethers.formatUnits(amount, decimals)).toFixed(6);
    }
    async loadActiveLoans() {
        try {
            const chains = this.chainService.getActiveChains();
            for (const chainName of chains) {
                const activeLoans = await this.databaseService.getActiveLoans(chainName);
                this.logger.log(`[${chainName}] Found ${activeLoans.length} active loans in database`);
                if (activeLoans.length > 0) {
                    this.logger.log(`[${chainName}] Checking health factors for ${activeLoans.length} active loans...`);
                    const provider = await this.chainService.getProvider(chainName);
                    const config = this.chainService.getChainConfig(chainName);
                    const abiPath = path.join(process.cwd(), 'abi', `${chainName.toUpperCase()}_AAVE_V3_POOL_ABI.json`);
                    const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
                    const contract = new ethers_1.ethers.Contract(config.contracts.lendingPool, abi, provider);
                    for (const loan of activeLoans) {
                        this.checkHealthFactor(chainName, loan.user, contract);
                    }
                }
            }
        }
        catch (error) {
            this.logger.error(`Error loading active loans: ${error.message}`);
        }
    }
    async startListening() {
        const chains = this.chainService.getActiveChains();
        this.logger.log(`Starting to listen on chains: ${chains.join(', ')}`);
        for (const chainName of chains) {
            try {
                const provider = await this.chainService.getProvider(chainName);
                const config = this.chainService.getChainConfig(chainName);
                const abiPath = path.join(process.cwd(), 'abi', `${chainName.toUpperCase()}_AAVE_V3_POOL_ABI.json`);
                const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
                const currentBlock = await provider.getBlockNumber();
                this.logger.log(`[${chainName}] Current block number: ${currentBlock}`);
                const code = await provider.getCode(config.contracts.lendingPool);
                if (code === '0x') {
                    this.logger.error(`[${chainName}] No contract code found at address ${config.contracts.lendingPool}`);
                    continue;
                }
                this.logger.log(`[${chainName}] Contract code found at ${config.contracts.lendingPool}`);
                const contract = new ethers_1.ethers.Contract(config.contracts.lendingPool, abi, provider);
                try {
                    const wethAddress = chainName === 'base'
                        ? '0x4200000000000000000000000000000000000006'
                        : '0x4200000000000000000000000000000000000006';
                    const reserveData = await contract.getReserveData(wethAddress);
                    this.logger.log(`[${chainName}] Successfully connected to Aave V3 Pool at ${config.contracts.lendingPool}`);
                    this.logger.log(`[${chainName}] WETH Reserve Data:`);
                    this.logger.log(`- Current Liquidity Rate: ${ethers_1.ethers.formatUnits(reserveData.currentLiquidityRate, 27)}`);
                    this.logger.log(`- Current Variable Borrow Rate: ${ethers_1.ethers.formatUnits(reserveData.currentVariableBorrowRate, 27)}`);
                    this.logger.log(`- Current Stable Borrow Rate: ${ethers_1.ethers.formatUnits(reserveData.currentStableBorrowRate, 27)}`);
                }
                catch (error) {
                    this.logger.error(`[${chainName}] Failed to verify contract connection: ${error.message}`);
                    continue;
                }
                this.logger.log(`[${chainName}] Setting up Borrow event listener...`);
                try {
                    contract.on('Borrow', async (reserve, user, onBehalfOf, amount, interestRateMode, borrowRate, referralCode, event) => {
                        var _a;
                        try {
                            const tokenInfo = await this.getTokenInfo(chainName, reserve, provider);
                            this.logger.log(`[${chainName}] Borrow event detected:`);
                            this.logger.log(`- Reserve: ${reserve} (${tokenInfo.symbol})`);
                            this.logger.log(`- User: ${user}`);
                            this.logger.log(`- OnBehalfOf: ${onBehalfOf}`);
                            this.logger.log(`- Amount: ${this.formatAmount(amount, tokenInfo.decimals)} ${tokenInfo.symbol}`);
                            this.logger.log(`- Interest Rate Mode: ${interestRateMode}`);
                            this.logger.log(`- Borrow Rate: ${ethers_1.ethers.formatUnits(borrowRate, 27)}`);
                            this.logger.log(`- Referral Code: ${referralCode}`);
                            this.logger.log(`- Transaction Hash: ${(event === null || event === void 0 ? void 0 : event.transactionHash) || ((_a = event === null || event === void 0 ? void 0 : event.log) === null || _a === void 0 ? void 0 : _a.transactionHash)}`);
                            if (!this.activeLoans.has(chainName)) {
                                this.activeLoans.set(chainName, new Set());
                            }
                            const activeLoansSet = this.activeLoans.get(chainName);
                            if (activeLoansSet) {
                                activeLoansSet.add(onBehalfOf);
                            }
                            await this.databaseService.createOrUpdateLoan(chainName, onBehalfOf, 0);
                            this.logger.log(`[${chainName}] Created/Updated loan record for user ${onBehalfOf}`);
                        }
                        catch (error) {
                            this.logger.error(`[${chainName}] Error processing Borrow event: ${error.message}`);
                        }
                    });
                    this.logger.log(`[${chainName}] Borrow event listener setup completed`);
                }
                catch (error) {
                    this.logger.error(`[${chainName}] Failed to set up Borrow event listener: ${error.message}`);
                }
                contract.on('Repay', async (reserve, user, repayer, amount, useATokens, event) => {
                    var _a;
                    try {
                        const tokenInfo = await this.getTokenInfo(chainName, reserve, provider);
                        this.logger.log(`[${chainName}] Repay event detected:`);
                        this.logger.log(`- Reserve: ${reserve} (${tokenInfo.symbol})`);
                        this.logger.log(`- User: ${user}`);
                        this.logger.log(`- Repayer: ${repayer}`);
                        this.logger.log(`- Amount: ${this.formatAmount(amount, tokenInfo.decimals)} ${tokenInfo.symbol}`);
                        this.logger.log(`- Use ATokens: ${useATokens}`);
                        this.logger.log(`- Transaction Hash: ${(event === null || event === void 0 ? void 0 : event.transactionHash) || ((_a = event === null || event === void 0 ? void 0 : event.log) === null || _a === void 0 ? void 0 : _a.transactionHash)}`);
                        const activeLoans = await this.databaseService.getActiveLoans(chainName);
                        const loanExists = activeLoans.some(loan => loan.user.toLowerCase() === user.toLowerCase());
                        if (!loanExists) {
                            this.logger.warn(`[${chainName}] Received Repay event for non-existent loan: user=${user}, amount=${this.formatAmount(amount, tokenInfo.decimals)} ${tokenInfo.symbol}`);
                        }
                    }
                    catch (error) {
                        this.logger.error(`[${chainName}] Error processing Repay event: ${error.message}`);
                    }
                });
                contract.on('LiquidationCall', async (collateralAsset, debtAsset, user, debtToCover, liquidatedCollateralAmount, liquidator, receiveAToken, event) => {
                    var _a, _b;
                    try {
                        const [collateralInfo, debtInfo] = await Promise.all([
                            this.getTokenInfo(chainName, collateralAsset, provider),
                            this.getTokenInfo(chainName, debtAsset, provider),
                        ]);
                        this.logger.log(`[${chainName}] LiquidationCall event detected:`);
                        this.logger.log(`- Collateral Asset: ${collateralAsset} (${collateralInfo.symbol})`);
                        this.logger.log(`- Debt Asset: ${debtAsset} (${debtInfo.symbol})`);
                        this.logger.log(`- User: ${user}`);
                        this.logger.log(`- Debt to Cover: ${this.formatAmount(debtToCover, debtInfo.decimals)} ${debtInfo.symbol}`);
                        this.logger.log(`- Liquidated Amount: ${this.formatAmount(liquidatedCollateralAmount, collateralInfo.decimals)} ${collateralInfo.symbol}`);
                        this.logger.log(`- Liquidator: ${liquidator}`);
                        this.logger.log(`- Receive AToken: ${receiveAToken}`);
                        this.logger.log(`- Transaction Hash: ${(event === null || event === void 0 ? void 0 : event.transactionHash) || ((_a = event === null || event === void 0 ? void 0 : event.log) === null || _a === void 0 ? void 0 : _a.transactionHash)}`);
                        const accountData = await this.getUserAccountData(contract, user);
                        if (!accountData) {
                            this.logger.warn(`[${chainName}] Could not get account data for user ${user}`);
                            return;
                        }
                        const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
                        const totalDebt = Number(ethers_1.ethers.formatUnits(accountData.totalDebtBase, 8));
                        await this.databaseService.recordLiquidation(chainName, user, liquidator, (event === null || event === void 0 ? void 0 : event.transactionHash) || ((_b = event === null || event === void 0 ? void 0 : event.log) === null || _b === void 0 ? void 0 : _b.transactionHash));
                        this.logger.log(`[${chainName}] Recorded liquidation for user ${user}`);
                        this.logger.log(`[${chainName}] Final health factor: ${healthFactor}`);
                        this.logger.log(`[${chainName}] Final total debt: ${totalDebt.toFixed(2)} USD`);
                        const activeLoansSet = this.activeLoans.get(chainName);
                        if (activeLoansSet) {
                            activeLoansSet.delete(user);
                        }
                    }
                    catch (error) {
                        this.logger.error(`[${chainName}] Error processing LiquidationCall event: ${error.message}`);
                    }
                });
                this.logger.log(`[${chainName}] Successfully set up event listeners and verified contract connection`);
            }
            catch (error) {
                this.logger.error(`Failed to set up event listeners for ${chainName}: ${error.message}`);
            }
        }
    }
    startHealthFactorChecker() {
        this.checkInterval = setInterval(async () => {
            try {
                const loansToCheck = await this.databaseService.getLoansToCheck();
                for (const loan of loansToCheck) {
                    const provider = await this.chainService.getProvider(loan.chainName);
                    const config = this.chainService.getChainConfig(loan.chainName);
                    const abiPath = path.join(process.cwd(), 'abi', `${loan.chainName.toUpperCase()}_AAVE_V3_POOL_ABI.json`);
                    const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
                    const contract = new ethers_1.ethers.Contract(config.contracts.lendingPool, abi, provider);
                    await this.checkHealthFactor(loan.chainName, loan.user, contract);
                }
            }
            catch (error) {
                this.logger.error(`Error in health factor checker: ${error.message}`);
            }
        }, 60000);
    }
    startHeartbeat() {
        this.printHeartbeat();
        this.heartbeatInterval = setInterval(() => {
            this.printHeartbeat();
        }, 60 * 60 * 1000);
    }
    formatDate(date) {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        const seconds = String(date.getSeconds()).padStart(2, '0');
        return `${year}/${month}/${day} ${hours}:${minutes}:${seconds}`;
    }
    printHeartbeat() {
        const chains = this.chainService.getActiveChains();
        this.logger.log(`心跳检测 - 正在监听的合约：`);
        for (const chainName of chains) {
            const config = this.chainService.getChainConfig(chainName);
            this.logger.log(`[${chainName}] LendingPool: ${config.contracts.lendingPool}`);
            this.databaseService.getActiveLoans(chainName)
                .then(activeLoans => {
                this.logger.log(`[${chainName}] 当前活跃贷款数量: ${activeLoans.length}`);
            })
                .catch(error => {
                this.logger.error(`[${chainName}] Error getting active loans count: ${error.message}`);
            });
        }
    }
    async checkHealthFactor(chainName, user, contract) {
        try {
            const accountData = await this.getUserAccountData(contract, user);
            if (!accountData)
                return;
            const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
            const totalDebt = Number(ethers_1.ethers.formatUnits(accountData.totalDebtBase, 8));
            this.logger.log(`[${chainName}] User ${user} health factor: ${healthFactor}`);
            this.logger.log(`[${chainName}] User ${user} total debt: ${totalDebt.toFixed(2)} USD`);
            if (totalDebt === 0) {
                await this.databaseService.deactivateLoan(chainName, user);
                this.logger.log(`[${chainName}] Deactivated loan for user ${user} as total debt is 0`);
                return;
            }
            const waitTime = this.calculateWaitTime(healthFactor);
            const nextCheckTime = new Date(Date.now() + waitTime);
            const formattedDate = this.formatDate(nextCheckTime);
            await this.databaseService.updateLoanHealthFactor(chainName, user, healthFactor, nextCheckTime, totalDebt);
            this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms (at ${formattedDate})`);
            if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
                await this.databaseService.markLiquidationDiscovered(chainName, user);
                await this.executeLiquidation(chainName, user, contract);
                return;
            }
        }
        catch (error) {
            this.logger.error(`[${chainName}] Error checking health factor for user ${user}: ${error.message}`);
        }
    }
    async getUserAccountData(contract, user) {
        try {
            const data = await contract.getUserAccountData(user);
            return {
                totalCollateralBase: data[0],
                totalDebtBase: data[1],
                availableBorrowsBase: data[2],
                currentLiquidationThreshold: data[3],
                ltv: data[4],
                healthFactor: data[5]
            };
        }
        catch (error) {
            this.logger.error(`Error getting user account data: ${error.message}`);
            return null;
        }
    }
    calculateHealthFactor(healthFactor) {
        return Number(healthFactor) / 1e18;
    }
    calculateWaitTime(healthFactor) {
        if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
            return this.MIN_WAIT_TIME;
        }
        if (healthFactor <= this.CRITICAL_THRESHOLD) {
            return this.MIN_WAIT_TIME * 2;
        }
        if (healthFactor <= this.HEALTH_FACTOR_THRESHOLD) {
            const baseTime = this.MIN_WAIT_TIME * 4;
            const maxTime = this.MAX_WAIT_TIME / 2;
            const factor = (healthFactor - this.CRITICAL_THRESHOLD) /
                (this.HEALTH_FACTOR_THRESHOLD - this.CRITICAL_THRESHOLD);
            return Math.floor(baseTime + (maxTime - baseTime) * Math.pow(factor, 2));
        }
        const baseTime = this.MAX_WAIT_TIME / 2;
        const maxTime = this.MAX_WAIT_TIME;
        const factor = (healthFactor - this.HEALTH_FACTOR_THRESHOLD) /
            (2 - this.HEALTH_FACTOR_THRESHOLD);
        return Math.min(Math.floor(baseTime + (maxTime - baseTime) * Math.log1p(factor)), this.MAX_WAIT_TIME);
    }
    async executeLiquidation(chainName, user, contract) {
        try {
            this.logger.log(`[${chainName}] Executing liquidation for user ${user}`);
        }
        catch (error) {
            this.logger.error(`[${chainName}] Error executing liquidation for user ${user}: ${error.message}`);
        }
    }
    recordLiquidationTime(chainName, user) {
        if (!this.liquidationTimes.has(chainName)) {
            this.liquidationTimes.set(chainName, new Map());
        }
        const chainLiquidationTimes = this.liquidationTimes.get(chainName);
        if (chainLiquidationTimes) {
            chainLiquidationTimes.set(user, Date.now());
        }
    }
    async onModuleDestroy() {
        if (this.checkInterval) {
            clearInterval(this.checkInterval);
        }
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
        }
    }
};
exports.BorrowDiscoveryService = BorrowDiscoveryService;
exports.BorrowDiscoveryService = BorrowDiscoveryService = BorrowDiscoveryService_1 = __decorate([
    (0, common_1.Injectable)(),
    __metadata("design:paramtypes", [chain_service_1.ChainService,
        config_1.ConfigService,
        database_service_1.DatabaseService])
], BorrowDiscoveryService);
//# sourceMappingURL=borrow-discovery.service.js.map