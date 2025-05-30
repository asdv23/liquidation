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
const path = require("path");
const config_1 = require("@nestjs/config");
let DatabaseService = DatabaseService_1 = class DatabaseService {
    constructor(configService) {
        this.configService = configService;
        this.logger = new common_1.Logger(DatabaseService_1.name);
        if (!DatabaseService_1.prisma) {
            const dbPath = this.configService.get('DATABASE_URL', path.join(process.cwd(), 'prisma', 'dev.db'));
            this.logger.log(`Using SQLite database at: ${dbPath}`);
            DatabaseService_1.prisma = new client_1.PrismaClient({
                datasources: {
                    db: {
                        url: `file:${dbPath}`
                    }
                }
            });
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
                        user: user.toLowerCase(),
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
                        user: user.toLowerCase(),
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
                where: Object.assign({ isActive: true }, (chainName ? { chainName } : {}))
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
                        user: user.toLowerCase(),
                    },
                },
                data: {
                    isActive: false,
                    updatedAt: new Date(),
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
                    isActive: true
                }
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
    async createOrUpdateLoan(chainName, user) {
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
                    isActive: true
                },
            });
        }
        catch (error) {
            this.logger.error(`Error creating/updating loan: ${error.message}`);
            throw error;
        }
    }
    async updateLoanHealthFactor(chainName, user, healthFactor, nextCheckTime) {
        try {
            await this.prisma.loan.update({
                where: {
                    chainName_user: {
                        chainName,
                        user: user.toLowerCase(),
                    },
                },
                data: {
                    healthFactor,
                    nextCheckTime,
                    updatedAt: new Date(),
                },
            });
        }
        catch (error) {
            this.logger.error(`Error updating loan health factor: ${error.message}`);
            throw error;
        }
    }
};
exports.DatabaseService = DatabaseService;
exports.DatabaseService = DatabaseService = DatabaseService_1 = __decorate([
    (0, common_1.Injectable)(),
    __metadata("design:paramtypes", [config_1.ConfigService])
], DatabaseService);
//# sourceMappingURL=database.service.js.map