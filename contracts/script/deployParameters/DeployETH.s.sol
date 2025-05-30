// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import {Deploy} from "../Deploy.s.sol";
import {Upgrade} from "../Upgrade.s.sol";
import {DeployParams, UpgradeParams} from "../Params.s.sol";

contract DeployETH is Deploy {
    function setUp() public override {
        params = DeployParams({
            aaveV3Pool: 0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2,
            swapRouter02: 0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45,
            factory: 0x1F98431c8aD98523631AE4a59f267346ea31F984,
            quoterV2: 0x61fFE014bA17989E743c5F6cB21bF9697530B21e,
            usdc: 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
        });
    }
}

contract UpgradeBase is Upgrade {
    function setUp() public override {
        params = UpgradeParams({
            uniswapV3Dex: 0x285FcE70D05e671d92db253B809Ed5FEc19cE7ac,
            flashLoanLiquidation: 0x44204C331b29E63053aEb45D4c1794Fc3B7a4287
        });
    }
}
