// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import {Script} from "forge-std/Script.sol";
import {console2} from "forge-std/console2.sol";
import {FlashLoanLiquidation} from "../src/FlashLoanLiquidation.sol";
import {UniswapV3Dex} from "../src/dex/UniswapV3Dex.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

import {DeployParams} from "./Params.s.sol";

abstract contract Deploy is Script {
    DeployParams public params;

    // set values for params and unsupported
    function setUp() public virtual;

    function run() public {
        vm.startBroadcast();

        // 部署 UniswapV3Dex
        UniswapV3Dex uniswapV3DexImpl = new UniswapV3Dex();
        bytes memory uniswapV3DexInitData = abi.encodeWithSelector(
            UniswapV3Dex.initialize.selector,
            address(params.swapRouter02), // Uniswap SwapRouter02
            address(params.factory), // Uniswap V3 Factory
            address(params.quoterV2), // Uniswap V3 QuoterV2
            address(params.usdc) // USDC
        );
        UniswapV3Dex uniswapV3Dex =
            UniswapV3Dex(address(new ERC1967Proxy(address(uniswapV3DexImpl), uniswapV3DexInitData)));

        // 部署 FlashLoanLiquidation
        FlashLoanLiquidation flashLoanLiquidationImpl = new FlashLoanLiquidation();
        bytes memory flashLoanLiquidationInitData = abi.encodeWithSelector(
            FlashLoanLiquidation.initialize.selector,
            address(params.aaveV3Pool), // Aave V3 Pool
            address(uniswapV3Dex) // UniswapV3Dex
        );
        FlashLoanLiquidation flashLoanLiquidation = FlashLoanLiquidation(
            address(new ERC1967Proxy(address(flashLoanLiquidationImpl), flashLoanLiquidationInitData))
        );

        console2.log("uniswapV3Dex", address(uniswapV3Dex));
        console2.log("flashLoanLiquidation", address(flashLoanLiquidation));

        vm.stopBroadcast();
    }
}
