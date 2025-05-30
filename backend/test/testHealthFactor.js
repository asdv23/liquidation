import { ethers } from "ethers";

// 输入的十六进制数据
const hexData = "0x000000000000000000000000000000000000000000000000000000036263d12a00000000000000000000000000000000000000000000000000000002d07e88100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000206c0000000000000000000000000000000000000000000000000000000000001f400000000000000000000000000000000000000000000000000dd9db455390454b";

// 使用 AbiCoder 解码
const abiCoder = new ethers.AbiCoder();
const decoded = abiCoder.decode(
  ["uint256", "uint256", "uint256", "uint256", "uint256", "uint256"],
  hexData
);

// 提取并格式化结果
const [
  totalCollateralBase,
  totalDebtBase,
  availableBorrowsBase,
  currentLiquidationThreshold,
  ltv,
  healthFactor
] = decoded;

// 假设 base 单位是 8 位小数（常见于 Aave 协议）
console.log("totalCollateralBase:", ethers.formatUnits(totalCollateralBase, 8), "USD");
console.log("totalDebtBase:", ethers.formatUnits(totalDebtBase, 8), "USD");
console.log("availableBorrowsBase:", ethers.formatUnits(availableBorrowsBase, 8), "USD");
console.log("currentLiquidationThreshold:", currentLiquidationThreshold.toString(), "bps");
console.log("ltv:", ltv.toString(), "bps");
console.log("healthFactor:", ethers.formatUnits(healthFactor, 18));