export interface UserAccountData {
    totalCollateralBase: bigint;
    totalDebtBase: bigint;
    availableBorrowsBase: bigint;
    currentLiquidationThreshold: bigint;
    ltv: bigint;
    healthFactor: bigint;
}

export interface TokenInfo {
    symbol: string;
    decimals: number;
}

export interface LoanInfo {
    nextCheckTime: Date;
    healthFactor: number;
}

export interface LiquidationInfo {
    maxDebtAsset: string;
    maxDebtAmount: bigint;
    maxCollateralAsset: string;
    maxCollateralAmount: bigint;
    collateralTokenInfo: TokenInfo;
    debtTokenInfo: TokenInfo;
    debtPrice: bigint;
    collateralPrice: bigint;
    user: string;
    healthFactor: number;
    timestamp: number;
    retryCount: number;
} 