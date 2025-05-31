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
        this.SAME_ASSET_LIQUIDATION_THRESHOLD = 1.0005;
        this.LIQUIDATION_THRESHOLD = 1.005;
        this.CRITICAL_THRESHOLD = 1.01;
        this.HEALTH_FACTOR_THRESHOLD = 1.1;
        this.MIN_DEBT = 5;
        this.abiCache = new Map();
        this.MIN_WAIT_TIME = this.configService.get('MIN_CHECK_INTERVAL', 200);
        this.MAX_WAIT_TIME = this.configService.get('MAX_CHECK_INTERVAL', 4 * 60 * 60 * 1000);
        this.BATCH_CHECK_TIMEOUT = this.configService.get('BATCH_CHECK_TIMEOUT', 5000);
        this.PRIVATE_KEY = this.configService.get('PRIVATE_KEY');
    }
    async onModuleInit() {
        this.logger.log('BorrowDiscoveryService initializing...');
        await this.initializeAbis();
        await this.loadTokenCache();
        await this.loadActiveLoans();
        await this.startListening();
        this.startHealthFactorChecker();
    }
    async onModuleDestroy() {
        this.logger.log('BorrowDiscoveryService destroying...');
    }
    async initializeAbis() {
        try {
            const abiPaths = {
                aaveV3Pool: path.join(process.cwd(), 'abi', 'Pool.abi.json'),
                multicall: path.join(process.cwd(), 'abi', 'Multicall3.abi.json'),
                flashLoanLiquidation: path.join(process.cwd(), 'abi', 'FlashLoanLiquidation.abi.json'),
                addressesProvider: path.join(process.cwd(), 'abi', 'PoolAddressesProvider.abi.json'),
                dataProvider: path.join(process.cwd(), 'abi', 'AaveProtocolDataProvider.abi.json'),
                priceOracle: path.join(process.cwd(), 'abi', 'AaveOracle.abi.json')
            };
            for (const [name, abiPath] of Object.entries(abiPaths)) {
                const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
                this.abiCache.set(name, abi);
            }
            this.logger.log('All ABIs initialized successfully');
        }
        catch (error) {
            this.logger.error(`Error initializing ABIs: ${error.message}`);
            throw error;
        }
    }
    async getSigner(chainName) {
        const provider = await this.chainService.getProvider(chainName);
        return new ethers_1.ethers.Wallet(this.PRIVATE_KEY, provider);
    }
    getAbi(name) {
        const abi = this.abiCache.get(name);
        if (!abi) {
            throw new Error(`Abi not initialized for chain ${name}`);
        }
        return abi;
    }
    async getMulticall(chainName) {
        const signer = await this.getSigner(chainName);
        const multicallContract = new ethers_1.ethers.Contract('0xcA11bde05977b3631167028862bE2a173976CA11', this.getAbi('multicall'), signer);
        return multicallContract;
    }
    async getAaveV3Pool(chainName) {
        const signer = await this.getSigner(chainName);
        const config = this.chainService.getChainConfig(chainName);
        const contract = new ethers_1.ethers.Contract(config.contracts.aavev3Pool, this.getAbi('aaveV3Pool'), signer);
        return contract;
    }
    async getFlashLoanLiquidation(chainName) {
        const signer = await this.getSigner(chainName);
        const config = this.chainService.getChainConfig(chainName);
        const flashLoanLiquidation = new ethers_1.ethers.Contract(config.contracts.flashLoanLiquidation, this.getAbi('flashLoanLiquidation'), signer);
        return flashLoanLiquidation;
    }
    async getDataProvider(chainName) {
        const signer = await this.getSigner(chainName);
        const aaveV3Pool = await this.getAaveV3Pool(chainName);
        const addressesProviderAddress = await aaveV3Pool.ADDRESSES_PROVIDER();
        const addressesProvider = new ethers_1.ethers.Contract(addressesProviderAddress, this.getAbi('addressesProvider'), signer);
        const dataProviderAddress = await addressesProvider.getPoolDataProvider();
        const dataProvider = new ethers_1.ethers.Contract(dataProviderAddress, this.getAbi('dataProvider'), signer);
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
        const providerToUse = provider || await this.chainService.getProvider(chainName);
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
                    for (const loan of activeLoans) {
                        activeLoansMap.set(loan.user, {
                            nextCheckTime: loan.nextCheckTime,
                            healthFactor: loan.healthFactor
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
                this.logger.log(`[${chainName}] 🩷 Borrow event detected:`);
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
                user = user.toLowerCase();
                const [collateralInfo, debtInfo] = await Promise.all([
                    this.getTokenInfo(chainName, collateralAsset, provider),
                    this.getTokenInfo(chainName, debtAsset, provider),
                ]);
                this.logger.log(`[${chainName}] 😄 LiquidationCall event detected:`);
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
    async startListening() {
        const chains = this.chainService.getActiveChains();
        this.logger.log(`Starting to listen on chains: ${chains.join(', ')}`);
        await Promise.all(chains.map(async (chainName) => {
            try {
                const provider = await this.chainService.getProvider(chainName);
                const config = this.chainService.getChainConfig(chainName);
                const currentBlock = await provider.getBlockNumber();
                this.logger.log(`[${chainName}] Current block number: ${currentBlock}`);
                const code = await provider.getCode(config.contracts.aavev3Pool);
                if (code === '0x') {
                    this.logger.error(`[${chainName}] No contract code found at address ${config.contracts.aavev3Pool}`);
                    return;
                }
                const code2 = await provider.getCode(config.contracts.flashLoanLiquidation);
                if (code2 === '0x') {
                    this.logger.error(`[${chainName}] No contract code found at address ${config.contracts.flashLoanLiquidation}`);
                    return;
                }
                this.logger.log(`[${chainName}] Contract code found at ${config.contracts.aavev3Pool}, ${config.contracts.flashLoanLiquidation}`);
                const aaveV3Pool = await this.getAaveV3Pool(chainName);
                await this.setupEventListeners(chainName, aaveV3Pool, provider);
                this.logger.log(`[${chainName}] Successfully set up event listeners and verified contract connection`);
            }
            catch (error) {
                this.logger.error(`Failed to set up event listeners for ${chainName}: ${error.message}`);
            }
        }));
    }
    startHealthFactorChecker() {
        let isChecking = false;
        const checkAllLoans = async () => {
            if (isChecking) {
                return;
            }
            isChecking = true;
            try {
                const chains = Array.from(this.activeLoans.keys());
                await Promise.all(chains.map(chainName => this.checkHealthFactorsBatch(chainName)));
            }
            catch (error) {
                this.logger.error(`Error in health factor checker: ${error.message}`);
            }
            finally {
                isChecking = false;
                setTimeout(checkAllLoans, this.MIN_WAIT_TIME);
            }
        };
        checkAllLoans();
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
            const usersToCheck = Array.from(activeLoansMap.entries())
                .filter(([_, info]) => info.nextCheckTime <= new Date())
                .map(([user]) => user);
            if (usersToCheck.length === 0)
                return;
            const BATCH_SIZE = 100;
            const batches = [];
            for (let i = 0; i < usersToCheck.length; i += BATCH_SIZE) {
                const batchUsers = usersToCheck.slice(i, i + BATCH_SIZE);
                batches.push(batchUsers);
            }
            this.logger.log(`[${chainName}] Processing ${batches.length} batches concurrently...`);
            await Promise.all(batches.map(async (batchUsers, batchIndex) => {
                try {
                    this.logger.log(`[${chainName}] Processing batch ${batchIndex + 1}/${batches.length} (${batchUsers.length} users)...`);
                    await Promise.race([
                        this.processBatch(chainName, batchUsers, activeLoansMap),
                        new Promise((_, reject) => setTimeout(() => reject(new Error(`Batch check timeout after ${this.BATCH_CHECK_TIMEOUT}ms`)), this.BATCH_CHECK_TIMEOUT))
                    ]);
                }
                catch (error) {
                    this.logger.error(`[${chainName}] Error processing batch ${batchIndex + 1}/${batches.length}: ${error.message}`);
                }
            }));
            this.logger.log(`[${chainName}] Completed processing all ${batches.length} batches`);
        }
        catch (error) {
            this.logger.error(`[${chainName}] Error checking health factors batch: ${error.message}`);
        }
    }
    async processBatch(chainName, batchUsers, activeLoansMap) {
        const aaveV3Pool = await this.getAaveV3Pool(chainName);
        const accountDataMap = await this.getUserAccountDataBatch(chainName, aaveV3Pool, batchUsers);
        const lastLiquidationMap = this.lastLiquidationAttempt.get(chainName);
        await Promise.all(batchUsers.map(async (user) => {
            const accountData = accountDataMap.get(user);
            if (!accountData)
                return;
            await this.processUser(chainName, user, accountData, activeLoansMap, lastLiquidationMap, aaveV3Pool);
        }));
    }
    async processUser(chainName, user, accountData, activeLoansMap, lastLiquidationMap, aaveV3Pool) {
        const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
        const totalDebt = Number(ethers_1.ethers.formatUnits(accountData.totalDebtBase, 8));
        if (totalDebt < this.MIN_DEBT) {
            activeLoansMap.delete(user);
            lastLiquidationMap.delete(user);
            await this.databaseService.deactivateLoan(chainName, user);
            this.logger.log(`[${chainName}] Removed user ${user} from active loans and database as total debt is less than ${this.MIN_DEBT} USD`);
            return;
        }
        if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
            const lastAttemptHealthFactor = lastLiquidationMap.get(user);
            if (lastAttemptHealthFactor === undefined || healthFactor < lastAttemptHealthFactor.healthFactor) {
                this.logger.log(`[${chainName}] Liquidation threshold ${healthFactor} <= ${this.LIQUIDATION_THRESHOLD} reached for user ${user}, attempting liquidation`);
                lastLiquidationMap.set(user, { healthFactor: healthFactor, retryCount: (lastAttemptHealthFactor === null || lastAttemptHealthFactor === void 0 ? void 0 : lastAttemptHealthFactor.retryCount) + 1 || 1 });
                await this.executeLiquidation(chainName, user, healthFactor, aaveV3Pool);
            }
            else {
                this.logger.log(`[${chainName}] Skip liquidation for ${user} as health factor ${healthFactor} >= ${lastAttemptHealthFactor.healthFactor}, retry ${lastAttemptHealthFactor.retryCount}`);
            }
            return;
        }
        if (lastLiquidationMap.has(user)) {
            lastLiquidationMap.delete(user);
        }
        const waitTime = this.calculateWaitTime(chainName, healthFactor);
        const nextCheckTime = new Date(Date.now() + waitTime);
        const formattedDate = this.formatDate(nextCheckTime);
        this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms (at ${formattedDate}), healthFactor: ${healthFactor}`);
        activeLoansMap.set(user, {
            nextCheckTime: nextCheckTime,
            healthFactor: healthFactor
        });
        await this.databaseService.updateLoanHealthFactor(chainName, user, healthFactor, nextCheckTime);
    }
    async getUserAccountDataBatch(chainName, contract, users) {
        try {
            const multicallContract = await this.getMulticall(chainName);
            const calls = users.map(user => ({
                target: contract.target,
                callData: contract.interface.encodeFunctionData('getUserAccountData', [user])
            }));
            const [, returnData] = await multicallContract.aggregate.staticCall(calls);
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
    calculateWaitTime(chainName, healthFactor) {
        const minWaitTime = this.chainService.getChainConfig(chainName).minWaitTime;
        if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
            return minWaitTime;
        }
        if (healthFactor <= this.CRITICAL_THRESHOLD) {
            return minWaitTime * 2;
        }
        if (healthFactor <= this.HEALTH_FACTOR_THRESHOLD) {
            const baseTime = minWaitTime * 4;
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
    async executeLiquidation(chainName, user, healthFactor, aaveV3Pool) {
        try {
            const multicall = await this.getMulticall(chainName);
            const calls = [
                {
                    target: aaveV3Pool.target, callData: aaveV3Pool.interface.encodeFunctionData('getUserConfiguration', [user])
                },
                {
                    target: aaveV3Pool.target, callData: aaveV3Pool.interface.encodeFunctionData('getReservesList')
                }
            ];
            const [, returnData] = await multicall.aggregate.staticCall(calls);
            const [userConfig,] = aaveV3Pool.interface.decodeFunctionResult('getUserConfiguration', returnData[0]);
            const [reservesList,] = aaveV3Pool.interface.decodeFunctionResult('getReservesList', returnData[1]);
            const dataProvider = await this.getDataProvider(chainName);
            const reserveCalls = [];
            const borrowingAssets = [];
            const collateralAssets = [];
            for (let i = 0; i < reservesList.length; i++) {
                const asset = reservesList[i];
                const isBorrowing = (BigInt(userConfig.data) >> (BigInt(i) << BigInt(1))) !== BigInt(0);
                const isUsingAsCollateral = ((BigInt(userConfig.data) >> ((BigInt(i) << BigInt(1)) + BigInt(1))) & BigInt(1)) !== BigInt(0);
                if (isBorrowing || isUsingAsCollateral) {
                    reserveCalls.push({
                        target: dataProvider.target,
                        callData: dataProvider.interface.encodeFunctionData('getUserReserveData', [asset, user])
                    });
                    if (isBorrowing) {
                        borrowingAssets.push(asset);
                    }
                    if (isUsingAsCollateral) {
                        collateralAssets.push(asset);
                    }
                }
            }
            if (reserveCalls.length === 0) {
                this.logger.log(`[${chainName}] No borrowing or collateral assets found for user ${user}`);
                return;
            }
            const [, reserveReturnData] = await multicall.aggregate.staticCall(reserveCalls);
            let maxDebtAsset = null;
            let maxDebtAmount = BigInt(0);
            let maxCollateralAsset = null;
            let maxCollateralAmount = BigInt(0);
            let callIndex = 0;
            for (let i = 0; i < reservesList.length; i++) {
                const asset = reservesList[i];
                const isBorrowing = (BigInt(userConfig.data) >> (BigInt(i) << BigInt(1))) !== BigInt(0);
                const isUsingAsCollateral = ((BigInt(userConfig.data) >> ((BigInt(i) << BigInt(1)) + BigInt(1))) & BigInt(1)) !== BigInt(0);
                if (isBorrowing || isUsingAsCollateral) {
                    const userReserveData = dataProvider.interface.decodeFunctionResult('getUserReserveData', reserveReturnData[callIndex]);
                    callIndex++;
                    if (isBorrowing) {
                        const currentStableDebt = BigInt(userReserveData.currentStableDebt);
                        const currentVariableDebt = BigInt(userReserveData.currentVariableDebt);
                        const totalDebt = currentStableDebt + currentVariableDebt;
                        if (totalDebt > maxDebtAmount) {
                            maxDebtAmount = totalDebt;
                            maxDebtAsset = asset;
                        }
                    }
                    if (isUsingAsCollateral) {
                        const collateralAmount = BigInt(userReserveData.currentATokenBalance);
                        if (collateralAmount > maxCollateralAmount) {
                            maxCollateralAmount = collateralAmount;
                            maxCollateralAsset = asset;
                        }
                    }
                }
            }
            if (maxDebtAmount === BigInt(0)) {
                this.logger.log(`[${chainName}] No debt found for user ${user}`);
                return;
            }
            if (maxCollateralAmount === BigInt(0)) {
                this.logger.log(`[${chainName}] No collateral assets found for user ${user}`);
                return;
            }
            if (maxCollateralAsset == maxDebtAsset && healthFactor > this.SAME_ASSET_LIQUIDATION_THRESHOLD) {
                this.logger.log(`[${chainName}] Skip liquidation for ${user} as same asset and health factor ${healthFactor} > ${this.SAME_ASSET_LIQUIDATION_THRESHOLD}`);
                return;
            }
            const collateralTokenInfo = await this.getTokenInfo(chainName, maxCollateralAsset);
            const debtTokenInfo = await this.getTokenInfo(chainName, maxDebtAsset);
            this.logger.log(`[${chainName}] 💰 Executing flash loan liquidation:`);
            this.logger.log(`- User: ${user}`);
            this.logger.log(`- Health Factor: ${healthFactor}`);
            this.logger.log(`- Collateral Asset: ${maxCollateralAsset} (${(Number(maxCollateralAmount) / Number(10 ** collateralTokenInfo.decimals)).toFixed(6)} ${collateralTokenInfo.symbol})`);
            this.logger.log(`- Debt Asset: ${maxDebtAsset} (${(Number(maxDebtAmount) / Number(10 ** debtTokenInfo.decimals)).toFixed(6)} ${debtTokenInfo.symbol})`);
            const flashLoanLiquidation = await this.getFlashLoanLiquidation(chainName);
            try {
                const gasPrice = await flashLoanLiquidation.runner.provider.getFeeData();
                const maxPriorityFeePerGas = gasPrice.maxPriorityFeePerGas ? gasPrice.maxPriorityFeePerGas * BigInt(15) / BigInt(10) : ethers_1.ethers.parseUnits('1', 'gwei');
                const maxFeePerGas = gasPrice.maxFeePerGas ? gasPrice.maxFeePerGas + (maxPriorityFeePerGas || BigInt(0)) : undefined;
                this.logger.log(`[${chainName}] gasPrice: ${gasPrice.gasPrice}, maxFeePerGas: ${maxFeePerGas}, maxPriorityFeePerGas: ${maxPriorityFeePerGas}`);
                const tx = await flashLoanLiquidation.executeLiquidation(maxCollateralAsset, maxDebtAsset, user, {
                    maxFeePerGas,
                    maxPriorityFeePerGas,
                });
                this.logger.log(`[${chainName}] Flash loan liquidation executed successfully, tx: ${tx.hash}`);
                await tx.wait();
            }
            catch (error) {
                this.logger.error(`[${chainName}] Error executing flash loan liquidation for user ${user}: ${error.message}`);
            }
        }
        catch (error) {
            this.logger.error(`[${chainName}] Error executing liquidation for user ${user}: ${error.message}`);
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