// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {Script} from "forge-std/Script.sol";
import {FlashLoanLiquidation} from "../src/FlashLoanLiquidation.sol";

contract UpgradeScript is Script {
    function run() public {
        // 从环境变量获取私钥
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");

        // 获取当前合约地址
        address currentContract = vm.envAddress("FLASH_LOAN_LIQUIDATION");

        // 开始广播交易
        vm.startBroadcast(deployerPrivateKey);

        // 部署新合约
        FlashLoanLiquidation newContract = new FlashLoanLiquidation();

        // 调用当前合约的升级函数
        // 注意：这里假设当前合约有 upgradeTo 函数，如果没有，需要根据实际情况修改
        (bool success,) = currentContract.call(abi.encodeWithSignature("upgradeTo(address)", address(newContract)));
        require(success, "Upgrade failed");

        vm.stopBroadcast();

        // 输出新合约地址
        console2.log("New contract deployed at:", address(newContract));
    }
}
