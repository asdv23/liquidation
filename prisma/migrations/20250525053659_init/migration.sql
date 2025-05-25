-- CreateTable
CREATE TABLE "Loan" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "chainName" TEXT NOT NULL,
    "user" TEXT NOT NULL,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "liquidationDiscoveredAt" DATETIME,
    "liquidationTxHash" TEXT,
    "liquidationTime" DATETIME,
    "liquidator" TEXT,
    "liquidationDelay" INTEGER
);

-- CreateTable
CREATE TABLE "Token" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "chainName" TEXT NOT NULL,
    "address" TEXT NOT NULL,
    "symbol" TEXT NOT NULL,
    "decimals" INTEGER NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateIndex
CREATE INDEX "Loan_chainName_isActive_idx" ON "Loan"("chainName", "isActive");

-- CreateIndex
CREATE UNIQUE INDEX "Loan_chainName_user_key" ON "Loan"("chainName", "user");

-- CreateIndex
CREATE INDEX "Token_chainName_idx" ON "Token"("chainName");

-- CreateIndex
CREATE UNIQUE INDEX "Token_chainName_address_key" ON "Token"("chainName", "address");
