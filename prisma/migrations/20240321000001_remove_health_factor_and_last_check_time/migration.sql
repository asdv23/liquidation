-- 删除 healthFactor 字段
ALTER TABLE "Loan" DROP COLUMN "healthFactor";

-- 删除 lastCheckTime 字段
ALTER TABLE "Loan" DROP COLUMN "lastCheckTime"; 