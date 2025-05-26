// SPDX-License-Identifier: MIT
pragma solidity ^0.8.10;

import "forge-std/Script.sol";
import "../src/FlashLoanLiquidation.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract DeployFlashLoanLiquidation is Script {
    function run() external {
        uint256 deployerPrivateKey = vm.envUint("PRIVATE_KEY");
        vm.startBroadcast(deployerPrivateKey);

        // 部署实现合约
        FlashLoanLiquidation implementation = new FlashLoanLiquidation();

        // 部署代理合约
        ERC1967Proxy proxy = new ERC1967Proxy(
            address(implementation),
            abi.encodeWithSelector(
                FlashLoanLiquidation.initialize.selector,
                address(0) // 替换为实际的 Aave Pool 地址
            )
        );

        FlashLoanLiquidation liquidation = FlashLoanLiquidation(address(proxy));

        liquidation.owner();

        vm.stopBroadcast();
    }
}
