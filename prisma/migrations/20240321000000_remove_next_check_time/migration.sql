-- 删除 nextCheckTime 字段
ALTER TABLE "Loan" DROP COLUMN "nextCheckTime";

-- 删除 totalDebt 字段
ALTER TABLE "Loan" DROP COLUMN "totalDebt"; 