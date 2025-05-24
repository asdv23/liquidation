import { Injectable, Logger, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { ChainService } from '../chain/chain.service';
import { ethers } from 'ethers';
import { ChainConfig } from '../interfaces/chain-config.interface';
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

@Injectable()
export class BorrowDiscoveryService implements OnModuleInit, OnModuleDestroy {
    private readonly logger = new Logger(BorrowDiscoveryService.name);
    private activeLoans: Map<string, Set<string>> = new Map();
    private liquidationTimes: Map<string, Map<string, number>> = new Map();
    private tokenCache: Map<string, Map<string, TokenInfo>> = new Map();
    private readonly LIQUIDATION_THRESHOLD = 1.05; // 清算阈值
    private readonly CRITICAL_THRESHOLD = 1.1; // 危险阈值
    private readonly HEALTH_FACTOR_THRESHOLD = 1.2; // 健康阈值
    private readonly MIN_WAIT_TIME: number; // 最小等待时间（毫秒）
    private readonly MAX_WAIT_TIME: number; // 最大等待时间（毫秒）
    private checkInterval: NodeJS.Timeout;
    private heartbeatInterval: NodeJS.Timeout; // 心跳定时器

    constructor(
        private readonly chainService: ChainService,
        private readonly configService: ConfigService,
        private readonly databaseService: DatabaseService,
    ) {
        // 从环境变量读取配置，默认值：最小200ms，最大30分钟
        this.MIN_WAIT_TIME = this.configService.get<number>('MIN_CHECK_INTERVAL', 200);
        this.MAX_WAIT_TIME = this.configService.get<number>('MAX_CHECK_INTERVAL', 30 * 60 * 1000);
    }

    async onModuleInit() {
        this.logger.log('BorrowDiscoveryService initializing...');
        await this.loadTokenCache();
        await this.loadActiveLoans();
        await this.startListening();
        this.startHealthFactorChecker();
        this.startHeartbeat();
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
                    this.logger.log(`[${chainName}] Checking health factors for ${activeLoans.length} active loans...`);

                    // 获取合约实例
                    const provider = await this.chainService.getProvider(chainName);
                    const config = this.chainService.getChainConfig(chainName);
                    const abiPath = path.join(process.cwd(), 'abi', `${chainName.toUpperCase()}_AAVE_V3_POOL_ABI.json`);
                    const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
                    const contract = new ethers.Contract(
                        config.contracts.lendingPool,
                        abi,
                        provider
                    );

                    // 检查每个活跃借款的健康因子
                    for (const loan of activeLoans) {
                        try {
                            const accountData = await this.getUserAccountData(contract, loan.user);
                            if (!accountData) {
                                this.logger.warn(`[${chainName}] Could not get account data for user ${loan.user}`);
                                continue;
                            }

                            const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
                            const totalDebt = Number(ethers.formatUnits(accountData.totalDebtBase, 8));
                            this.logger.log(`[${chainName}] User ${loan.user} health factor: ${healthFactor}, total debt: ${totalDebt.toFixed(2)} USD`);

                            // 如果总债务为 0，关闭借款记录
                            if (totalDebt === 0) {
                                await this.databaseService.deactivateLoan(chainName, loan.user);
                                this.logger.log(`[${chainName}] Deactivated loan for user ${loan.user} as total debt is 0`);
                                continue;
                            }

                            // 计算下次检查的等待时间
                            const waitTime = this.calculateWaitTime(healthFactor);
                            const nextCheckTime = new Date(Date.now() + waitTime);
                            const formattedDate = this.formatDate(nextCheckTime);

                            // 更新数据库中的健康因子和下次检查时间
                            await this.databaseService.updateLoanHealthFactor(
                                chainName,
                                loan.user,
                                healthFactor,
                                nextCheckTime,
                                totalDebt
                            );

                            this.logger.log(`[${chainName}] Next check for user ${loan.user} in ${waitTime}ms (at ${formattedDate})`);

                            // 如果健康因子低于清算阈值，执行清算
                            if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
                                await this.databaseService.markLiquidationDiscovered(chainName, loan.user);
                                await this.executeLiquidation(chainName, loan.user, contract);
                            }

                            // 更新内存中的活跃贷款集合
                            if (!this.activeLoans.has(chainName)) {
                                this.activeLoans.set(chainName, new Set());
                            }
                            const activeLoansSet = this.activeLoans.get(chainName);
                            if (activeLoansSet) {
                                activeLoansSet.add(loan.user);
                            }
                        } catch (error) {
                            this.logger.error(`[${chainName}] Error checking health factor for user ${loan.user}: ${error.message}`);
                        }
                    }
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

                // 读取对应链的 ABI 文件
                const abiPath = path.join(process.cwd(), 'abi', `${chainName.toUpperCase()}_AAVE_V3_POOL_ABI.json`);
                const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));

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

                const contract = new ethers.Contract(
                    config.contracts.lendingPool,
                    abi,
                    provider as ethers.ContractRunner
                );

                // 验证合约连接
                try {
                    // 尝试获取一个已知的储备资产数据
                    const wethAddress = chainName === 'base'
                        ? '0x4200000000000000000000000000000000000006'  // Base WETH
                        : '0x4200000000000000000000000000000000000006'; // Optimism WETH

                    const reserveData = await contract.getReserveData(wethAddress);
                    this.logger.log(`[${chainName}] Successfully connected to Aave V3 Pool at ${config.contracts.lendingPool}`);
                    this.logger.log(`[${chainName}] WETH Reserve Data:`);
                    this.logger.log(`- Current Liquidity Rate: ${ethers.formatUnits(reserveData.currentLiquidityRate, 27)}`);
                    this.logger.log(`- Current Variable Borrow Rate: ${ethers.formatUnits(reserveData.currentVariableBorrowRate, 27)}`);
                    this.logger.log(`- Current Stable Borrow Rate: ${ethers.formatUnits(reserveData.currentStableBorrowRate, 27)}`);
                } catch (error) {
                    this.logger.error(`[${chainName}] Failed to verify contract connection: ${error.message}`);
                    continue;
                }

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

                            if (!this.activeLoans.has(chainName)) {
                                this.activeLoans.set(chainName, new Set());
                            }
                            const activeLoansSet = this.activeLoans.get(chainName);
                            if (activeLoansSet) {
                                activeLoansSet.add(onBehalfOf);
                            }

                            // 获取用户账户数据并创建/更新贷款记录
                            const accountData = await this.getUserAccountData(contract, onBehalfOf);
                            if (accountData) {
                                const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
                                const totalDebt = Number(ethers.formatUnits(accountData.totalDebtBase, 8));
                                const waitTime = this.calculateWaitTime(healthFactor);
                                const nextCheckTime = new Date(Date.now() + waitTime);

                                await this.databaseService.updateLoanHealthFactor(
                                    chainName,
                                    onBehalfOf,
                                    healthFactor,
                                    nextCheckTime,
                                    totalDebt
                                );

                                this.logger.log(`[${chainName}] Created/Updated loan record for user ${onBehalfOf}`);
                                this.logger.log(`[${chainName}] Health factor: ${healthFactor}`);
                                this.logger.log(`[${chainName}] Total debt: ${totalDebt.toFixed(2)} USD`);
                            }

                            await this.checkHealthFactor(chainName, onBehalfOf, contract);
                        } catch (error) {
                            this.logger.error(`[${chainName}] Error processing Borrow event: ${error.message}`);
                        }
                    });

                    this.logger.log(`[${chainName}] Borrow event listener setup completed`);
                } catch (error) {
                    this.logger.error(`[${chainName}] Failed to set up Borrow event listener: ${error.message}`);
                }

                // 监听 Repay 事件
                contract.on('Repay', async (reserve, user, repayer, amount, useATokens, event) => {
                    try {
                        const tokenInfo = await this.getTokenInfo(chainName, reserve, provider);
                        this.logger.log(`[${chainName}] Repay event detected:`);
                        this.logger.log(`- Reserve: ${reserve} (${tokenInfo.symbol})`);
                        this.logger.log(`- User: ${user}`);
                        this.logger.log(`- Repayer: ${repayer}`);
                        this.logger.log(`- Amount: ${this.formatAmount(amount, tokenInfo.decimals)} ${tokenInfo.symbol}`);
                        this.logger.log(`- Use ATokens: ${useATokens}`);
                        this.logger.log(`- Transaction Hash: ${event?.transactionHash || event?.log?.transactionHash}`);

                        // 获取用户账户数据
                        const accountData = await this.getUserAccountData(contract, user);
                        if (!accountData) {
                            this.logger.warn(`[${chainName}] Could not get account data for user ${user}`);
                            return;
                        }

                        const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
                        const totalDebt = Number(ethers.formatUnits(accountData.totalDebtBase, 8));
                        const waitTime = this.calculateWaitTime(healthFactor);
                        const nextCheckTime = new Date(Date.now() + waitTime);

                        // 检查贷款记录是否存在
                        const activeLoans = await this.databaseService.getActiveLoans(chainName);
                        const loanExists = activeLoans.some(loan => loan.user.toLowerCase() === user.toLowerCase());

                        if (loanExists) {
                            // 更新数据库中的健康因子和下次检查时间
                            await this.databaseService.updateLoanHealthFactor(
                                chainName,
                                user,
                                healthFactor,
                                nextCheckTime,
                                totalDebt
                            );

                            this.logger.log(`[${chainName}] Updated loan record for user ${user}`);
                            this.logger.log(`[${chainName}] Health factor: ${healthFactor}`);
                            this.logger.log(`[${chainName}] Total debt: ${totalDebt.toFixed(2)} USD`);

                            // 如果总债务为 0，关闭借款记录
                            if (totalDebt === 0) {
                                await this.databaseService.deactivateLoan(chainName, user);
                                this.logger.log(`[${chainName}] Deactivated loan for user ${user} as total debt is 0`);

                                // 从内存中移除该贷款
                                const activeLoansSet = this.activeLoans.get(chainName);
                                if (activeLoansSet) {
                                    activeLoansSet.delete(user);
                                }
                            }
                        } else {
                            // 如果贷款记录不存在，记录告警
                            this.logger.warn(`[${chainName}] Received Repay event for non-existent loan: user=${user}, amount=${this.formatAmount(amount, tokenInfo.decimals)} ${tokenInfo.symbol}`);
                            this.logger.warn(`[${chainName}] This might indicate a missed Borrow event or database inconsistency`);
                        }
                    } catch (error) {
                        this.logger.error(`[${chainName}] Error processing Repay event: ${error.message}`);
                    }
                });

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

                        // 获取用户账户数据
                        const accountData = await this.getUserAccountData(contract, user);
                        if (!accountData) {
                            this.logger.warn(`[${chainName}] Could not get account data for user ${user}`);
                            return;
                        }

                        const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
                        const totalDebt = Number(ethers.formatUnits(accountData.totalDebtBase, 8));

                        // 记录清算信息
                        await this.databaseService.recordLiquidation(
                            chainName,
                            user,
                            liquidator,
                            event?.transactionHash || event?.log?.transactionHash
                        );

                        this.logger.log(`[${chainName}] Recorded liquidation for user ${user}`);
                        this.logger.log(`[${chainName}] Final health factor: ${healthFactor}`);
                        this.logger.log(`[${chainName}] Final total debt: ${totalDebt.toFixed(2)} USD`);

                        // 从内存中移除该贷款
                        const activeLoansSet = this.activeLoans.get(chainName);
                        if (activeLoansSet) {
                            activeLoansSet.delete(user);
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
        // 每分钟检查一次需要更新的贷款
        this.checkInterval = setInterval(async () => {
            try {
                const loansToCheck = await this.databaseService.getLoansToCheck();
                for (const loan of loansToCheck) {
                    const provider = await this.chainService.getProvider(loan.chainName);
                    const config = this.chainService.getChainConfig(loan.chainName);
                    const abiPath = path.join(process.cwd(), 'abi', `${loan.chainName.toUpperCase()}_AAVE_V3_POOL_ABI.json`);
                    const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
                    const contract = new ethers.Contract(
                        config.contracts.lendingPool,
                        abi,
                        provider
                    );
                    await this.checkHealthFactor(loan.chainName, loan.user, contract);
                }
            } catch (error) {
                this.logger.error(`Error in health factor checker: ${error.message}`);
            }
        }, 60000); // 每分钟检查一次
    }

    private startHeartbeat() {
        // 启动时立即执行一次心跳
        this.printHeartbeat();
        // 每小时执行一次心跳
        this.heartbeatInterval = setInterval(() => {
            this.printHeartbeat();
        }, 60 * 60 * 1000);
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

    private printHeartbeat() {
        const chains = this.chainService.getActiveChains();
        const now = new Date();
        const formattedDate = this.formatDate(now);

        this.logger.log(`心跳检测 - 正在监听的合约：`);
        for (const chainName of chains) {
            const config = this.chainService.getChainConfig(chainName);
            this.logger.log(`[${chainName}] LendingPool: ${config.contracts.lendingPool}`);

            // 从数据库获取当前活跃贷款数量
            this.databaseService.getActiveLoans(chainName)
                .then(activeLoans => {
                    this.logger.log(`[${chainName}] 当前活跃贷款数量: ${activeLoans.length}`);
                })
                .catch(error => {
                    this.logger.error(`[${chainName}] Error getting active loans count: ${error.message}`);
                });
        }
    }

    private async checkHealthFactor(chainName: string, user: string, contract: ethers.Contract) {
        try {
            const accountData = await this.getUserAccountData(contract, user);
            if (!accountData) return;

            const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
            const totalDebt = Number(ethers.formatUnits(accountData.totalDebtBase, 8));
            this.logger.log(`[${chainName}] User ${user} health factor: ${healthFactor}`);
            this.logger.log(`[${chainName}] User ${user} total debt: ${totalDebt.toFixed(2)} USD`);

            // 计算下次检查的等待时间（毫秒）
            const waitTime = this.calculateWaitTime(healthFactor);
            const nextCheckTime = new Date(Date.now() + waitTime);
            const formattedDate = this.formatDate(nextCheckTime);

            // 更新数据库中的健康因子、总债务和下次检查时间
            await this.databaseService.updateLoanHealthFactor(
                chainName,
                user,
                healthFactor,
                nextCheckTime,
                totalDebt
            );

            this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms (at ${formattedDate})`);

            // 如果健康因子低于清算阈值，执行清算
            if (healthFactor <= this.LIQUIDATION_THRESHOLD) {
                // 记录发现可清算的时间
                await this.databaseService.markLiquidationDiscovered(chainName, user);
                await this.executeLiquidation(chainName, user, contract);
                return;
            }

        } catch (error) {
            this.logger.error(`[${chainName}] Error checking health factor for user ${user}: ${error.message}`);
        }
    }

    private async getUserAccountData(contract: ethers.Contract, user: string): Promise<UserAccountData | null> {
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
        } catch (error) {
            this.logger.error(`Error getting user account data: ${error.message}`);
            return null;
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
            this.logger.log(`[${chainName}] Executing liquidation for user ${user}`);
            // TODO: 实现清算逻辑
            // 1. 获取用户的债务信息
            // 2. 计算清算金额
            // 3. 发送清算交易

            // 清算后，将贷款标记为非活跃
            // await this.databaseService.deactivateLoan(chainName, user);
        } catch (error) {
            this.logger.error(`[${chainName}] Error executing liquidation for user ${user}: ${error.message}`);
        }
    }

    private recordLiquidationTime(chainName: string, user: string) {
        if (!this.liquidationTimes.has(chainName)) {
            this.liquidationTimes.set(chainName, new Map());
        }
        const chainLiquidationTimes = this.liquidationTimes.get(chainName);
        if (chainLiquidationTimes) {
            chainLiquidationTimes.set(user, Date.now());
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