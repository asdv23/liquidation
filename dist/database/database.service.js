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
var DatabaseService_1;
Object.defineProperty(exports, "__esModule", { value: true });
exports.DatabaseService = void 0;
const common_1 = require("@nestjs/common");
const client_1 = require("@prisma/client");
let DatabaseService = DatabaseService_1 = class DatabaseService {
    constructor() {
        this.logger = new common_1.Logger(DatabaseService_1.name);
        if (!DatabaseService_1.prisma) {
            DatabaseService_1.prisma = new client_1.PrismaClient();
        }
    }
    async onModuleInit() {
        try {
            await DatabaseService_1.prisma.$connect();
            this.logger.log('Database connection established');
        }
        catch (error) {
            this.logger.error(`Failed to connect to database: ${error.message}`);
            throw error;
        }
    }
    get prisma() {
        return DatabaseService_1.prisma;
    }
    async updateLoanHealthFactor(chainName, user, healthFactor, nextCheckTime, totalDebt) {
        await this.prisma.loan.updateMany({
            where: {
                chainName,
                user: user.toLowerCase(),
                isActive: true
            },
            data: {
                healthFactor,
                lastCheckTime: new Date(),
                nextCheckTime,
                totalDebt
            }
        });
    }
    async markLiquidationDiscovered(chainName, user) {
        try {
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
        }
        catch (error) {
            this.logger.error(`Error marking liquidation discovered: ${error.message}`);
            throw error;
        }
    }
    async recordLiquidation(chainName, user, liquidator, txHash) {
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
        }
        catch (error) {
            this.logger.error(`Error recording liquidation: ${error.message}`);
            throw error;
        }
    }
    async getActiveLoans(chainName) {
        try {
            return await this.prisma.loan.findMany({
                where: Object.assign({ isActive: true }, (chainName ? { chainName } : {})),
                orderBy: {
                    nextCheckTime: 'asc',
                },
            });
        }
        catch (error) {
            this.logger.error(`Error getting active loans: ${error.message}`);
            throw error;
        }
    }
    async deactivateLoan(chainName, user) {
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
        }
        catch (error) {
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
        }
        catch (error) {
            this.logger.error(`Error getting loans to check: ${error.message}`);
            throw error;
        }
    }
    async getToken(chainName, address) {
        try {
            return await this.prisma.token.findUnique({
                where: {
                    chainName_address: {
                        chainName,
                        address: address.toLowerCase(),
                    },
                },
            });
        }
        catch (error) {
            this.logger.error(`Error getting token: ${error.message}`);
            throw error;
        }
    }
    async saveToken(chainName, address, symbol, decimals) {
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
        }
        catch (error) {
            this.logger.error(`Error saving token: ${error.message}`);
            throw error;
        }
    }
    async getAllTokens(chainName) {
        try {
            return await this.prisma.token.findMany({
                where: chainName ? { chainName } : undefined,
            });
        }
        catch (error) {
            this.logger.error(`Error getting all tokens: ${error.message}`);
            throw error;
        }
    }
};
exports.DatabaseService = DatabaseService;
exports.DatabaseService = DatabaseService = DatabaseService_1 = __decorate([
    (0, common_1.Injectable)(),
    __metadata("design:paramtypes", [])
], DatabaseService);
//# sourceMappingURL=database.service.js.map