// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

struct DeployParams {
    // FlashLoanLiquidation
    address aaveV3Pool;
    // UniswapV3Dex
    address swapRouter02;
    address factory;
    address quoterV2;
    address usdc;
}

struct UpgradeParams {
    // UniswapV3Dex
    address uniswapV3Dex;
    // FlashLoanLiquidation
    address flashLoanLiquidation;
}
