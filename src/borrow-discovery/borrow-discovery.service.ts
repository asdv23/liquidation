import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ChainService } from '../chain/chain.service';
import { ethers } from 'ethers';
import { ChainConfig } from '../interfaces/chain-config.interface';

@Injectable()
export class BorrowDiscoveryService implements OnModuleInit {
    private readonly logger = new Logger(BorrowDiscoveryService.name);
    private activeLoans: Map<string, Set<string>> = new Map();
    private liquidationTimes: Map<string, Map<string, number>> = new Map();

    constructor(
        private readonly chainService: ChainService,
    ) { }

    async onModuleInit() {
        this.logger.log('BorrowDiscoveryService initializing...');
        await this.startListening();
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

                    const activeLoansSet = this.activeLoans.get(chainName);
                    if (activeLoansSet) {
                        activeLoansSet.delete(user);
                    }
                    this.recordLiquidationTime(chainName, user);
                });

                this.logger.log(`[${chainName}] Successfully set up event listeners and verified contract connection`);
            } catch (error) {
                this.logger.error(`Failed to set up event listeners for ${chainName}: ${error.message}`);
            }
        }
    }

    private async checkHealthFactor(chainName: string, user: string, contract: ethers.Contract) {
        try {
            const healthFactor = await contract.getUserAccountData(user);
            this.logger.log(`[${chainName}] Health factor for user ${user}: ${ethers.formatUnits(healthFactor.healthFactor, 18)}`);
            // TODO: 实现健康因子检查逻辑
        } catch (error) {
            this.logger.error(`Failed to check health factor for user ${user} on ${chainName}: ${error.message}`);
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
} 