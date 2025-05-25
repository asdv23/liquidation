import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { PrismaClient } from '@prisma/client';
import * as path from 'path';

@Injectable()
export class DatabaseService implements OnModuleInit {
    private readonly logger = new Logger(DatabaseService.name);
    private static prisma: PrismaClient;

    constructor() {
        if (!DatabaseService.prisma) {
            const dbPath = path.join(process.cwd(), 'prisma', 'dev.db');
            this.logger.log(`Using SQLite database at: ${dbPath}`);
            DatabaseService.prisma = new PrismaClient();
        }
    }

    async onModuleInit() {
        try {
            await DatabaseService.prisma.$connect();
            this.logger.log('Database connection established');
        } catch (error) {
            this.logger.error(`Failed to connect to database: ${error.message}`);
            throw error;
        }
    }

    get prisma() {
        return DatabaseService.prisma;
    }

    async markLiquidationDiscovered(chainName: string, user: string) {
        try {
            // 检查是否存在贷款记录
            const existingLoan = await this.prisma.loan.findFirst({
                where: {
                    chainName,
                    user: user.toLowerCase(),
                },
            });

            if (!existingLoan) {
                this.logger.warn(`[${chainName}] No active loan found for user ${user} when marking liquidation discovered`);
                return;
            }

            // 更新清算发现时间
            return await this.prisma.loan.update({
                where: {
                    chainName_user: {
                        chainName,
                        user: user.toLowerCase(),
                    },
                },
                data: {
                    liquidationDiscoveredAt: new Date(),
                },
            });
        } catch (error) {
            this.logger.error(`Error marking liquidation discovered: ${error.message}`);
            throw error;
        }
    }

    async recordLiquidation(
        chainName: string,
        user: string,
        liquidator: string,
        txHash: string
    ) {
        try {
            const loan = await this.prisma.loan.findUnique({
                where: {
                    chainName_user: {
                        chainName,
                        user,
                    },
                },
            });

            if (!loan) {
                throw new Error(`Loan not found for user ${user} on chain ${chainName}`);
            }

            const now = new Date();
            const liquidationDelay = loan.liquidationDiscoveredAt
                ? now.getTime() - loan.liquidationDiscoveredAt.getTime()
                : null;

            return await this.prisma.loan.update({
                where: {
                    chainName_user: {
                        chainName,
                        user,
                    },
                },
                data: {
                    isActive: false,
                    liquidationTime: now,
                    liquidator,
                    liquidationTxHash: txHash,
                    liquidationDelay,
                },
            });
        } catch (error) {
            this.logger.error(`Error recording liquidation: ${error.message}`);
            throw error;
        }
    }

    async getActiveLoans(chainName?: string) {
        try {
            return await this.prisma.loan.findMany({
                where: {
                    isActive: true,
                    ...(chainName ? { chainName } : {}),
                }
            });
        } catch (error) {
            this.logger.error(`Error getting active loans: ${error.message}`);
            throw error;
        }
    }

    async deactivateLoan(chainName: string, user: string) {
        try {
            return await this.prisma.loan.update({
                where: {
                    chainName_user: {
                        chainName,
                        user,
                    },
                },
                data: {
                    isActive: false,
                },
            });
        } catch (error) {
            this.logger.error(`Error deactivating loan: ${error.message}`);
            throw error;
        }
    }

    async getLoansToCheck() {
        try {
            return await this.prisma.loan.findMany({
                where: {
                    isActive: true
                }
            });
        } catch (error) {
            this.logger.error(`Error getting loans to check: ${error.message}`);
            throw error;
        }
    }

    // Token 相关方法
    async getToken(chainName: string, address: string) {
        try {
            return await this.prisma.token.findUnique({
                where: {
                    chainName_address: {
                        chainName,
                        address: address.toLowerCase(),
                    },
                },
            });
        } catch (error) {
            this.logger.error(`Error getting token: ${error.message}`);
            throw error;
        }
    }

    async saveToken(chainName: string, address: string, symbol: string, decimals: number) {
        try {
            return await this.prisma.token.upsert({
                where: {
                    chainName_address: {
                        chainName,
                        address: address.toLowerCase(),
                    },
                },
                update: {
                    symbol,
                    decimals,
                    updatedAt: new Date(),
                },
                create: {
                    chainName,
                    address: address.toLowerCase(),
                    symbol,
                    decimals,
                },
            });
        } catch (error) {
            this.logger.error(`Error saving token: ${error.message}`);
            throw error;
        }
    }

    async getAllTokens(chainName?: string) {
        try {
            return await this.prisma.token.findMany({
                where: chainName ? { chainName } : undefined,
            });
        } catch (error) {
            this.logger.error(`Error getting all tokens: ${error.message}`);
            throw error;
        }
    }

    async createOrUpdateLoan(chainName: string, user: string) {
        try {
            await this.prisma.loan.upsert({
                where: {
                    chainName_user: {
                        chainName,
                        user: user.toLowerCase(),
                    },
                },
                update: {
                    isActive: true,
                    updatedAt: new Date(),
                },
                create: {
                    chainName,
                    user: user.toLowerCase(),
                    isActive: true,
                    healthFactor: 0
                },
            });
        } catch (error) {
            this.logger.error(`Error creating/updating loan: ${error.message}`);
            throw error;
        }
    }
} 