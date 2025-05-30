import { Injectable, Logger, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { ChainService } from '../chain/chain.service';
import { ethers, MaxUint256 } from 'ethers';
import { ConfigService } from '@nestjs/config';
import { DatabaseService } from '../database/database.service';
import * as fs from 'fs';
import * as path from 'path';
import { WebSocket } from 'ws';

interface UserAccountData {
    totalCollateralBase: bigint;
    totalDebtBase: bigint;
    availableBorrowsBase: bigint;
    currentLiquidationThreshold: bigint;
    ltv: bigint;
    healthFactor: bigint;
}

interface TokenInfo {
    symbol: string;
    decimals: number;
}

interface LoanInfo {
    nextCheckTime: Date;
    healthFactor: number;
}

@Injectable()
export class BorrowDiscoveryService implements OnModuleInit, OnModuleDestroy {
    private readonly logger = new Logger(BorrowDiscoveryService.name);
    private activeLoans: Map<string, Map<string, LoanInfo>> = new Map();
    private tokenCache: Map<string, Map<string, TokenInfo>> = new Map();
    private lastLiquidationAttempt: Map<string, Map<string, number>> = new Map(); // 新增：记录上次清算尝试的健康因子
    private readonly LIQUIDATION_THRESHOLD = 1.00001; // 清算阈值
    private readonly CRITICAL_THRESHOLD = 1.001; // 危险阈值
    private readonly HEALTH_FACTOR_THRESHOLD = 1.1; // 健康阈值
    private readonly MIN_WAIT_TIME: number; // 最小等待时间（毫秒）
    private readonly MAX_WAIT_TIME: number; // 最大等待时间（毫秒）
    private readonly PRIVATE_KEY: string; // EOA 私钥
    private checkInterval: NodeJS.Timeout;
    private aaveV3PoolCache: Map<string, ethers.Contract> = new Map(); // 合约缓存
    private providerCache: Map<string, ethers.Provider> = new Map(); // Provider 缓存
    private signerCache: Map<string, ethers.Signer> = new Map(); // Signer 缓存
    private dataProviderCache: Map<string, ethers.Contract> = new Map(); // DataProvider 缓存
    private priceOracleCache: Map<string, ethers.Contract> = new Map(); // PriceOracle 缓存
    private flashLoanLiquidationCache: Map<string, ethers.Contract> = new Map(); // FlashLoanLiquidation 合约缓存
    private abiCache: Map<string, any> = new Map(); // 新增：ABI 缓存
    private multicallCache: Map<string, ethers.Contract> = new Map(); // 新增：multicall 合约缓存

    constructor(
        private readonly chainService: ChainService,
        private readonly configService: ConfigService,
        private readonly databaseService: DatabaseService,
    ) {
        // 从环境变量读取配置，默认值：最小1s，最大4小时
        this.MIN_WAIT_TIME = this.configService.get<number>('MIN_CHECK_INTERVAL', 1000);
        this.MAX_WAIT_TIME = this.configService.get<number>('MAX_CHECK_INTERVAL', 4 * 60 * 60 * 1000);
        this.PRIVATE_KEY = this.configService.get<string>('PRIVATE_KEY');
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

    async onModuleInit() {
        this.logger.log('BorrowDiscoveryService initializing...');
        await this.initializeAbis(); // 新增：初始化 ABI
        await this.initializeResources();
        await this.loadTokenCache();
        await this.loadActiveLoans();
        await this.startListening();
        this.startHealthFactorChecker();
    }

    private async initializeResources() {
        const chains = this.chainService.getActiveChains();
        for (const chainName of chains) {
            try {
                // 初始化 provider
                const provider = await this.chainService.getProvider(chainName);
                this.providerCache.set(chainName, provider);

                // 初始化 signer
                const signer = new ethers.Wallet(this.PRIVATE_KEY, provider);
                this.signerCache.set(chainName, signer);
                this.logger.log(`[${chainName}] Initialized signer: ${signer.address}`);

                // 初始化 multicall 合约
                const multicallContract = new ethers.Contract(
                    '0xcA11bde05977b3631167028862bE2a173976CA11',
                    this.getAbi('multicall'),
                    signer
                );
                this.multicallCache.set(chainName, multicallContract);

                // 初始化合约
                const config = this.chainService.getChainConfig(chainName);
                const contract = new ethers.Contract(
                    config.contracts.aavev3Pool,
                    this.getAbi('aaveV3Pool'),
                    signer
                );
                this.aaveV3PoolCache.set(chainName, contract);

                // 初始化 FlashLoanLiquidation 合约
                const flashLoanLiquidation = new ethers.Contract(
                    config.contracts.flashLoanLiquidation,
                    this.getAbi('flashLoanLiquidation'),
                    signer
                );
                this.flashLoanLiquidationCache.set(chainName, flashLoanLiquidation);

                // 初始化 DataProvider
                const addressesProviderAddress = await contract.ADDRESSES_PROVIDER();
                const addressesProvider = new ethers.Contract(
                    addressesProviderAddress,
                    this.getAbi('addressesProvider'),
                    signer
                );

                // dataProvider
                const dataProviderAddress = await addressesProvider.getPoolDataProvider();
                const dataProvider = new ethers.Contract(
                    dataProviderAddress,
                    this.getAbi('dataProvider'),
                    signer
                );
                this.dataProviderCache.set(chainName, dataProvider);

                // priceOracle
                const priceOracleAddress = await addressesProvider.getPriceOracle();
                const priceOracle = new ethers.Contract(
                    priceOracleAddress,
                    this.getAbi('priceOracle'),
                    signer
                );
                this.priceOracleCache.set(chainName, priceOracle);

                this.logger.log(`[${chainName}] Initialized provider, signer, contract, dataProvider and priceOracle`);
            } catch (error) {
                this.logger.error(`[${chainName}] Failed to initialize resources: ${error.message}`);
            }
        }
    }

    private getAbi(name: string): any {
        const abi = this.abiCache.get(name);
        if (!abi) {
            throw new Error(`Abi not initialized for chain ${name}`);
        }
        return abi;
    }

    private getAaveV3Pool(chainName: string): ethers.Contract {
        const contract = this.aaveV3PoolCache.get(chainName);
        if (!contract) {
            throw new Error(`Contract not initialized for chain ${chainName}`);
        }
        return contract;
    }

    private getDataProvider(chainName: string): ethers.Contract {
        const dataProvider = this.dataProviderCache.get(chainName);
        if (!dataProvider) {
            throw new Error(`DataProvider not initialized for chain ${chainName}`);
        }
        return dataProvider;
    }

    private getFlashLoanLiquidation(chainName: string): ethers.Contract {
        const contract = this.flashLoanLiquidationCache.get(chainName);
        if (!contract) {
            throw new Error(`FlashLoanLiquidation contract not initialized for chain ${chainName}`);
        }
        return contract;
    }

    private getMulticall(chainName: string): ethers.Contract {
        const multicall = this.multicallCache.get(chainName);
        if (!multicall) {
            throw new Error(`Multicall contract not initialized for chain ${chainName}`);
        }
        return multicall;
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
        const providerToUse = provider || this.providerCache.get(chainName);
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
                this.logger.log(`[${chainName}] Borrow event detected:`);
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
                this.logger.log(`- Transaction Hash: ${event?.transactionHash || event?.log?.transactionHash}`);

                const activeLoansMap = this.activeLoans.get(chainName);
                if (activeLoansMap && activeLoansMap.has(user)) {
                    activeLoansMap.delete(user);
                    await this.databaseService.recordLiquidation(
                        chainName,
                        user,
                        liquidator,
                        event?.transactionHash || event?.log?.transactionHash
                    );
                    this.logger.log(`[${chainName}] Recorded liquidation for user ${user}`);
                } else {
                    this.logger.log(`[${chainName}] No loan found for user ${user}, skipping liquidation record`);
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

    private async reinitializeEventListeners(chainName: string) {
        try {
            const provider = await this.chainService.getProvider(chainName);
            const contract = this.getAaveV3Pool(chainName);

            this.logger.log(`[${chainName}] Reinitializing event listeners...`);
            await this.setupEventListeners(chainName, contract, provider);
            this.logger.log(`[${chainName}] Event listeners reinitialized successfully`);
        } catch (error) {
            this.logger.error(`[${chainName}] Failed to reinitialize event listeners: ${error.message}`);
        }
    }

    private async startListening() {
        const chains = this.chainService.getActiveChains();
        this.logger.log(`Starting to listen on chains: ${chains.join(', ')}`);

        for (const chainName of chains) {
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
                    continue;
                }
                const code2 = await provider.getCode(config.contracts.flashLoanLiquidation);
                if (code2 === '0x') {
                    this.logger.error(`[${chainName}] No contract code found at address ${config.contracts.flashLoanLiquidation}`);
                    continue;
                }
                this.logger.log(`[${chainName}] Contract code found at ${config.contracts.aavev3Pool}, ${config.contracts.flashLoanLiquidation}`);

                const contract = this.getAaveV3Pool(chainName);

                // 设置WebSocket重连后的回调
                const ws = provider.websocket as WebSocket;
                ws.on('close', async () => {
                    this.logger.warn(`[${chainName}] WebSocket connection closed, will attempt to reinitialize event listeners after reconnection...`);
                });

                ws.on('open', async () => {
                    this.logger.log(`[${chainName}] WebSocket connection reopened, reinitializing event listeners...`);
                    await this.reinitializeEventListeners(chainName);
                });

                // 初始化事件监听器
                await this.setupEventListeners(chainName, contract, provider);

                this.logger.log(`[${chainName}] Successfully set up event listeners and verified contract connection`);
            } catch (error) {
                this.logger.error(`Failed to set up event listeners for ${chainName}: ${error.message}`);
            }
        }
    }

    private startHealthFactorChecker() {
        let isChecking = false;

        const checkAllLoans = async () => {
            if (isChecking) {
                return;
            }

            isChecking = true;
            try {
                // 从内存中获取所有需要检查的贷款
                for (const chainName of this.activeLoans.keys()) {
                    await this.checkHealthFactorsBatch(chainName);
                }
            } catch (error) {
                this.logger.error(`Error in health factor checker: ${error.message}`);
            } finally {
                isChecking = false;
            }
        };

        // 立即执行一次检查
        checkAllLoans();

        // 设置定时器，每最小检查间隔执行一次
        this.checkInterval = setInterval(checkAllLoans, this.MIN_WAIT_TIME);
    }

    private formatDate(date: Date): string {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        const seconds = String(date.getSeconds()).padStart(2, '0');
        return `${year}/${month}/${day} ${hours}:${minutes}:${seconds}`;
    }

    private async checkHealthFactorsBatch(chainName: string) {
        try {
            const activeLoansMap = this.activeLoans.get(chainName);
            if (!activeLoansMap || activeLoansMap.size === 0) return;

            // 初始化链的清算尝试记录
            if (!this.lastLiquidationAttempt.has(chainName)) {
                this.lastLiquidationAttempt.set(chainName, new Map());
            }
            const lastLiquidationMap = this.lastLiquidationAttempt.get(chainName)!;

            const usersToCheck = Array.from(activeLoansMap.entries())
                .filter(([_, info]) => info.nextCheckTime <= new Date())
                .map(([user]) => user);

            if (usersToCheck.length === 0) return;

            // 将用户分批处理，每批最多 100 个
            const BATCH_SIZE = 100;
            for (let i = 0; i < usersToCheck.length; i += BATCH_SIZE) {
                const batchUsers = usersToCheck.slice(i, i + BATCH_SIZE);
                this.logger.log(`[${chainName}] Checking health factors for batch ${i / BATCH_SIZE + 1}/${Math.ceil(usersToCheck.length / BATCH_SIZE)} (${batchUsers.length}/${activeLoansMap.size} users)...`);

                const aaveV3Pool = this.getAaveV3Pool(chainName);
                const accountDataMap = await this.getUserAccountDataBatch(chainName, aaveV3Pool, batchUsers);

                for (const user of batchUsers) {
                    const accountData = accountDataMap.get(user);
                    if (!accountData) continue;

                    const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
                    const totalDebt = Number(ethers.formatUnits(accountData.totalDebtBase, 8));

                    // 如果总债务小于 1 USD，则从内存和数据库中移除该用户
                    if (totalDebt < 1) {
                        activeLoansMap.delete(user);
                        lastLiquidationMap.delete(user); // 清理清算记录
                        await this.databaseService.deactivateLoan(chainName, user);
                        this.logger.log(`[${chainName}] Removed user ${user} from active loans and database as total debt is less than 1 USD`);
                        continue;
                    }

                    // 如果健康因子低于清算阈值，尝试清算
                    if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
                        const lastAttemptHealthFactor = lastLiquidationMap.get(user);

                        // 如果是首次尝试清算，或者新的健康因子比上次更低，则执行清算
                        if (lastAttemptHealthFactor === undefined || healthFactor < lastAttemptHealthFactor) {
                            this.logger.log(`[${chainName}] Liquidation threshold ${healthFactor} <= ${this.LIQUIDATION_THRESHOLD} reached for user ${user}, attempting liquidation`);
                            // 记录此次清算尝试的健康因子
                            lastLiquidationMap.set(user, healthFactor);
                            await this.executeLiquidation(chainName, user, healthFactor, aaveV3Pool);
                        } else {
                            this.logger.log(`[${chainName}] Skipping liquidation for user ${user} as current health factor ${healthFactor} is not lower than last attempt ${lastAttemptHealthFactor}`);
                        }
                        continue;
                    }

                    // 如果健康因子高于清算阈值，清除清算记录
                    if (lastLiquidationMap.has(user)) {
                        lastLiquidationMap.delete(user);
                    }

                    // 更新下次检查时间
                    const waitTime = this.calculateWaitTime(healthFactor);
                    const nextCheckTime = new Date(Date.now() + waitTime);
                    const formattedDate = this.formatDate(nextCheckTime);
                    this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms (at ${formattedDate}), healthFactor: ${healthFactor}`);

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
            }
        } catch (error) {
            this.logger.error(`[${chainName}] Error checking health factors batch: ${error.message}`);
        }
    }

    private safeStringify(obj: any): string {
        return JSON.stringify(obj, (key, value) =>
            typeof value === 'bigint' ? value.toString() : value
        );
    }

    private async getUserAccountDataBatch(chainName: string, contract: ethers.Contract, users: string[]): Promise<Map<string, UserAccountData>> {
        try {
            const multicallContract = this.getMulticall(chainName);

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

    private calculateWaitTime(healthFactor: number): number {
        // 如果健康因子低于清算阈值，使用最小检查间隔
        if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
            return this.MIN_WAIT_TIME;
        }

        // 如果健康因子低于危险阈值，使用较短检查间隔
        if (healthFactor <= this.CRITICAL_THRESHOLD) {
            return this.MIN_WAIT_TIME * 2; // 400ms
        }

        // 如果健康因子低于健康阈值，使用中等检查间隔
        if (healthFactor <= this.HEALTH_FACTOR_THRESHOLD) {
            // 使用指数函数计算等待时间，健康因子越接近阈值，等待时间越短
            const baseTime = this.MIN_WAIT_TIME * 4; // 800ms
            const maxTime = this.MAX_WAIT_TIME / 2; // 15分钟
            const factor = (healthFactor - this.CRITICAL_THRESHOLD) /
                (this.HEALTH_FACTOR_THRESHOLD - this.CRITICAL_THRESHOLD);

            return Math.floor(baseTime + (maxTime - baseTime) * Math.pow(factor, 2));
        }

        // 如果健康因子高于健康阈值，使用较长检查间隔
        // 使用对数函数计算等待时间，健康因子越高，等待时间越长
        const baseTime = this.MAX_WAIT_TIME / 2; // 15分钟
        const maxTime = this.MAX_WAIT_TIME; // 30分钟
        const factor = (healthFactor - this.HEALTH_FACTOR_THRESHOLD) /
            (2 - this.HEALTH_FACTOR_THRESHOLD); // 假设最大健康因子为2

        // 确保等待时间不超过最大值
        return Math.min(
            Math.floor(baseTime + (maxTime - baseTime) * Math.log1p(factor)),
            this.MAX_WAIT_TIME
        );
    }

    private async executeLiquidation(chainName: string, user: string, healthFactor: number, aaveV3Pool: ethers.Contract) {
        try {
            // 1. 使用 multicall 批量获取用户配置和储备资产列表
            const multicall = this.getMulticall(chainName);
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
            const dataProvider = this.getDataProvider(chainName);

            // 准备 multicall 调用数据
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
                return;
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
                        this.logger.log(`- Debt Value in USD: ${currentVariableDebt} USD, Debt Amount: ${currentStableDebt}, Total Debt: ${totalDebt}`);

                        if (totalDebt > maxDebtAmount) {
                            maxDebtAmount = totalDebt;
                            maxDebtAsset = asset;
                        }
                    }

                    // 处理抵押品数据
                    if (isUsingAsCollateral) {
                        const collateralAmount = BigInt(userReserveData.currentATokenBalance);
                        if (collateralAmount > maxCollateralAmount) {
                            maxCollateralAmount = collateralAmount;
                            maxCollateralAsset = asset;
                        }
                    }
                }
            }

            // 3. 对最大债务资产执行清算
            if (maxDebtAmount === BigInt(0)) {
                this.logger.log(`[${chainName}] No debt found for user ${user}`);
                return;
            }
            if (maxCollateralAmount === BigInt(0)) {
                this.logger.log(`[${chainName}] No collateral assets found for user ${user}`);
                return;
            }
            this.logger.log(`[${chainName}] Executing flash loan liquidation:`);
            this.logger.log(`- User: ${user}`);
            this.logger.log(`- Health Factor: ${healthFactor}`);
            this.logger.log(`- Collateral Asset: ${maxCollateralAsset} (${maxCollateralAmount})`);
            this.logger.log(`- Debt Asset: ${maxDebtAsset} (${maxDebtAmount})`);

            // 4. 执行闪电贷清算
            const flashLoanLiquidation = this.getFlashLoanLiquidation(chainName);

            try {
                // 获取当前 gas 价格并提高 50%
                const gasPrice = await flashLoanLiquidation.runner.provider.getFeeData();
                const maxFeePerGas = gasPrice.maxFeePerGas ? gasPrice.maxFeePerGas * BigInt(15) / BigInt(10) : undefined;
                const maxPriorityFeePerGas = gasPrice.maxPriorityFeePerGas ? gasPrice.maxPriorityFeePerGas * BigInt(15) / BigInt(10) : undefined;

                const tx = await flashLoanLiquidation.executeLiquidation(
                    maxCollateralAsset,
                    maxDebtAsset,
                    user,
                    {
                        maxFeePerGas,
                        maxPriorityFeePerGas
                    }
                );
                this.logger.log(`[${chainName}] Flash loan liquidation executed successfully, tx: ${tx.hash}`);
                await tx.wait();
            } catch (error) {
                this.logger.error(`[${chainName}] Error executing flash loan liquidation: ${error.message}`);
            }
        } catch (error) {
            this.logger.error(`[${chainName}] Error executing liquidation for user ${user}: ${error.message}`);
        }
    }

    private async getTokenPrice(chainName: string, tokenAddress: string): Promise<number> {
        const priceOracle = this.priceOracleCache.get(chainName);
        const price = await priceOracle.getAssetPrice(tokenAddress);
        return Number(price) / 1e8;
    }

    // 在服务销毁时清理定时器
    async onModuleDestroy() {
        if (this.checkInterval) {
            clearInterval(this.checkInterval);
        }
    }
} 