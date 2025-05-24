import { Injectable, Logger } from '@nestjs/common';
import { PrismaClient } from '@prisma/client';

@Injectable()
export class DatabaseService {
    private readonly logger = new Logger(DatabaseService.name);
    private prisma: PrismaClient;

    constructor() {
        this.prisma = new PrismaClient();
    }

    async updateLoanHealthFactor(
        chainName: string,
        user: string,
        healthFactor: number,
        nextCheckTime: Date
    ) {
        try {
            return await this.prisma.loan.upsert({
                where: {
                    chainName_user: {
                        chainName,
                        user,
                    },
                },
                update: {
                    healthFactor,
                    lastCheckTime: new Date(),
                    nextCheckTime,
                    isActive: true,
                },
                create: {
                    chainName,
                    user,
                    healthFactor,
                    nextCheckTime,
                    isActive: true,
                },
            });
        } catch (error) {
            this.logger.error(`Error updating loan health factor: ${error.message}`);
            throw error;
        }
    }

    async markLiquidationDiscovered(chainName: string, user: string) {
        try {
            return await this.prisma.loan.update({
                where: {
                    chainName_user: {
                        chainName,
                        user,
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

    async getActiveLoans() {
        try {
            return await this.prisma.loan.findMany({
                where: {
                    isActive: true,
                },
                orderBy: {
                    nextCheckTime: 'asc',
                },
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
                    isActive: true,
                    nextCheckTime: {
                        lte: new Date(),
                    },
                },
            });
        } catch (error) {
            this.logger.error(`Error getting loans to check: ${error.message}`);
            throw error;
        }
    }
} 