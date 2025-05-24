/*
  Warnings:

  - You are about to drop the column `borrowTime` on the `Loan` table. All the data in the column will be lost.
  - You are about to drop the column `borrowTxHash` on the `Loan` table. All the data in the column will be lost.
  - You are about to drop the column `chain` on the `Loan` table. All the data in the column will be lost.
  - You are about to drop the column `lastHealthFactor` on the `Loan` table. All the data in the column will be lost.
  - You are about to drop the column `liquidateTime` on the `Loan` table. All the data in the column will be lost.
  - You are about to drop the column `liquidateTxHash` on the `Loan` table. All the data in the column will be lost.
  - You are about to drop the column `repayTime` on the `Loan` table. All the data in the column will be lost.
  - You are about to drop the column `repayTxHash` on the `Loan` table. All the data in the column will be lost.
  - You are about to drop the column `status` on the `Loan` table. All the data in the column will be lost.
  - You are about to drop the column `userAddress` on the `Loan` table. All the data in the column will be lost.
  - Added the required column `chainName` to the `Loan` table without a default value. This is not possible if the table is not empty.
  - Added the required column `healthFactor` to the `Loan` table without a default value. This is not possible if the table is not empty.
  - Added the required column `nextCheckTime` to the `Loan` table without a default value. This is not possible if the table is not empty.
  - Added the required column `updatedAt` to the `Loan` table without a default value. This is not possible if the table is not empty.
  - Added the required column `user` to the `Loan` table without a default value. This is not possible if the table is not empty.

*/
-- RedefineTables
PRAGMA defer_foreign_keys=ON;
PRAGMA foreign_keys=OFF;
CREATE TABLE "new_Loan" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "chainName" TEXT NOT NULL,
    "user" TEXT NOT NULL,
    "healthFactor" REAL NOT NULL,
    "lastCheckTime" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "nextCheckTime" DATETIME NOT NULL,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "liquidationDiscoveredAt" DATETIME,
    "liquidationTxHash" TEXT,
    "liquidationTime" DATETIME,
    "liquidator" TEXT,
    "liquidationDelay" INTEGER,
    "totalDebt" REAL NOT NULL DEFAULT 0
);
INSERT INTO "new_Loan" ("id") SELECT "id" FROM "Loan";
DROP TABLE "Loan";
ALTER TABLE "new_Loan" RENAME TO "Loan";
CREATE INDEX "Loan_chainName_isActive_idx" ON "Loan"("chainName", "isActive");
CREATE INDEX "Loan_healthFactor_idx" ON "Loan"("healthFactor");
CREATE INDEX "Loan_nextCheckTime_idx" ON "Loan"("nextCheckTime");
CREATE UNIQUE INDEX "Loan_chainName_user_key" ON "Loan"("chainName", "user");
PRAGMA foreign_keys=ON;
PRAGMA defer_foreign_keys=OFF;
