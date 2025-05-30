// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Script} from "forge-std/Script.sol";
import {console2} from "forge-std/console2.sol";
import {UpgradeParams} from "./Params.s.sol";
import {FlashLoanLiquidation} from "../src/FlashLoanLiquidation.sol";
import {UniswapV3Dex} from "../src/dex/UniswapV3Dex.sol";

abstract contract Upgrade is Script {
    UpgradeParams public params;

    // set values for params and unsupported
    function setUp() public virtual;

    function run() public {
        vm.startBroadcast();

        if (params.uniswapV3Dex != address(0)) {
            UniswapV3Dex uniswapV3Dex = new UniswapV3Dex();
            (bool success,) = params.uniswapV3Dex.call(
                abi.encodeWithSignature("upgradeToAndCall(address,bytes)", address(uniswapV3Dex), "")
            );
            if (!success) revert("Upgrade UniswapV3Dex failed");
            console2.log("UniswapV3Dex implementation upgraded to:", address(uniswapV3Dex));
        }

        if (params.flashLoanLiquidation != address(0)) {
            FlashLoanLiquidation flashLoanLiquidation = new FlashLoanLiquidation();
            (bool success,) = params.flashLoanLiquidation.call(
                abi.encodeWithSignature("upgradeToAndCall(address,bytes)", address(flashLoanLiquidation), "")
            );
            if (!success) revert("Upgrade FlashLoanLiquidation failed");
            console2.log("FlashLoanLiquidation implementation upgraded to:", address(flashLoanLiquidation));
        }

        vm.stopBroadcast();
    }
}
