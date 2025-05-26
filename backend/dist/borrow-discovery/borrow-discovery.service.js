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
        this.tokenCache = new Map();
        this.lastLiquidationAttempt = new Map();
        this.LIQUIDATION_THRESHOLD = 1.0001;
        this.CRITICAL_THRESHOLD = 1.001;
        this.HEALTH_FACTOR_THRESHOLD = 1.01;
        this.contractCache = new Map();
        this.providerCache = new Map();
        this.signerCache = new Map();
        this.dataProviderCache = new Map();
        this.priceOracleCache = new Map();
        this.MIN_WAIT_TIME = this.configService.get('MIN_CHECK_INTERVAL', 1000);
        this.MAX_WAIT_TIME = this.configService.get('MAX_CHECK_INTERVAL', 30 * 60 * 1000);
        this.PRIVATE_KEY = this.configService.get('PRIVATE_KEY');
    }
    async onModuleInit() {
        this.logger.log('BorrowDiscoveryService initializing...');
        await this.initializeResources();
        await this.loadTokenCache();
        await this.loadActiveLoans();
        await this.startListening();
        this.startHealthFactorChecker();
    }
    async initializeResources() {
        const chains = this.chainService.getActiveChains();
        for (const chainName of chains) {
            try {
                const provider = await this.chainService.getProvider(chainName);
                this.providerCache.set(chainName, provider);
                const signer = new ethers_1.ethers.Wallet(this.PRIVATE_KEY, provider);
                this.signerCache.set(chainName, signer);
                const config = this.chainService.getChainConfig(chainName);
                const abiPath = path.join(process.cwd(), 'abi', `${chainName.toUpperCase()}_AAVE_V3_POOL_ABI.json`);
                const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
                const contract = new ethers_1.ethers.Contract(config.contracts.lendingPool, abi, signer);
                this.contractCache.set(chainName, contract);
                const addressesProviderAddress = await contract.ADDRESSES_PROVIDER();
                const addressesProviderAbi = [
                    'function getPoolDataProvider() view returns (address)',
                    'function getPriceOracle() view returns (address)'
                ];
                const addressesProvider = new ethers_1.ethers.Contract(addressesProviderAddress, addressesProviderAbi, signer);
                const dataProviderAddress = await addressesProvider.getPoolDataProvider();
                const dataProviderAbi = [
                    'function getUserReserveData(address asset, address user) view returns (uint256 currentATokenBalance, uint256 currentStableDebt, uint256 currentVariableDebt, uint256 principalStableDebt, uint256 scaledVariableDebt, uint256 stableBorrowRate, uint256 liquidityRate, uint40 stableRateLastUpdated, bool usageAsCollateralEnabled)'
                ];
                const dataProvider = new ethers_1.ethers.Contract(dataProviderAddress, dataProviderAbi, signer);
                this.dataProviderCache.set(chainName, dataProvider);
                const priceOracleAddress = await addressesProvider.getPriceOracle();
                const priceOracleAbi = [
                    'function getAssetPrice(address asset) view returns (uint256)'
                ];
                const priceOracle = new ethers_1.ethers.Contract(priceOracleAddress, priceOracleAbi, signer);
                this.priceOracleCache.set(chainName, priceOracle);
                this.logger.log(`[${chainName}] Initialized provider, signer, contract, dataProvider and priceOracle`);
            }
            catch (error) {
                this.logger.error(`[${chainName}] Failed to initialize resources: ${error.message}`);
            }
        }
    }
    getContract(chainName) {
        const contract = this.contractCache.get(chainName);
        if (!contract) {
            throw new Error(`Contract not initialized for chain ${chainName}`);
        }
        return contract;
    }
    getDataProvider(chainName) {
        const dataProvider = this.dataProviderCache.get(chainName);
        if (!dataProvider) {
            throw new Error(`DataProvider not initialized for chain ${chainName}`);
        }
        return dataProvider;
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
        const providerToUse = provider || this.providerCache.get(chainName);
        if (!providerToUse) {
            throw new Error(`Provider not initialized for chain ${chainName}`);
        }
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
            const contract = new ethers_1.ethers.Contract(normalizedAddress, erc20Abi, providerToUse);
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
                    this.logger.log(`[${chainName}] Loading active loans into memory...`);
                    if (!this.activeLoans.has(chainName)) {
                        this.activeLoans.set(chainName, new Map());
                    }
                    const activeLoansMap = this.activeLoans.get(chainName);
                    const now = new Date();
                    for (const loan of activeLoans) {
                        activeLoansMap.set(loan.user, {
                            nextCheckTime: now,
                            healthFactor: 1.0
                        });
                    }
                    this.logger.log(`[${chainName}] Loaded ${activeLoansMap.size} active loans into memory, will check immediately`);
                }
            }
        }
        catch (error) {
            this.logger.error(`Error loading active loans: ${error.message}`);
        }
    }
    createBorrowEventHandler(chainName, provider) {
        return async (reserve, user, onBehalfOf, amount, interestRateMode, borrowRate, referralCode, event) => {
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
                    this.activeLoans.set(chainName, new Map());
                }
                const activeLoansMap = this.activeLoans.get(chainName);
                if (activeLoansMap) {
                    activeLoansMap.set(onBehalfOf, {
                        nextCheckTime: new Date(),
                        healthFactor: 1.0
                    });
                }
                await this.databaseService.createOrUpdateLoan(chainName, onBehalfOf);
                this.logger.log(`[${chainName}] Created/Updated loan record for user ${onBehalfOf}`);
            }
            catch (error) {
                this.logger.error(`[${chainName}] Error processing Borrow event: ${error.message}`);
            }
        };
    }
    createLiquidationCallEventHandler(chainName, provider) {
        return async (collateralAsset, debtAsset, user, debtToCover, liquidatedCollateralAmount, liquidator, receiveAToken, event) => {
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
                const activeLoansMap = this.activeLoans.get(chainName);
                if (activeLoansMap && activeLoansMap.has(user)) {
                    activeLoansMap.delete(user);
                    await this.databaseService.recordLiquidation(chainName, user, liquidator, (event === null || event === void 0 ? void 0 : event.transactionHash) || ((_b = event === null || event === void 0 ? void 0 : event.log) === null || _b === void 0 ? void 0 : _b.transactionHash));
                    this.logger.log(`[${chainName}] Recorded liquidation for user ${user}`);
                }
                else {
                    this.logger.log(`[${chainName}] No loan found for user ${user}, skipping liquidation record`);
                }
            }
            catch (error) {
                this.logger.error(`[${chainName}] Error processing LiquidationCall event: ${error.message}`);
            }
        };
    }
    async setupEventListeners(chainName, contract, provider) {
        contract.removeAllListeners('Borrow');
        contract.removeAllListeners('LiquidationCall');
        contract.on('Borrow', this.createBorrowEventHandler(chainName, provider));
        contract.on('LiquidationCall', this.createLiquidationCallEventHandler(chainName, provider));
    }
    async reinitializeEventListeners(chainName) {
        try {
            const provider = await this.chainService.getProvider(chainName);
            const contract = this.getContract(chainName);
            this.logger.log(`[${chainName}] Reinitializing event listeners...`);
            await this.setupEventListeners(chainName, contract, provider);
            this.logger.log(`[${chainName}] Event listeners reinitialized successfully`);
        }
        catch (error) {
            this.logger.error(`[${chainName}] Failed to reinitialize event listeners: ${error.message}`);
        }
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
                const contract = this.getContract(chainName);
                const ws = provider.websocket;
                ws.on('close', async () => {
                    this.logger.warn(`[${chainName}] WebSocket connection closed, will attempt to reinitialize event listeners after reconnection...`);
                });
                ws.on('open', async () => {
                    this.logger.log(`[${chainName}] WebSocket connection reopened, reinitializing event listeners...`);
                    await this.reinitializeEventListeners(chainName);
                });
                await this.setupEventListeners(chainName, contract, provider);
                this.logger.log(`[${chainName}] Successfully set up event listeners and verified contract connection`);
            }
            catch (error) {
                this.logger.error(`Failed to set up event listeners for ${chainName}: ${error.message}`);
            }
        }
    }
    startHealthFactorChecker() {
        let isChecking = false;
        const checkAllLoans = async () => {
            if (isChecking) {
                return;
            }
            isChecking = true;
            try {
                for (const chainName of this.activeLoans.keys()) {
                    await this.checkHealthFactorsBatch(chainName);
                }
            }
            catch (error) {
                this.logger.error(`Error in health factor checker: ${error.message}`);
            }
            finally {
                isChecking = false;
            }
        };
        checkAllLoans();
        this.checkInterval = setInterval(checkAllLoans, this.MIN_WAIT_TIME);
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
    async checkHealthFactorsBatch(chainName) {
        try {
            const activeLoansMap = this.activeLoans.get(chainName);
            if (!activeLoansMap || activeLoansMap.size === 0)
                return;
            if (!this.lastLiquidationAttempt.has(chainName)) {
                this.lastLiquidationAttempt.set(chainName, new Map());
            }
            const lastLiquidationMap = this.lastLiquidationAttempt.get(chainName);
            const usersToCheck = Array.from(activeLoansMap.entries())
                .filter(([_, info]) => info.nextCheckTime <= new Date())
                .map(([user]) => user);
            if (usersToCheck.length === 0)
                return;
            const BATCH_SIZE = 100;
            for (let i = 0; i < usersToCheck.length; i += BATCH_SIZE) {
                const batchUsers = usersToCheck.slice(i, i + BATCH_SIZE);
                this.logger.log(`[${chainName}] Checking health factors for batch ${i / BATCH_SIZE + 1}/${Math.ceil(usersToCheck.length / BATCH_SIZE)} (${batchUsers.length}/${activeLoansMap.size} users)...`);
                const contract = this.getContract(chainName);
                const accountDataMap = await this.getUserAccountDataBatch(contract, batchUsers);
                for (const user of batchUsers) {
                    const accountData = accountDataMap.get(user);
                    if (!accountData)
                        continue;
                    const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
                    const totalDebt = Number(ethers_1.ethers.formatUnits(accountData.totalDebtBase, 8));
                    if (totalDebt < 1) {
                        activeLoansMap.delete(user);
                        lastLiquidationMap.delete(user);
                        await this.databaseService.deactivateLoan(chainName, user);
                        this.logger.log(`[${chainName}] Removed user ${user} from active loans and database as total debt is less than 1 USD`);
                        continue;
                    }
                    if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
                        const lastAttemptHealthFactor = lastLiquidationMap.get(user);
                        if (lastAttemptHealthFactor === undefined || healthFactor < lastAttemptHealthFactor) {
                            this.logger.log(`[${chainName}] Liquidation threshold ${healthFactor} <= ${this.LIQUIDATION_THRESHOLD} reached for user ${user}, attempting liquidation`);
                            lastLiquidationMap.set(user, healthFactor);
                            await this.executeLiquidation(chainName, user, healthFactor, contract);
                        }
                        else {
                            this.logger.log(`[${chainName}] Skipping liquidation for user ${user} as current health factor ${healthFactor} is not lower than last attempt ${lastAttemptHealthFactor}`);
                        }
                        continue;
                    }
                    if (lastLiquidationMap.has(user)) {
                        lastLiquidationMap.delete(user);
                    }
                    const waitTime = this.calculateWaitTime(healthFactor);
                    const nextCheckTime = new Date(Date.now() + waitTime);
                    const formattedDate = this.formatDate(nextCheckTime);
                    this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms (at ${formattedDate}), healthFactor: ${healthFactor}`);
                    activeLoansMap.set(user, {
                        nextCheckTime: new Date(Date.now() + waitTime),
                        healthFactor: healthFactor
                    });
                }
            }
        }
        catch (error) {
            this.logger.error(`[${chainName}] Error checking health factors batch: ${error.message}`);
        }
    }
    async getUserAccountDataBatch(contract, users) {
        try {
            const multicallAddress = '0xcA11bde05977b3631167028862bE2a173976CA11';
            const multicallAbi = [
                'function aggregate(tuple(address target, bytes callData)[] calls) view returns (uint256 blockNumber, bytes[] returnData)'
            ];
            const multicallContract = new ethers_1.ethers.Contract(multicallAddress, multicallAbi, contract.runner);
            const calls = users.map(user => ({
                target: contract.target,
                callData: contract.interface.encodeFunctionData('getUserAccountData', [user])
            }));
            const [, returnData] = await multicallContract.aggregate(calls);
            const results = new Map();
            for (let i = 0; i < users.length; i++) {
                const decodedData = contract.interface.decodeFunctionResult('getUserAccountData', returnData[i]);
                results.set(users[i], {
                    totalCollateralBase: decodedData[0],
                    totalDebtBase: decodedData[1],
                    availableBorrowsBase: decodedData[2],
                    currentLiquidationThreshold: decodedData[3],
                    ltv: decodedData[4],
                    healthFactor: decodedData[5]
                });
            }
            return results;
        }
        catch (error) {
            this.logger.error(`Error getting user account data batch: ${error.message}`);
            return new Map();
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
    async executeLiquidation(chainName, user, healthFactor, contract) {
        try {
            const userConfig = await contract.getUserConfiguration(user);
            const reservesList = await contract.getReservesList();
            const multicallAddress = '0xcA11bde05977b3631167028862bE2a173976CA11';
            const multicallAbi = [
                'function aggregate(tuple(address target, bytes callData)[] calls) view returns (uint256 blockNumber, bytes[] returnData)'
            ];
            const multicallContract = new ethers_1.ethers.Contract(multicallAddress, multicallAbi, contract.runner);
            const dataProvider = this.getDataProvider(chainName);
            const calls = [];
            const borrowingAssets = [];
            for (let i = 0; i < reservesList.length; i++) {
                const asset = reservesList[i];
                const isBorrowing = (BigInt(userConfig.data) >> (BigInt(i) << BigInt(1))) !== BigInt(0);
                if (isBorrowing) {
                    borrowingAssets.push(asset);
                    calls.push({
                        target: dataProvider.target,
                        callData: dataProvider.interface.encodeFunctionData('getUserReserveData', [asset, user])
                    });
                }
            }
            if (calls.length === 0) {
                this.logger.log(`[${chainName}] No borrowing assets found for user ${user}`);
                return;
            }
            const [, returnData] = await multicallContract.aggregate(calls);
            let maxDebtAsset = null;
            let maxDebtAmount = BigInt(0);
            for (let i = 0; i < borrowingAssets.length; i++) {
                const asset = borrowingAssets[i];
                const userReserveData = dataProvider.interface.decodeFunctionResult('getUserReserveData', returnData[i]);
                const currentStableDebt = BigInt(userReserveData.currentStableDebt);
                const currentVariableDebt = BigInt(userReserveData.currentVariableDebt);
                const totalDebt = currentStableDebt + currentVariableDebt;
                if (totalDebt > maxDebtAmount) {
                    maxDebtAmount = totalDebt;
                    maxDebtAsset = {
                        asset,
                        currentStableDebt,
                        currentVariableDebt,
                        usageAsCollateralEnabled: userReserveData.usageAsCollateralEnabled
                    };
                }
            }
            if (!maxDebtAsset) {
                this.logger.log(`[${chainName}] No debt found for user ${user}`);
                return;
            }
            const tokenPrice = await this.getTokenPrice(chainName, maxDebtAsset.asset);
            const tokenInfo = await this.getTokenInfo(chainName, maxDebtAsset.asset);
            const debtValueInUsd = Number(maxDebtAmount) * tokenPrice / (10 ** tokenInfo.decimals);
            this.logger.log(`[${chainName}] Executing liquidation for user ${user}, healthFactor: ${healthFactor}, maxDebtAsset: ${maxDebtAsset.asset} (${tokenInfo.symbol}), debtAmount: ${ethers_1.ethers.formatUnits(maxDebtAmount, tokenInfo.decimals)} ${tokenInfo.symbol}, debtValueInUsd: ${debtValueInUsd.toFixed(2)} USD`);
        }
        catch (error) {
            this.logger.error(`[${chainName}] Error executing liquidation for user ${user}: ${error.message}`);
        }
    }
    async getTokenPrice(chainName, tokenAddress) {
        const priceOracle = this.priceOracleCache.get(chainName);
        const price = await priceOracle.getAssetPrice(tokenAddress);
        return Number(price) / 1e8;
    }
    async onModuleDestroy() {
        if (this.checkInterval) {
            clearInterval(this.checkInterval);
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