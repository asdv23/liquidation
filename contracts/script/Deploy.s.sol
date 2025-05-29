// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import {Script} from "forge-std/Script.sol";
import {console2} from "forge-std/console2.sol";
import {FlashLoanLiquidation} from "../src/FlashLoanLiquidation.sol";
import {UniswapV3Dex} from "../src/dex/UniswapV3Dex.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract DeployLiquidation is Script {
    // set values for params and unsupported
    function setUp() public virtual;

    function run() public {
        vm.startBroadcast();

        // 部署 UniswapV3Dex
        UniswapV3Dex uniswapV3DexImpl = new UniswapV3Dex();
        bytes memory uniswapV3DexInitData = abi.encodeWithSelector(
            UniswapV3Dex.initialize.selector,
            address(0x2626664c2603336E57B271c5C0b26F421741e481) // Uniswap SwapRouter02
        );
        UniswapV3Dex uniswapV3Dex =
            UniswapV3Dex(address(new ERC1967Proxy(address(uniswapV3DexImpl), uniswapV3DexInitData)));

        // 部署 FlashLoanLiquidation
        FlashLoanLiquidation flashLoanLiquidationImpl = new FlashLoanLiquidation();
        bytes memory flashLoanLiquidationInitData = abi.encodeWithSelector(
            FlashLoanLiquidation.initialize.selector,
            address(0xA238Dd80C259a72e81d7e4664a9801593F98d1c5), // Aave V3 Pool
            address(uniswapV3Dex), // UniswapV3Dex
            address(0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913) // USDC
        );
        FlashLoanLiquidation flashLoanLiquidation = FlashLoanLiquidation(
            address(new ERC1967Proxy(address(flashLoanLiquidationImpl), flashLoanLiquidationInitData))
        );

        vm.stopBroadcast();

        console2.log("UniswapV3Dex deployed to:", address(uniswapV3Dex));
        console2.log("FlashLoanLiquidation deployed to:", address(flashLoanLiquidation));
    }
}
