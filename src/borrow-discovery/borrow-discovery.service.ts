import { Injectable, Logger, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { ChainService } from '../chain/chain.service';
import { ethers } from 'ethers';
import { ConfigService } from '@nestjs/config';
import { DatabaseService } from '../database/database.service';
import * as fs from 'fs';
import * as path from 'path';

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
    private readonly LIQUIDATION_THRESHOLD = 1.005; // 清算阈值
    private readonly CRITICAL_THRESHOLD = 1.01; // 危险阈值
    private readonly HEALTH_FACTOR_THRESHOLD = 1.02; // 健康阈值
    private readonly MIN_WAIT_TIME: number; // 最小等待时间（毫秒）
    private readonly MAX_WAIT_TIME: number; // 最大等待时间（毫秒）
    private readonly PRIVATE_KEY: string; // EOA 私钥
    private checkInterval: NodeJS.Timeout;
    private heartbeatInterval: NodeJS.Timeout; // 心跳定时器
    private contractCache: Map<string, ethers.Contract> = new Map(); // 合约缓存
    private providerCache: Map<string, ethers.Provider> = new Map(); // Provider 缓存
    private signerCache: Map<string, ethers.Signer> = new Map(); // Signer 缓存
    private dataProviderCache: Map<string, ethers.Contract> = new Map(); // DataProvider 缓存

    constructor(
        private readonly chainService: ChainService,
        private readonly configService: ConfigService,
        private readonly databaseService: DatabaseService,
    ) {
        // 从环境变量读取配置，默认值：最小200ms，最大30分钟
        this.MIN_WAIT_TIME = this.configService.get<number>('MIN_CHECK_INTERVAL', 1000);
        this.MAX_WAIT_TIME = this.configService.get<number>('MAX_CHECK_INTERVAL', 30 * 60 * 1000);
        this.PRIVATE_KEY = this.configService.get<string>('PRIVATE_KEY');
    }

    async onModuleInit() {
        this.logger.log('BorrowDiscoveryService initializing...');
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

                // 初始化合约
                const config = this.chainService.getChainConfig(chainName);
                const abiPath = path.join(process.cwd(), 'abi', `${chainName.toUpperCase()}_AAVE_V3_POOL_ABI.json`);
                const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
                const contract = new ethers.Contract(
                    config.contracts.lendingPool,
                    abi,
                    signer
                );
                this.contractCache.set(chainName, contract);

                // 初始化 DataProvider
                const addressesProviderAddress = await contract.ADDRESSES_PROVIDER();
                const addressesProviderAbi = [
                    'function getPoolDataProvider() view returns (address)'
                ];
                const addressesProvider = new ethers.Contract(
                    addressesProviderAddress,
                    addressesProviderAbi,
                    signer
                );
                const dataProviderAddress = await addressesProvider.getPoolDataProvider();

                const dataProviderAbi = [
                    'function getUserReserveData(address asset, address user) view returns (uint256 currentATokenBalance, uint256 currentStableDebt, uint256 currentVariableDebt, uint256 principalStableDebt, uint256 scaledVariableDebt, uint256 stableBorrowRate, uint256 liquidityRate, uint40 stableRateLastUpdated, bool usageAsCollateralEnabled)'
                ];
                const dataProvider = new ethers.Contract(
                    dataProviderAddress,
                    dataProviderAbi,
                    signer
                );
                this.dataProviderCache.set(chainName, dataProvider);

                this.logger.log(`[${chainName}] Initialized provider, signer, contract and dataProvider`);
            } catch (error) {
                this.logger.error(`[${chainName}] Failed to initialize resources: ${error.message}`);
            }
        }
    }

    private getContract(chainName: string): ethers.Contract {
        const contract = this.contractCache.get(chainName);
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

    private async getTokenInfo(chainName: string, address: string, provider: ethers.Provider): Promise<TokenInfo> {
        const normalizedAddress = address.toLowerCase();

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
            const contract = new ethers.Contract(normalizedAddress, erc20Abi, provider);
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
                    const now = new Date();
                    for (const loan of activeLoans) {
                        activeLoansMap.set(loan.user, {
                            nextCheckTime: now, // 设置为当前时间，确保立即检查
                            healthFactor: 1.0 // 假设健康因子为1.0，需要根据实际情况更新
                        });
                    }

                    this.logger.log(`[${chainName}] Loaded ${activeLoansMap.size} active loans into memory, will check immediately`);
                }
            }
        } catch (error) {
            this.logger.error(`Error loading active loans: ${error.message}`);
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
                const code = await provider.getCode(config.contracts.lendingPool);
                if (code === '0x') {
                    this.logger.error(`[${chainName}] No contract code found at address ${config.contracts.lendingPool}`);
                    continue;
                }
                this.logger.log(`[${chainName}] Contract code found at ${config.contracts.lendingPool}`);

                const contract = this.getContract(chainName);

                // 监听 Borrow 事件
                this.logger.log(`[${chainName}] Setting up Borrow event listener...`);
                try {
                    contract.on('Borrow', async (reserve, user, onBehalfOf, amount, interestRateMode, borrowRate, referralCode, event) => {
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

                            // 更新内存中的活跃贷款集合
                            if (!this.activeLoans.has(chainName)) {
                                this.activeLoans.set(chainName, new Map());
                            }
                            const activeLoansMap = this.activeLoans.get(chainName);
                            if (activeLoansMap) {
                                // 设置初始检查时间为现在，让健康因子检查器立即检查
                                activeLoansMap.set(onBehalfOf, {
                                    nextCheckTime: new Date(), // 设置为当前时间，确保立即检查
                                    healthFactor: 1.0
                                });
                            }

                            // 创建贷款记录
                            await this.databaseService.createOrUpdateLoan(chainName, onBehalfOf);

                            this.logger.log(`[${chainName}] Created/Updated loan record for user ${onBehalfOf}`);
                        } catch (error) {
                            this.logger.error(`[${chainName}] Error processing Borrow event: ${error.message}`);
                        }
                    });

                    this.logger.log(`[${chainName}] Borrow event listener setup completed`);
                } catch (error) {
                    this.logger.error(`[${chainName}] Failed to set up Borrow event listener: ${error.message}`);
                }

                // 监听 LiquidationCall 事件
                contract.on('LiquidationCall', async (collateralAsset, debtAsset, user, debtToCover, liquidatedCollateralAmount, liquidator, receiveAToken, event) => {
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

                        // 从内存中移除该贷款
                        const activeLoansMap = this.activeLoans.get(chainName);
                        if (activeLoansMap) {
                            activeLoansMap.delete(user);

                            // 如果数据库中有借款信息，则记录清算信息
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
                });

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

            const usersToCheck = Array.from(activeLoansMap.entries())
                .filter(([_, info]) => info.nextCheckTime <= new Date())
                .map(([user]) => user);

            if (usersToCheck.length === 0) return;

            // 将用户分批处理，每批最多 100 个
            const BATCH_SIZE = 100;
            for (let i = 0; i < usersToCheck.length; i += BATCH_SIZE) {
                const batchUsers = usersToCheck.slice(i, i + BATCH_SIZE);
                this.logger.log(`[${chainName}] Checking health factors for batch ${i / BATCH_SIZE + 1}/${Math.ceil(usersToCheck.length / BATCH_SIZE)} (${batchUsers.length}/${activeLoansMap.size} users)...`);

                const contract = this.getContract(chainName);
                const accountDataMap = await this.getUserAccountDataBatch(contract, batchUsers);

                for (const user of batchUsers) {
                    const accountData = accountDataMap.get(user);
                    if (!accountData) continue;

                    const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
                    const totalDebt = Number(ethers.formatUnits(accountData.totalDebtBase, 8));

                    // 如果总债务等于 0，则从内存和数据库中移除该用户
                    if (totalDebt === 0) {
                        activeLoansMap.delete(user);
                        await this.databaseService.deactivateLoan(chainName, user);
                        this.logger.log(`[${chainName}] Removed user ${user} from active loans and database as total debt is 0`);
                        continue;
                    }

                    // 如果健康因子低于清算阈值，尝试清算
                    if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
                        this.logger.log(`[${chainName}] Liquidation threshold ${healthFactor} <= ${this.LIQUIDATION_THRESHOLD} reached for user ${user}, attempting liquidation`);
                        await this.executeLiquidation(chainName, user, contract);
                        continue;
                    }

                    // 如果健康因子低于健康阈值，记录日志
                    if (healthFactor <= this.HEALTH_FACTOR_THRESHOLD) {
                        this.logger.log(`[${chainName}] Health factor ${healthFactor} <= ${this.HEALTH_FACTOR_THRESHOLD} reached for user ${user}`);
                    }

                    // 更新下次检查时间
                    const waitTime = this.calculateWaitTime(healthFactor);
                    const nextCheckTime = new Date(Date.now() + waitTime);
                    const formattedDate = this.formatDate(nextCheckTime);
                    this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms (at ${formattedDate}), healthFactor: ${healthFactor}`);

                    // 更新内存中的健康因子和下次检查时间
                    activeLoansMap.set(user, {
                        nextCheckTime: new Date(Date.now() + waitTime),
                        healthFactor: healthFactor
                    });
                }
            }
        } catch (error) {
            this.logger.error(`[${chainName}] Error checking health factors batch: ${error.message}`);
        }
    }

    private async getUserAccountDataBatch(contract: ethers.Contract, users: string[]): Promise<Map<string, UserAccountData>> {
        try {
            // 创建 multicall 合约
            const multicallAddress = '0xcA11bde05977b3631167028862bE2a173976CA11'; // 通用 multicall 地址
            const multicallAbi = [
                'function aggregate(tuple(address target, bytes callData)[] calls) view returns (uint256 blockNumber, bytes[] returnData)'
            ];
            const multicallContract = new ethers.Contract(multicallAddress, multicallAbi, contract.runner);

            // 准备调用数据
            const calls = users.map(user => ({
                target: contract.target,
                callData: contract.interface.encodeFunctionData('getUserAccountData', [user])
            }));

            // 执行批量调用
            const [, returnData] = await multicallContract.aggregate(calls);

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

    private async executeLiquidation(chainName: string, user: string, contract: ethers.Contract) {
        try {
            // 1. 获取用户的配置信息
            const userConfig = await contract.getUserConfiguration(user);

            // 2. 获取所有储备资产列表
            const reservesList = await contract.getReservesList();

            // 3. 使用 multicall 批量查询所有借贷资产的债务数据
            const multicallAddress = '0xcA11bde05977b3631167028862bE2a173976CA11';
            const multicallAbi = [
                'function aggregate(tuple(address target, bytes callData)[] calls) view returns (uint256 blockNumber, bytes[] returnData)'
            ];
            const multicallContract = new ethers.Contract(multicallAddress, multicallAbi, contract.runner);
            const dataProvider = this.getDataProvider(chainName);

            // 准备 multicall 调用数据
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

            // 执行批量查询
            const [, returnData] = await multicallContract.aggregate(calls);

            // 解析返回数据并找出最大债务资产
            let maxDebtAsset = null;
            let maxDebtAmount = BigInt(0);

            for (let i = 0; i < borrowingAssets.length; i++) {
                const asset = borrowingAssets[i];
                const userReserveData = dataProvider.interface.decodeFunctionResult(
                    'getUserReserveData',
                    returnData[i]
                );

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

            // 4. 对最大债务资产执行清算
            this.logger.log(`[${chainName}] Executing liquidation for user ${user}, maxDebtAsset: ${maxDebtAsset.asset}, debtAmount: ${maxDebtAmount}`);
            // try {
            //     const hasVariableDebt = maxDebtAsset.currentVariableDebt > BigInt(0);
            //     const debtAmount = hasVariableDebt
            //         ? maxDebtAsset.currentVariableDebt
            //         : maxDebtAsset.currentStableDebt;

            //     // 执行闪电贷清算
            //     const flashLoanParams = {
            //         receiverAddress: contract.target,
            //         asset: maxDebtAsset.asset,
            //         amount: debtAmount,
            //         params: ethers.AbiCoder.defaultAbiCoder().encode(
            //             ['address', 'bool'],
            //             [user, false] // user address and receiveAToken flag
            //         ),
            //         referralCode: 0
            //     };

            //     // 发送清算交易
            //     const tx = await contract.flashLoanSimple(
            //         flashLoanParams.receiverAddress,
            //         flashLoanParams.asset,
            //         flashLoanParams.amount,
            //         flashLoanParams.params,
            //         flashLoanParams.referralCode
            //     );

            //     this.logger.log(`[${chainName}] Liquidation transaction sent for asset ${maxDebtAsset.asset}: ${tx.hash}`);

            //     // 等待交易确认
            //     const receipt = await tx.wait();
            //     this.logger.log(`[${chainName}] Liquidation transaction confirmed for asset ${maxDebtAsset.asset}: ${receipt.hash}`);
            // } catch (error) {
            //     this.logger.error(`[${chainName}] Error executing liquidation for asset ${maxDebtAsset.asset}: ${error.message}`);
            // }
        } catch (error) {
            this.logger.error(`[${chainName}] Error executing liquidation for user ${user}: ${error.message}`);
        }
    }

    // 在服务销毁时清理定时器
    async onModuleDestroy() {
        if (this.checkInterval) {
            clearInterval(this.checkInterval);
        }
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
        }
    }
} 