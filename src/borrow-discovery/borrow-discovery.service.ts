import { Injectable, Logger, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { ChainService } from '../chain/chain.service';
import { ethers } from 'ethers';
import { ChainConfig } from '../interfaces/chain-config.interface';
import { ConfigService } from '@nestjs/config';
import { DatabaseService } from '../database/database.service';

interface UserAccountData {
    totalCollateralBase: bigint;
    totalDebtBase: bigint;
    availableBorrowsBase: bigint;
    currentLiquidationThreshold: bigint;
    ltv: bigint;
    healthFactor: bigint;
}

@Injectable()
export class BorrowDiscoveryService implements OnModuleInit, OnModuleDestroy {
    private readonly logger = new Logger(BorrowDiscoveryService.name);
    private activeLoans: Map<string, Set<string>> = new Map();
    private liquidationTimes: Map<string, Map<string, number>> = new Map();
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
        await this.startListening();
        this.startHealthFactorChecker();
        this.startHeartbeat(); // 启动心跳
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

                const contract = new ethers.Contract(
                    config.contracts.lendingPool,
                    [
                        'event Borrow(address indexed user, address indexed onBehalfOf, uint256 amount, uint256 interestRateMode, uint256 borrowRate, uint16 indexed referral)',
                        'event Repay(address indexed user, address indexed repayer, uint256 amount, bool useATokens)',
                        'event LiquidationCall(address indexed collateralAsset, address indexed debtAsset, address indexed user, uint256 debtToCover, uint256 liquidatedCollateralAmount, address liquidator, bool receiveAToken)',
                        'function getAddressesProvider() view returns (address)',
                        'function getReserveData(address asset) view returns (tuple(uint256 configuration, uint128 liquidityIndex, uint128 currentLiquidityRate, uint128 variableBorrowIndex, uint128 currentVariableBorrowRate, uint128 currentStableBorrowRate, uint40 lastUpdateTimestamp, uint16 id, address aTokenAddress, address stableDebtTokenAddress, address variableDebtTokenAddress, address interestRateStrategyAddress, uint128 accruedToTreasury, uint128 unbacked, uint128 isolationModeTotalDebt))'
                    ],
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
                contract.on('Borrow', async (user, onBehalfOf, amount, interestRateMode, borrowRate, referral, event) => {
                    this.logger.log(`[${chainName}] Borrow event detected: user=${user}, amount=${ethers.formatEther(amount)} ETH`);
                    if (!this.activeLoans.has(chainName)) {
                        this.activeLoans.set(chainName, new Set());
                    }
                    const activeLoansSet = this.activeLoans.get(chainName);
                    if (activeLoansSet) {
                        activeLoansSet.add(onBehalfOf);
                    }
                    await this.checkHealthFactor(chainName, onBehalfOf, contract);
                });

                // 监听 Repay 事件
                contract.on('Repay', async (user, repayer, amount, useATokens, event) => {
                    this.logger.log(`[${chainName}] Repay event detected: user=${user}, amount=${ethers.formatEther(amount)} ETH`);
                    const activeLoansSet = this.activeLoans.get(chainName);
                    if (activeLoansSet) {
                        activeLoansSet.delete(user);
                    }
                });

                // 监听 LiquidationCall 事件
                contract.on('LiquidationCall', async (collateralAsset, debtAsset, user, debtToCover, liquidatedCollateralAmount, liquidator, receiveAToken, event) => {
                    this.logger.log(`[${chainName}] LiquidationCall event detected:`);
                    this.logger.log(`- User: ${user}`);
                    this.logger.log(`- Debt to Cover: ${ethers.formatEther(debtToCover)} ETH`);
                    this.logger.log(`- Liquidated Amount: ${ethers.formatEther(liquidatedCollateralAmount)} ETH`);
                    this.logger.log(`- Liquidator: ${liquidator}`);

                    // 记录清算信息
                    await this.databaseService.recordLiquidation(
                        chainName,
                        user,
                        liquidator,
                        event.transactionHash
                    );
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
                    const contract = new ethers.Contract(
                        config.contracts.lendingPool,
                        [
                            'function getUserAccountData(address user) view returns (tuple(uint256 totalCollateralBase, uint256 totalDebtBase, uint256 availableBorrowsBase, uint256 currentLiquidationThreshold, uint256 ltv, uint256 healthFactor))'
                        ],
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

    private printHeartbeat() {
        const chains = this.chainService.getActiveChains();
        const now = new Date().toISOString();

        this.logger.log(`[${now}] 心跳检测 - 正在监听的合约：`);
        chains.forEach(chainName => {
            const config = this.chainService.getChainConfig(chainName);
            this.logger.log(`[${chainName}] LendingPool: ${config.contracts.lendingPool}`);

            // 输出当前活跃贷款数量
            const activeLoansSet = this.activeLoans.get(chainName);
            const activeLoansCount = activeLoansSet ? activeLoansSet.size : 0;
            this.logger.log(`[${chainName}] 当前活跃贷款数量: ${activeLoansCount}`);
        });
    }

    private async checkHealthFactor(chainName: string, user: string, contract: ethers.Contract) {
        try {
            const accountData = await this.getUserAccountData(contract, user);
            if (!accountData) return;

            const healthFactor = this.calculateHealthFactor(accountData.healthFactor);
            this.logger.log(`[${chainName}] User ${user} health factor: ${healthFactor}`);

            // 计算下次检查的等待时间（毫秒）
            const waitTime = this.calculateWaitTime(healthFactor);
            const nextCheckTime = new Date(Date.now() + waitTime);

            // 更新数据库中的健康因子和下次检查时间
            await this.databaseService.updateLoanHealthFactor(
                chainName,
                user,
                healthFactor,
                nextCheckTime
            );

            this.logger.log(`[${chainName}] Next check for user ${user} in ${waitTime}ms (at ${nextCheckTime})`);

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

        return Math.floor(baseTime + (maxTime - baseTime) * Math.log1p(factor));
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