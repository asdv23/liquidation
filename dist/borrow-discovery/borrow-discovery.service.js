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
let BorrowDiscoveryService = BorrowDiscoveryService_1 = class BorrowDiscoveryService {
    constructor(chainService) {
        this.chainService = chainService;
        this.logger = new common_1.Logger(BorrowDiscoveryService_1.name);
        this.activeLoans = new Map();
        this.liquidationTimes = new Map();
    }
    async onModuleInit() {
        this.logger.log('BorrowDiscoveryService initializing...');
        await this.startListening();
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
                    const activeLoansSet = this.activeLoans.get(chainName);
                    if (activeLoansSet) {
                        activeLoansSet.delete(user);
                    }
                    this.recordLiquidationTime(chainName, user);
                });
                this.logger.log(`[${chainName}] Successfully set up event listeners and verified contract connection`);
            }
            catch (error) {
                this.logger.error(`Failed to set up event listeners for ${chainName}: ${error.message}`);
            }
        }
    }
    async checkHealthFactor(chainName, user, contract) {
        try {
            const healthFactor = await contract.getUserAccountData(user);
            this.logger.log(`[${chainName}] Health factor for user ${user}: ${ethers_1.ethers.formatUnits(healthFactor.healthFactor, 18)}`);
        }
        catch (error) {
            this.logger.error(`Failed to check health factor for user ${user} on ${chainName}: ${error.message}`);
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
};
exports.BorrowDiscoveryService = BorrowDiscoveryService;
exports.BorrowDiscoveryService = BorrowDiscoveryService = BorrowDiscoveryService_1 = __decorate([
    (0, common_1.Injectable)(),
    __metadata("design:paramtypes", [chain_service_1.ChainService])
], BorrowDiscoveryService);
//# sourceMappingURL=borrow-discovery.service.js.map