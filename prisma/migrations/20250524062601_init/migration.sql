-- CreateTable
CREATE TABLE "Loan" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "chain" TEXT NOT NULL,
    "userAddress" TEXT NOT NULL,
    "borrowTxHash" TEXT NOT NULL,
    "borrowTime" DATETIME NOT NULL,
    "repayTxHash" TEXT,
    "repayTime" DATETIME,
    "liquidateTxHash" TEXT,
    "liquidateTime" DATETIME,
    "lastHealthFactor" DECIMAL,
    "status" TEXT NOT NULL
);

-- CreateTable
CREATE TABLE "LoanHealthHistory" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "loanId" INTEGER NOT NULL,
    "checkTime" DATETIME NOT NULL,
    "healthFactor" DECIMAL NOT NULL,
    CONSTRAINT "LoanHealthHistory_loanId_fkey" FOREIGN KEY ("loanId") REFERENCES "Loan" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);
