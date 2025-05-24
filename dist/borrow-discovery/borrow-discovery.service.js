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
let BorrowDiscoveryService = BorrowDiscoveryService_1 = class BorrowDiscoveryService {
    constructor(chainService, configService, databaseService) {
        this.chainService = chainService;
        this.configService = configService;
        this.databaseService = databaseService;
        this.logger = new common_1.Logger(BorrowDiscoveryService_1.name);
        this.activeLoans = new Map();
        this.liquidationTimes = new Map();
        this.LIQUIDATION_THRESHOLD = 1.05;
        this.CRITICAL_THRESHOLD = 1.1;
        this.HEALTH_FACTOR_THRESHOLD = 1.2;
        this.MIN_WAIT_TIME = this.configService.get('MIN_CHECK_INTERVAL', 200);
        this.MAX_WAIT_TIME = this.configService.get('MAX_CHECK_INTERVAL', 30 * 60 * 1000);
    }
    async onModuleInit() {
        this.logger.log('💓 Heartbeat: BorrowDiscoveryService 已启动，正在监控中...');
        this.logger.log('BorrowDiscoveryService initializing...');
        await this.startListening();
        this.startHealthFactorChecker();
        this.startHeartbeat();
    }
    async startListening() {
        const chains = this.chainService.getActiveChains();
        this.logger.log(`Starting to listen on chains: ${chains.join(', ')}`);
        for (const chainName of chains) {
            try {
                const provider = await this.chainService.getProvider(chainName);
                const config = this.chainService.getChainConfig(chainName);
                const currentBlock = await provider.getBlockNumber();
                this.logger.log(`[${chainName}] Current block number: ${currentBlock}`);
                const code = await provider.getCode(config.contracts.lendingPool);
                if (code === '0x') {
                    this.logger.error(`[${chainName}] No contract code found at address ${config.contracts.lendingPool}`);
                    continue;
                }
                this.logger.log(`[${chainName}] Contract code found at ${config.contracts.lendingPool}`);
                const contract = new ethers_1.ethers.Contract(config.contracts.lendingPool, [
                    'event Borrow(address indexed user, address indexed onBehalfOf, uint256 amount, uint256 interestRateMode, uint256 borrowRate, uint16 indexed referral)',
                    'event Repay(address indexed user, address indexed repayer, uint256 amount, bool useATokens)',
                    'event LiquidationCall(address indexed collateralAsset, address indexed debtAsset, address indexed user, uint256 debtToCover, uint256 liquidatedCollateralAmount, address liquidator, bool receiveAToken)',
                    'function getAddressesProvider() view returns (address)',
                    'function getReserveData(address asset) view returns (tuple(uint256 configuration, uint128 liquidityIndex, uint128 currentLiquidityRate, uint128 variableBorrowIndex, uint128 currentVariableBorrowRate, uint128 currentStableBorrowRate, uint40 lastUpdateTimestamp, uint16 id, address aTokenAddress, address stableDebtTokenAddress, address variableDebtTokenAddress, address interestRateStrategyAddress, uint128 accruedToTreasury, uint128 unbacked, uint128 isolationModeTotalDebt))'
                ], provider);
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
                contract.on('Borrow', async (user, onBehalfOf, amount, interestRateMode, borrowRate, referral, event) => {
                    this.logger.log(`[${chainName}] Borrow event detected: user=${user}, amount=${ethers_1.ethers.formatEther(amount)} ETH`);
                    if (!this.activeLoans.has(chainName)) {
                        this.activeLoans.set(chainName, new Set());
                    }
                    const activeLoansSet = this.activeLoans.get(chainName);
                    if (activeLoansSet) {
                        activeLoansSet.add(onBehalfOf);
                    }
                    await this.checkHealthFactor(chainName, onBehalfOf, contract);
                });
                contract.on('Repay', async (user, repayer, amount, useATokens, event) => {
                    this.logger.log(`[${chainName}] Repay event detected: user=${user}, amount=${ethers_1.ethers.formatEther(amount)} ETH`);
                    const activeLoansSet = this.activeLoans.get(chainName);
                    if (activeLoansSet) {
                        activeLoansSet.delete(user);
                    }
                });
                contract.on('LiquidationCall', async (collateralAsset, debtAsset, user, debtToCover, liquidatedCollateralAmount, liquidator, receiveAToken, event) => {
                    this.logger.log(`[${chainName}] LiquidationCall event detected:`);
                    this.logger.log(`- User: ${user}`);
                    this.logger.log(`- Debt to Cover: ${ethers_1.ethers.formatEther(debtToCover)} ETH`);
                    this.logger.log(`- Liquidated Amount: ${ethers_1.ethers.formatEther(liquidatedCollateralAmount)} ETH`);
                    this.logger.log(`- Liquidator: ${liquidator}`);
                    await this.databaseService.recordLiquidation(chainName, user, liquidator, event.transactionHash);
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
                    const contract = new ethers_1.ethers.Contract(config.contracts.lendingPool, [
                        'function getUserAccountData(address user) view returns (tuple(uint256 totalCollateralBase, uint256 totalDebtBase, uint256 availableBorrowsBase, uint256 currentLiquidationThreshold, uint256 ltv, uint256 healthFactor))'
                    ], provider);
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
    printHeartbeat() {
        const chains = this.chainService.getActiveChains();
        const now = new Date().toISOString();
        this.logger.log(`[${now}] 心跳检测 - 正在监听的合约：`);
        chains.forEach(chainName => {
            const config = this.chainService.getChainConfig(chainName);
            this.logger.log(`[${chainName}] LendingPool: ${config.contracts.lendingPool}`);
            const activeLoansSet = this.activeLoans.get(chainName);
            const activeLoansCount = activeLoansSet ? activeLoansSet.size : 0;
            this.logger.log(`[${chainName}] 当前活跃贷款数量: ${activeLoansCount}`);
        });
    }
    async checkHealthFactor(chainName, user, contract) {
        try {
            const accountData = await this.getUserAccountData(contract, user);
            if (!accountData)
                return;
            const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
            this.logger.log(`[${chainName}] User ${user} health factor: ${healthFactor}`);
            const waitTime = this.calculateWaitTime(healthFactor);
            const nextCheckTime = new Date(Date.now() + waitTime);
            await this.databaseService.updateLoanHealthFactor(chainName, user, healthFactor, nextCheckTime);
            this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms (at ${nextCheckTime})`);
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
        return Math.floor(baseTime + (maxTime - baseTime) * Math.log1p(factor));
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