import { OnModuleInit } from '@nestjs/common';
import { PrismaClient } from '@prisma/client';
import { ConfigService } from '@nestjs/config';
export declare class DatabaseService implements OnModuleInit {
    private readonly configService;
    private readonly logger;
    private static prisma;
    constructor(configService: ConfigService);
    onModuleInit(): Promise<void>;
    get prisma(): PrismaClient<import(".prisma/client").Prisma.PrismaClientOptions, never, import("@prisma/client/runtime/library").DefaultArgs>;
    markLiquidationDiscovered(chainName: string, user: string): Promise<{
        id: number;
        chainName: string;
        user: string;
        isActive: boolean;
        createdAt: Date;
        updatedAt: Date;
        liquidationDiscoveredAt: Date | null;
        liquidationTxHash: string | null;
        liquidationTime: Date | null;
        liquidator: string | null;
        liquidationDelay: number | null;
    }>;
    recordLiquidation(chainName: string, user: string, liquidator: string, txHash: string): Promise<{
        id: number;
        chainName: string;
        user: string;
        isActive: boolean;
        createdAt: Date;
        updatedAt: Date;
        liquidationDiscoveredAt: Date | null;
        liquidationTxHash: string | null;
        liquidationTime: Date | null;
        liquidator: string | null;
        liquidationDelay: number | null;
    }>;
    getActiveLoans(chainName?: string): Promise<{
        id: number;
        chainName: string;
        user: string;
        isActive: boolean;
        createdAt: Date;
        updatedAt: Date;
        liquidationDiscoveredAt: Date | null;
        liquidationTxHash: string | null;
        liquidationTime: Date | null;
        liquidator: string | null;
        liquidationDelay: number | null;
    }[]>;
    deactivateLoan(chainName: string, user: string): Promise<{
        id: number;
        chainName: string;
        user: string;
        isActive: boolean;
        createdAt: Date;
        updatedAt: Date;
        liquidationDiscoveredAt: Date | null;
        liquidationTxHash: string | null;
        liquidationTime: Date | null;
        liquidator: string | null;
        liquidationDelay: number | null;
    }>;
    getLoansToCheck(): Promise<{
        id: number;
        chainName: string;
        user: string;
        isActive: boolean;
        createdAt: Date;
        updatedAt: Date;
        liquidationDiscoveredAt: Date | null;
        liquidationTxHash: string | null;
        liquidationTime: Date | null;
        liquidator: string | null;
        liquidationDelay: number | null;
    }[]>;
    getToken(chainName: string, address: string): Promise<{
        symbol: string;
        id: number;
        chainName: string;
        createdAt: Date;
        updatedAt: Date;
        address: string;
        decimals: number;
    }>;
    saveToken(chainName: string, address: string, symbol: string, decimals: number): Promise<{
        symbol: string;
        id: number;
        chainName: string;
        createdAt: Date;
        updatedAt: Date;
        address: string;
        decimals: number;
    }>;
    getAllTokens(chainName?: string): Promise<{
        symbol: string;
        id: number;
        chainName: string;
        createdAt: Date;
        updatedAt: Date;
        address: string;
        decimals: number;
    }[]>;
    createOrUpdateLoan(chainName: string, user: string): Promise<void>;
}
