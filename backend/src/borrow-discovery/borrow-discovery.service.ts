import { Injectable, Logger, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { ChainService } from '../chain/chain.service';
import { ethers, MaxUint256 } from 'ethers';
import { ConfigService } from '@nestjs/config';
import { DatabaseService } from '../database/database.service';
import * as fs from 'fs';
import * as path from 'path';
import { UserAccountData, TokenInfo, LoanInfo, LiquidationInfo } from './interfaces';
import axios from 'axios';

@Injectable()
export class BorrowDiscoveryService implements OnModuleInit, OnModuleDestroy {
    private readonly logger = new Logger(BorrowDiscoveryService.name);
    private activeLoans: Map<string, Map<string, LoanInfo>> = new Map();
    private tokenCache: Map<string, Map<string, TokenInfo>> = new Map();
    private liquidationInfoCache: Map<string, LiquidationInfo> = new Map();
    private readonly LIQUIDATION_THRESHOLD = 1.00000005; // 清算阈值
    private readonly LIQUIDATION_THRESHOLD_2 = 1.001; // 清算阈值
    private readonly HEALTH_FACTOR_THRESHOLD = 2; // 健康阈值
    private readonly MIN_WAIT_TIME: number; // 最小等待时间（毫秒）
    private readonly MAX_WAIT_TIME: number; // 最大等待时间（毫秒）
    private readonly BATCH_CHECK_TIMEOUT: number; // 批次检查超时时间（毫秒）
    private readonly CACHE_TTL = 45000; // 45 seconds in milliseconds
    private abiCache: Map<string, any> = new Map(); // 新增：ABI 缓存

    constructor(
        private readonly chainService: ChainService,
        private readonly configService: ConfigService,
        private readonly databaseService: DatabaseService,
    ) {
        // 从环境变量读取配置，默认值：最小1s，最大4小时
        this.MIN_WAIT_TIME = this.configService.get<number>('MIN_CHECK_INTERVAL', 200);
        this.MAX_WAIT_TIME = this.configService.get<number>('MAX_CHECK_INTERVAL', 4 * 60 * 60 * 1000);
        this.BATCH_CHECK_TIMEOUT = this.configService.get<number>('BATCH_CHECK_TIMEOUT', 5000); // 默认 5 秒
    }

    async onModuleInit() {
        this.logger.log('BorrowDiscoveryService initializing...');
        await this.initializeAbis(); // 新增：初始化 ABI
        await this.loadTokenCache();
        await this.loadActiveLoans();
        await this.startListening();
        this.startHealthFactorChecker();
    }

    async onModuleDestroy() {
        this.logger.log('BorrowDiscoveryService destroying...');
    }

    private async initializeAbis() {
        try {
            // 初始化所有需要的 ABI
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
        } catch (error) {
            this.logger.error(`Error initializing ABIs: ${error.message}`);
            throw error;
        }
    }

    private getAbi(name: string): any {
        const abi = this.abiCache.get(name);
        if (!abi) {
            throw new Error(`Abi not initialized for chain ${name}`);
        }
        return abi;
    }

    private async getMulticall(chainName: string): Promise<ethers.Contract> {
        const signer = await this.chainService.getSigner(chainName);
        const multicallContract = new ethers.Contract(
            '0xcA11bde05977b3631167028862bE2a173976CA11',
            this.getAbi('multicall'),
            signer
        );
        return multicallContract;
    }

    private async getAaveV3Pool(chainName: string): Promise<ethers.Contract> {
        const signer = await this.chainService.getSigner(chainName);
        const config = this.chainService.getChainConfig(chainName);
        const contract = new ethers.Contract(
            config.contracts.aavev3Pool,
            this.getAbi('aaveV3Pool'),
            signer
        );
        return contract;
    }

    private async getFlashLoanLiquidation(chainName: string): Promise<ethers.Contract> {
        const signer = await this.chainService.getSigner(chainName);
        const config = this.chainService.getChainConfig(chainName);
        const flashLoanLiquidation = new ethers.Contract(
            config.contracts.flashLoanLiquidation,
            this.getAbi('flashLoanLiquidation'),
            signer
        );
        return flashLoanLiquidation;
    }

    private async getDataProvider(chainName: string): Promise<ethers.Contract> {
        const signer = await this.chainService.getSigner(chainName);
        // aaveV3Pool
        const aaveV3Pool = await this.getAaveV3Pool(chainName);
        // addressesProvider
        let addressesProviderAddress = this.configService.get(`${chainName}-addressesProviderAddress`);
        if (!addressesProviderAddress) {
            addressesProviderAddress = await aaveV3Pool.ADDRESSES_PROVIDER();
            this.configService.set(`${chainName}-addressesProviderAddress`, addressesProviderAddress);
            this.logger.log(`[${chainName}] set addressesProviderAddress: ${addressesProviderAddress}`);
        }
        const addressesProvider = new ethers.Contract(
            addressesProviderAddress,
            this.getAbi('addressesProvider'),
            signer
        );

        // dataProvider
        let dataProviderAddress = this.configService.get(`${chainName}-dataProviderAddress`);
        if (!dataProviderAddress) {
            dataProviderAddress = await addressesProvider.getPoolDataProvider();
            this.configService.set(`${chainName}-dataProviderAddress`, dataProviderAddress);
            this.logger.log(`[${chainName}] set dataProviderAddress: ${dataProviderAddress}`);
        }
        const dataProvider = new ethers.Contract(
            dataProviderAddress,
            this.getAbi('dataProvider'),
            signer
        );
        return dataProvider;
    }

    private async getPriceOracle(chainName: string): Promise<ethers.Contract> {
        const signer = await this.chainService.getSigner(chainName);
        // aaveV3Pool
        const aaveV3Pool = await this.getAaveV3Pool(chainName);
        // addressesProvider
        let addressesProviderAddress = this.configService.get(`${chainName}-addressesProviderAddress`);
        if (!addressesProviderAddress) {
            addressesProviderAddress = await aaveV3Pool.ADDRESSES_PROVIDER();
            this.configService.set(`${chainName}-addressesProviderAddress`, addressesProviderAddress);
            this.logger.log(`[${chainName}] set addressesProviderAddress: ${addressesProviderAddress}`);
        }
        const addressesProvider = new ethers.Contract(
            addressesProviderAddress,
            this.getAbi('addressesProvider'),
            signer
        );

        // priceOracle
        let priceOracleAddress = this.configService.get(`${chainName}-priceOracleAddress`);
        if (!priceOracleAddress) {
            priceOracleAddress = await addressesProvider.getPriceOracle();
            this.configService.set(`${chainName}-priceOracleAddress`, priceOracleAddress);
            this.logger.log(`[${chainName}] set priceOracleAddress: ${priceOracleAddress}`);
        }
        const priceOracle = new ethers.Contract(
            priceOracleAddress,
            this.getAbi('priceOracle'),
            signer
        );
        return priceOracle;
    }


    private async loadTokenCache() {
        try {
            const tokens = await this.databaseService.getAllTokens();
            for (const token of tokens) {
                if (!this.tokenCache.has(token.chainName)) {
                    this.tokenCache.set(token.chainName, new Map());
                }
                const chainTokens = this.tokenCache.get(token.chainName)!;
                chainTokens.set(token.address.toLowerCase(), {
                    symbol: token.symbol,
                    decimals: token.decimals,
                });
            }
            this.logger.log(`Loaded ${tokens.length} tokens into cache`);
        } catch (error) {
            this.logger.error(`Error loading token cache: ${error.message}`);
        }
    }

    private async getTokenInfo(chainName: string, address: string, provider?: ethers.Provider): Promise<TokenInfo> {
        const normalizedAddress = address.toLowerCase();

        // 如果没有传入 provider，则从缓存获取
        const providerToUse = provider || await this.chainService.getProvider(chainName);
        if (!providerToUse) {
            throw new Error(`Provider not initialized for chain ${chainName}`);
        }

        // 检查缓存
        const chainTokens = this.tokenCache.get(chainName);
        if (chainTokens?.has(normalizedAddress)) {
            return chainTokens.get(normalizedAddress)!;
        }

        // 检查数据库
        const dbToken = await this.databaseService.getToken(chainName, normalizedAddress);
        if (dbToken) {
            // 更新缓存
            if (!this.tokenCache.has(chainName)) {
                this.tokenCache.set(chainName, new Map());
            }
            const tokenInfo = {
                symbol: dbToken.symbol,
                decimals: dbToken.decimals,
            };
            this.tokenCache.get(chainName)!.set(normalizedAddress, tokenInfo);
            return tokenInfo;
        }

        // 查询链上数据
        try {
            const erc20Abi = [
                'function symbol() view returns (string)',
                'function decimals() view returns (uint8)',
            ];
            const contract = new ethers.Contract(normalizedAddress, erc20Abi, providerToUse);
            const [symbol, decimals] = await Promise.all([
                contract.symbol(),
                contract.decimals(),
            ]);

            // 保存到数据库和缓存
            await this.databaseService.saveToken(chainName, normalizedAddress, symbol, Number(decimals));
            if (!this.tokenCache.has(chainName)) {
                this.tokenCache.set(chainName, new Map());
            }
            const tokenInfo = { symbol, decimals: Number(decimals) };
            this.tokenCache.get(chainName)!.set(normalizedAddress, tokenInfo);
            return tokenInfo;
        } catch (error) {
            this.logger.error(`Error getting token info for ${normalizedAddress} on ${chainName}: ${error.message}`);
            // 如果查询失败，返回默认值
            return { symbol: 'UNKNOWN', decimals: 18 };
        }
    }

    private formatAmount(amount: bigint, decimals: number): string {
        return Number(ethers.formatUnits(amount, decimals)).toFixed(6);
    }

    private amountToUSD(amount: bigint, decimals: number, price: bigint): number {
        return Number(ethers.formatUnits(amount, decimals)) * Number(price) / 1e8;
    }

    private USDToAmount(usd: number, decimals: number, price: bigint): number {
        return Number((usd * 1e8 / Number(price) * (10 ** decimals)).toFixed(0));
    }

    private async loadActiveLoans() {
        try {
            const chains = this.chainService.getActiveChains();
            for (const chainName of chains) {
                const activeLoans = await this.databaseService.getActiveLoans(chainName);
                this.logger.log(`[${chainName}] Found ${activeLoans.length} active loans in database`);

                if (activeLoans.length > 0) {
                    this.logger.log(`[${chainName}] Loading active loans into memory...`);

                    // 初始化内存中的活跃贷款集合
                    if (!this.activeLoans.has(chainName)) {
                        this.activeLoans.set(chainName, new Map());
                    }
                    const activeLoansMap = this.activeLoans.get(chainName);

                    // 将数据库中的活跃贷款加载到内存，设置 nextCheckTime 为当前时间
                    for (const loan of activeLoans) {
                        activeLoansMap.set(loan.user, {
                            nextCheckTime: loan.nextCheckTime,
                            healthFactor: loan.healthFactor
                        });
                    }

                    this.logger.log(`[${chainName}] Loaded ${activeLoansMap.size} active loans into memory, will check immediately`);
                }
            }
        } catch (error) {
            this.logger.error(`Error loading active loans: ${error.message}`);
        }
    }

    private createBorrowEventHandler(chainName: string, provider: ethers.Provider) {
        return async (reserve: string, user: string, onBehalfOf: string, amount: bigint,
            interestRateMode: number, borrowRate: bigint, referralCode: number, event: any) => {
            try {
                const tokenInfo = await this.getTokenInfo(chainName, reserve, provider);
                this.logger.log(`[${chainName}] 🩷 Borrow event detected:`);
                this.logger.log(`- Reserve: ${reserve} (${tokenInfo.symbol})`);
                this.logger.log(`- User: ${user}`);
                this.logger.log(`- OnBehalfOf: ${onBehalfOf}`);
                this.logger.log(`- Amount: ${this.formatAmount(amount, tokenInfo.decimals)} ${tokenInfo.symbol}`);
                this.logger.log(`- Interest Rate Mode: ${interestRateMode}`);
                this.logger.log(`- Borrow Rate: ${ethers.formatUnits(borrowRate, 27)}`);
                this.logger.log(`- Referral Code: ${referralCode}`);
                this.logger.log(`- Transaction Hash: ${event?.transactionHash || event?.log?.transactionHash}`);

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
            } catch (error) {
                this.logger.error(`[${chainName}] Error processing Borrow event: ${error.message}`);
            }
        };
    }

    private createLiquidationCallEventHandler(chainName: string, provider: ethers.Provider) {
        return async (collateralAsset: string, debtAsset: string, user: string,
            debtToCover: bigint, liquidatedCollateralAmount: bigint, liquidator: string,
            receiveAToken: boolean, event: any) => {
            try {
                user = user.toLowerCase();
                const priceOracle = await this.getPriceOracle(chainName);
                const [collateralInfo, debtInfo, collateralPrice, debtPrice] = await Promise.all([
                    this.getTokenInfo(chainName, collateralAsset, provider),
                    this.getTokenInfo(chainName, debtAsset, provider),
                    priceOracle.getAssetPrice(collateralAsset),
                    priceOracle.getAssetPrice(debtAsset)
                ]);
                this.logger.log(`[${chainName}] 😄 LiquidationCall event detected:`);
                this.logger.log(`- Collateral Asset: ${collateralAsset} (${collateralInfo.symbol})`);
                this.logger.log(`- Debt Asset: ${debtAsset} (${debtInfo.symbol})`);
                this.logger.log(`- User: ${user}`);
                this.logger.log(`- Debt to Cover: ${this.formatAmount(debtToCover, debtInfo.decimals)} ${debtInfo.symbol} = ${this.amountToUSD(debtToCover, debtInfo.decimals, debtPrice)} USD`);
                this.logger.log(`- Liquidated Amount: ${this.formatAmount(liquidatedCollateralAmount, collateralInfo.decimals)} ${collateralInfo.symbol} = ${this.amountToUSD(liquidatedCollateralAmount, collateralInfo.decimals, collateralPrice)} USD`);
                this.logger.log(`- Liquidator: ${liquidator}`);
                this.logger.log(`- Receive AToken: ${receiveAToken}`);
                this.logger.log(`- Transaction Hash: ${event?.transactionHash || event?.log?.transactionHash}`);

                const activeLoansMap = this.activeLoans.get(chainName);
                if (activeLoansMap && activeLoansMap.get(user)?.healthFactor > 0) {
                    activeLoansMap.delete(user);
                    await this.databaseService.recordLiquidation(
                        chainName,
                        user,
                        liquidator,
                        event?.transactionHash || event?.log?.transactionHash
                    );
                    this.logger.log(`[${chainName}] Recorded liquidation for user ${user}`);
                } else {
                    this.logger.log(`[${chainName}] No active loan found for user ${user}, skipping liquidation record`);
                }
            } catch (error) {
                this.logger.error(`[${chainName}] Error processing LiquidationCall event: ${error.message}`);
            }
        };
    }

    private async setupEventListeners(chainName: string, contract: ethers.Contract, provider: ethers.Provider) {
        // 移除旧的事件监听器
        contract.removeAllListeners('Borrow');
        contract.removeAllListeners('LiquidationCall');

        // 添加新的事件监听器
        contract.on('Borrow', this.createBorrowEventHandler(chainName, provider));
        contract.on('LiquidationCall', this.createLiquidationCallEventHandler(chainName, provider));
    }

    private async startListening() {
        const chains = this.chainService.getActiveChains();
        this.logger.log(`Starting to listen on chains: ${chains.join(', ')}`);

        // 并发执行所有链的初始化
        await Promise.all(chains.map(async (chainName) => {
            try {
                const provider = await this.chainService.getProvider(chainName);
                const config = this.chainService.getChainConfig(chainName);

                // 获取当前区块高度
                const currentBlock = await provider.getBlockNumber();
                this.logger.log(`[${chainName}] Current block number: ${currentBlock}`);

                // 检查合约代码是否存在
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

                // 初始化事件监听器
                await this.setupEventListeners(chainName, aaveV3Pool, provider);

                this.logger.log(`[${chainName}] Successfully set up event listeners and verified contract connection`);
            } catch (error) {
                this.logger.error(`Failed to set up event listeners for ${chainName}: ${error.message}`);
            }
        }));
    }

    private startHealthFactorChecker() {
        let isChecking = false;

        const checkAllLoans = async () => {
            if (isChecking) {
                return;
            }

            isChecking = true;
            try {
                // 并发执行所有链的检查
                const chains = Array.from(this.activeLoans.keys());
                await Promise.all(chains.map(chainName => this.checkHealthFactorsBatch(chainName)));
            } catch (error) {
                this.logger.error(`Error in health factor checker: ${error.message}`);
            } finally {
                isChecking = false;
                // 在完成检查后调度下一次检查
                setTimeout(checkAllLoans, this.MIN_WAIT_TIME);
            }
        };

        // 立即执行一次检查
        checkAllLoans();
    }

    private formatDate(date: Date): string {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        const seconds = String(date.getSeconds()).padStart(2, '0');
        const milliseconds = String(date.getMilliseconds()).padStart(3, '0');
        return `${year}/${month}/${day} ${hours}:${minutes}:${seconds}.${milliseconds}`;
    }

    private async checkHealthFactorsBatch(chainName: string) {
        try {
            const activeLoansMap = this.activeLoans.get(chainName);
            if (!activeLoansMap || activeLoansMap.size === 0) return;

            const usersToCheck = Array.from(activeLoansMap.entries())
                .filter(([_, info]) => info.nextCheckTime <= new Date())
                .map(([user]) => user);

            if (usersToCheck.length === 0) return;

            // 将用户分批处理，每批最多 100 个
            const BATCH_SIZE = 100;
            const batches = [];
            for (let i = 0; i < usersToCheck.length; i += BATCH_SIZE) {
                const batchUsers = usersToCheck.slice(i, i + BATCH_SIZE);
                batches.push(batchUsers);
            }

            this.logger.log(`[${chainName}] Processing ${batches.length} batches concurrently...`);

            // 并发处理所有批次
            await Promise.all(batches.map(async (batchUsers, batchIndex) => {
                try {
                    this.logger.log(`[${chainName}] Processing batch ${batchIndex + 1}/${batches.length} (${batchUsers.length} users)...`);
                    await Promise.race([
                        this.processBatch(chainName, batchUsers, activeLoansMap),
                        new Promise((_, reject) =>
                            setTimeout(() => reject(new Error(`Batch check timeout after ${this.BATCH_CHECK_TIMEOUT}ms`)),
                                this.BATCH_CHECK_TIMEOUT)
                        )
                    ]);
                } catch (error) {
                    this.logger.error(`[${chainName}] Error processing batch ${batchIndex + 1}/${batches.length}: ${error.message}`);
                }
            }));

            this.logger.log(`[${chainName}] Completed processing all ${batches.length} batches`);
        } catch (error) {
            this.logger.error(`[${chainName}] Error checking health factors batch: ${error.message}`);
        }
    }

    private async processBatch(
        chainName: string,
        batchUsers: string[],
        activeLoansMap: Map<string, LoanInfo>,
    ) {
        const aaveV3Pool = await this.getAaveV3Pool(chainName);
        const accountDataMap = await this.getUserAccountDataBatch(chainName, aaveV3Pool, batchUsers);

        // 并发处理每个用户
        await Promise.all(batchUsers.map(async (user) => {
            const accountData = accountDataMap.get(user);
            if (!accountData) return;

            await this.processUser(
                chainName,
                user,
                accountData,
                activeLoansMap,
                aaveV3Pool
            );
        }));
    }

    private async processUser(
        chainName: string,
        user: string,
        accountData: UserAccountData,
        activeLoansMap: Map<string, LoanInfo>,
        aaveV3Pool: ethers.Contract
    ) {
        const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
        const totalDebt = Number(ethers.formatUnits(accountData.totalDebtBase, 8));

        // 如果总债务小于 minDebtUSD, 则从内存和数据库中移除该用户
        if (totalDebt < this.chainService.getChainConfig(chainName).minDebtUSD) {
            this.activeLoans.get(chainName)?.delete(user);
            this.liquidationInfoCache.delete(`${chainName}-${user}`); // 清理清算记录
            await this.databaseService.deactivateLoan(chainName, user);
            this.logger.log(`[${chainName}] Removed user ${user} as total debt ${totalDebt} < ${this.chainService.getChainConfig(chainName).minDebtUSD} USD`);
            return;
        }

        const waitTime = this.calculateWaitTime(chainName, healthFactor);
        const nextCheckTime = new Date(Date.now() + waitTime);
        const formattedDate = this.formatDate(nextCheckTime);

        // 如果健康因子低于清算阈值，尝试清算
        if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
            const cacheKey = `${chainName}-${user}`;
            const cachedInfo = this.liquidationInfoCache.get(cacheKey);
            this.logger.log(`[${chainName}] ${user} totalDebt: ${totalDebt} USD, healthFactor: ${healthFactor}`);

            this.logger.log(`[${chainName}] Liquidation threshold ${healthFactor} <= ${this.LIQUIDATION_THRESHOLD} reached for user ${user}, attempting liquidation ${cachedInfo?.retryCount}`);
            await this.executeLiquidation(chainName, user, healthFactor, aaveV3Pool);
            this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms (at ${formattedDate}), healthFactor: ${healthFactor}`);
        } else {
            if (healthFactor < this.LIQUIDATION_THRESHOLD_2) {
                this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms, healthFactor: ${healthFactor.toFixed(8)} (${totalDebt.toFixed(2)} USD)`);
            }
            // 如果健康因子高于清算阈值，清除清算记录
            this.liquidationInfoCache.delete(`${chainName}-${user}`);
        }

        // 更新内存中的健康因子和下次检查时间
        activeLoansMap.set(user, {
            nextCheckTime: nextCheckTime,
            healthFactor: healthFactor
        });

        // 更新数据库中的健康因子和下次检查时间
        await this.databaseService.updateLoanHealthFactor(
            chainName,
            user,
            healthFactor,
            nextCheckTime
        );
    }

    private async getUserAccountDataBatch(chainName: string, contract: ethers.Contract, users: string[]): Promise<Map<string, UserAccountData>> {
        try {
            const multicallContract = await this.getMulticall(chainName);

            // 准备调用数据
            const calls = users.map(user => ({
                target: contract.target,
                callData: contract.interface.encodeFunctionData('getUserAccountData', [user])
            }));

            // 执行批量调用
            const [, returnData] = await multicallContract.aggregate.staticCall(calls);

            // 解析返回数据
            const results = new Map<string, UserAccountData>();
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
        } catch (error) {
            this.logger.error(`Error getting user account data batch: ${error.message}`);
            return new Map();
        }
    }


    private calculateHealthFactor(healthFactor: bigint): number {
        // 将 bigint 转换为 number，并除以 1e18 得到实际值
        return Number(healthFactor) / 1e18;
    }

    private calculateWaitTime(chainName: string, healthFactor: number): number {
        const c1 = this.chainService.getChainConfig(chainName).minWaitTime; // Minimum wait time
        const c2 = this.MAX_WAIT_TIME; // Maximum wait time
        const c3 = this.chainService.getChainConfig(chainName).blockTime;

        if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
            return c1;
        }
        if (healthFactor <= this.LIQUIDATION_THRESHOLD_2) {
            return c3;
        } else if (healthFactor <= this.HEALTH_FACTOR_THRESHOLD) {
            // 在 LIQUIDATION_THRESHOLD_2 和 HEALTH_FACTOR_THRESHOLD 之间线性增长
            const ratio = (healthFactor - this.LIQUIDATION_THRESHOLD_2) / (this.HEALTH_FACTOR_THRESHOLD - this.LIQUIDATION_THRESHOLD_2);
            return Math.floor(c3 + (c2 - c3) * ratio);
        } else {
            return c2;
        }
    }

    private async getLiquidationInfo(chainName: string, user: string, healthFactor: number, aaveV3Pool: ethers.Contract): Promise<LiquidationInfo | null> {
        const cacheKey = `${chainName}-${user}`;
        const cachedInfo = this.liquidationInfoCache.get(cacheKey);

        if (cachedInfo && Date.now() - cachedInfo.timestamp < this.CACHE_TTL) {
            this.logger.log(`[${chainName}] Using cached liquidation info for user ${user}`);
            return cachedInfo;
        }

        try {
            // 1. 使用 multicall 批量获取用户配置和储备资产列表
            const [multicall, dataProvider, priceOracle] = await Promise.all([
                this.getMulticall(chainName),
                this.getDataProvider(chainName),
                this.getPriceOracle(chainName)
            ]);
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

            // 2. 使用 multicall 批量查询所有借贷资产的债务数据
            const reserveCalls = [];
            const borrowingAssets = [];
            const collateralAssets = [];

            // 优化：使用单个循环处理所有资产
            for (let i = 0; i < reservesList.length; i++) {
                const asset = reservesList[i];
                // return (self.data >> (reserveIndex << 1)) & 1 != 0;
                const isBorrowing = (BigInt(userConfig.data) >> (BigInt(i) << BigInt(1))) !== BigInt(0);
                // return (self.data >> ((reserveIndex << 1) + 1)) & 1 != 0;
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
                return null;
            }

            // 执行批量查询
            const [, reserveReturnData] = await multicall.aggregate.staticCall(reserveCalls);

            // 解析返回数据并处理债务和抵押品
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
                    const userReserveData = dataProvider.interface.decodeFunctionResult(
                        'getUserReserveData',
                        reserveReturnData[callIndex]
                    );
                    callIndex++;

                    // 处理债务数据
                    if (isBorrowing) {
                        const currentStableDebt = BigInt(userReserveData.currentStableDebt);
                        const currentVariableDebt = BigInt(userReserveData.currentVariableDebt);
                        const totalDebt = currentStableDebt + currentVariableDebt;
                        if (totalDebt > maxDebtAmount) {
                            maxDebtAmount = totalDebt;
                            maxDebtAsset = asset;
                        }
                    }

                    // 处理抵押品数据
                    if (isUsingAsCollateral) {
                        const collateralAmount = BigInt(userReserveData.currentATokenBalance);
                        this.logger.log(`[${chainName}] ${user} Collateral: ${asset}, collateralAmount: ${collateralAmount}`);
                        if (collateralAmount > maxCollateralAmount) {
                            maxCollateralAmount = collateralAmount;
                            maxCollateralAsset = asset;
                        }
                    }
                }
            }

            if (maxDebtAmount === BigInt(0) || maxCollateralAmount === BigInt(0)) {
                this.logger.log(`[${chainName}] No debt or collateral found for user ${user}, maxDebtAmount: ${maxDebtAmount}, maxCollateralAmount: ${maxCollateralAmount}`);
                return null;
            }

            const [collateralTokenInfo, debtTokenInfo, debtPrice, collateralPrice] = await Promise.all([
                this.getTokenInfo(chainName, maxCollateralAsset),
                this.getTokenInfo(chainName, maxDebtAsset),
                priceOracle.getAssetPrice(maxDebtAsset),
                priceOracle.getAssetPrice(maxCollateralAsset)
            ]);

            const liquidationInfo: LiquidationInfo = {
                maxDebtAsset,
                maxDebtAmount,
                maxCollateralAsset,
                maxCollateralAmount,
                collateralTokenInfo,
                debtTokenInfo,
                debtPrice,
                collateralPrice,
                user,
                healthFactor,
                timestamp: Date.now(),
                retryCount: (cachedInfo?.retryCount || 0) + 1,
                data: "0x"
            };

            this.liquidationInfoCache.set(cacheKey, liquidationInfo);
            return liquidationInfo;
        } catch (error) {
            this.logger.error(`[${chainName}] Error getting liquidation info for user ${user}: ${error.message}`);
            return null;
        }
    }

    private async executeLiquidation(chainName: string, user: string, healthFactor: number, aaveV3Pool: ethers.Contract) {
        try {
            const liquidationInfo = await this.getLiquidationInfo(chainName, user, healthFactor, aaveV3Pool);
            if (!liquidationInfo) {
                return;
            }

            const { maxDebtAsset, maxDebtAmount, maxCollateralAsset, maxCollateralAmount, collateralTokenInfo, debtTokenInfo, collateralPrice, debtPrice } = liquidationInfo;
            const collateralFormatAmount = this.formatAmount(maxCollateralAmount, collateralTokenInfo.decimals);
            const collateralUSDAmount = this.amountToUSD(maxCollateralAmount, collateralTokenInfo.decimals, collateralPrice);
            const debtFormatAmount = this.formatAmount(maxDebtAmount, debtTokenInfo.decimals);
            const debtUSDAmount = this.amountToUSD(maxDebtAmount, debtTokenInfo.decimals, debtPrice);
            let debtToCover = MaxUint256 // 可以全额清算时必须全额清算，否则 103 错误
            let debtToCoverUSD = debtUSDAmount * (1 + 0.05);// 5% 奖励; 足额清算有奖励;
            if (debtUSDAmount > collateralUSDAmount) {
                debtToCoverUSD = collateralUSDAmount * (1 - 1 / 1000);
                debtToCover = BigInt(this.USDToAmount(debtToCoverUSD, debtTokenInfo.decimals, debtPrice));
                this.logger.log(`[${chainName}] partial liquidation, debtToCoverUSD: ${debtToCoverUSD}, collateralUSDAmount: ${collateralUSDAmount}`);
                // TODO: 优化：如果没有完全清算债务 debt，那合约中也没必要借入完整的 debt，可以借入部分 debt 以节省利息
            }
            // 如果总债务小于 minDebtUSD, 则从内存和数据库中移除该用户
            if (debtToCoverUSD < this.chainService.getChainConfig(chainName).minDebtUSD) {
                this.activeLoans.get(chainName)?.delete(user);
                this.liquidationInfoCache.delete(`${chainName}-${user}`); // 清理清算记录
                await this.databaseService.deactivateLoan(chainName, user);
                this.logger.log(`[${chainName}] Removed user ${user} as debtToCoverUSD: ${debtToCoverUSD} < ${this.chainService.getChainConfig(chainName).minDebtUSD} USD`);
                return;
            }

            this.logger.log(`[${chainName}] 💰 Executing flash loan liquidation with aggregator:`);
            this.logger.log(`- User: ${user}`);
            this.logger.log(`- Health Factor: ${healthFactor}`);
            this.logger.log(`- Collateral Asset: ${maxCollateralAsset} (${maxCollateralAmount} = ${collateralFormatAmount} ${collateralTokenInfo.symbol} ≈ ${collateralUSDAmount.toFixed(2)} USD)`);
            this.logger.log(`- Debt Asset: ${maxDebtAsset} (${maxDebtAmount} = ${debtFormatAmount} ${debtTokenInfo.symbol} ≈ ${debtUSDAmount.toFixed(2)} USD)`);
            this.logger.log(`- Debt To Cover: ${debtToCover} = ${this.formatAmount(debtToCover, debtTokenInfo.decimals)} ${debtTokenInfo.symbol} ≈ ${debtToCoverUSD.toFixed(2)} USD`);
            this.logger.log(`- Price Rate: 1 ${debtTokenInfo.symbol} = ${Number(debtPrice) / Number(collateralPrice)} ${collateralTokenInfo.symbol}`);


            // 使用 aggregator 清算, 如果失败则使用 UniswapV3 清算
            let data = "0x";
            try {
                if (liquidationInfo.retryCount >= 1) {
                    if (liquidationInfo.data === "0x") {
                        data = await this.getAggregatorData(chainName, liquidationInfo, debtToCoverUSD);
                        liquidationInfo.data = data; // add to cache
                    } else {
                        data = liquidationInfo.data;
                    }
                    this.logger.log(`[${chainName}] Use flash loan liquidation with aggregator, data: ${liquidationInfo.data}`);
                } else {
                    this.logger.log(`[${chainName}] Use flash loan liquidation with UniswapV3, data: ${liquidationInfo.data}`);
                }
            } catch (error) {
                this.logger.log(`[${chainName}] Use flash loan liquidation with UniswapV3, data: ${liquidationInfo.data}`);
            }

            try {
                // 执行闪电贷清算
                const flashLoanLiquidation = await this.getFlashLoanLiquidation(chainName);
                // 获取当前 gas 价格并提高 50%
                const gasPrice = await flashLoanLiquidation.runner.provider.getFeeData();
                const maxPriorityFeePerGas = gasPrice.maxPriorityFeePerGas ? gasPrice.maxPriorityFeePerGas * BigInt(15) / BigInt(10) : ethers.parseUnits('1', 'gwei');
                const maxFeePerGas = gasPrice.maxFeePerGas ? gasPrice.maxFeePerGas + (maxPriorityFeePerGas || BigInt(0)) : undefined;
                this.logger.log(`[${chainName}] gasPrice: ${gasPrice.gasPrice}, maxFeePerGas: ${maxFeePerGas}, maxPriorityFeePerGas: ${maxPriorityFeePerGas} `);

                let gasLimit = 3000000;
                if (healthFactor > 1) {
                    gasLimit = 0;
                }
                const tx = await flashLoanLiquidation.executeLiquidation(
                    maxCollateralAsset,
                    maxDebtAsset,
                    user,
                    debtToCover,
                    data,
                    {
                        maxFeePerGas,
                        maxPriorityFeePerGas,
                        gasLimit: gasLimit > 0 ? gasLimit : undefined
                    }
                );

                this.logger.log(`[${chainName}] Flash loan liquidation executed successfully, tx: ${tx.hash} gasLimit: ${tx.gasLimit}`);
                await tx.wait();

                // 清算成功后清除缓存
                this.liquidationInfoCache.delete(`${chainName} -${user} `);
            } catch (error) {
                this.logger.error(`[${chainName}] Error executing flash loan liquidation for user ${user}: ${error.message} `);
                // 清算失败时保留缓存，下次可以直接使用
            }
        } catch (error) {
            this.logger.error(`[${chainName}] Error executing liquidation for user ${user}: ${error.message} `);
        }
    }

    // now only support odos, usdc = 5% 
    private async getAggregatorData(chainName: string, liquidationInfo: LiquidationInfo, debtToCoverUSD: number): Promise<string> {
        const { collateralPrice, collateralTokenInfo } = liquidationInfo;
        const collateralAmount = this.USDToAmount(debtToCoverUSD, collateralTokenInfo.decimals, collateralPrice);
        this.logger.log(`[${chainName}] collateralUSDAmount: ${debtToCoverUSD}, collateralAmount: ${collateralAmount} `);

        let inputAmount = 0;
        let outputTokens = [];
        if (liquidationInfo.maxCollateralAsset === this.chainService.getChainConfig(chainName).contracts.usdc) {
            inputAmount = collateralAmount * 0.958;
            outputTokens = [
                {
                    "tokenAddress": liquidationInfo.maxDebtAsset,
                    "proportion": "1"
                }
            ]
        } else if (liquidationInfo.maxDebtAsset === this.chainService.getChainConfig(chainName).contracts.usdc) {
            inputAmount = collateralAmount * 0.992;
            outputTokens = [
                {
                    "tokenAddress": liquidationInfo.maxDebtAsset,
                    "proportion": "1"
                }
            ]
        } else {
            inputAmount = collateralAmount;
            outputTokens = [
                {
                    "tokenAddress": liquidationInfo.maxDebtAsset,
                    "proportion": "0.95"
                },
                {
                    "tokenAddress": this.chainService.getChainConfig(chainName).contracts.usdc,
                    "proportion": "0.05"
                }
            ]
        }

        const postData = {
            "chainId": this.chainService.getChainConfig(chainName).chainId,
            "inputTokens": [
                {
                    "tokenAddress": liquidationInfo.maxCollateralAsset,
                    "amount": inputAmount.toFixed(0).toString()
                }
            ],
            "outputTokens": outputTokens,
            "userAddr": this.chainService.getChainConfig(chainName).contracts.flashLoanLiquidation,
            "slippageLimitPercent": "3",
            "pathViz": "false",
            "pathVizImage": "false"
        };
        this.logger.log(`[${chainName}] postData: ${JSON.stringify(postData, null, 2)} `);

        try {
            const response = await axios.post('https://api.odos.xyz/sor/quote/v2', postData, {
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                }
            });

            // 返回 pathId
            if (response.data.pathId) {
                return await this.getPathData(this.chainService.getChainConfig(chainName).contracts.flashLoanLiquidation, this.chainService.getChainConfig(chainName).contracts.usdc, response.data.pathId);;
            } else {
                this.logger.error(`[${chainName}] No pathId in response`);
                return "0x";
            }
        } catch (error) {
            this.logger.error(`[${chainName}]Error in getAggregatorData: ${error.message} `);
            if (error.response) {
                this.logger.error(`[${chainName}] Error response data: ${JSON.stringify(error.response.data, null, 2)} `);
            }
            return "0x";
        }
    }

    private async getPathData(flashLoanLiquidation: string, usdc: string, pathId: string): Promise<string> {
        let data = JSON.stringify({
            "userAddr": flashLoanLiquidation,
            "pathId": pathId,
            "simulate": "false",
        });

        try {
            const response = await axios.post('https://api.odos.xyz/sor/assemble', data, {
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json',
                }
            });

            // this.logger.log(`[${flashLoanLiquidation}] output: ${JSON.stringify(response.data.outputTokens, null, 2)} `);
            // this.logger.log(`[response: ${JSON.stringify(response.data, null, 2)} `);

            // 返回 pathId
            if (response.data.transaction) {
                return ethers.AbiCoder.defaultAbiCoder().encode(
                    ['address', 'address', 'bytes'],
                    [usdc, response.data.transaction.to, response.data.transaction.data]
                );
            } else {
                this.logger.error(`${flashLoanLiquidation} No pathId in response`);
                return "0x";
            }
        } catch (error) {
            this.logger.error(`${flashLoanLiquidation} Error in getPathData: ${error.message} `);
            return "0x";
        }
    }
}
