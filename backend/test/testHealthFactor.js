import { ethers } from "ethers";

// node test/testHealthFactor.js 0x0000

// 从命令行参数获取 hexData，如果没有提供则使用默认值
const defaultHexData = "0x00000000000000000000000000000000000000000000000000000010f1c75ddb0000000000000000000000000000000000000000000000000000000df842cd9b0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000206c0000000000000000000000000000000000000000000000000000000000001f720000000000000000000000000000000000000000000000000df8a68498ec4da5";
const hexData = process.argv[2] || defaultHexData;

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