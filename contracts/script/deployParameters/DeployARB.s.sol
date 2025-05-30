// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import {Deploy} from "../Deploy.s.sol";
import {Upgrade} from "../Upgrade.s.sol";
import {DeployParams, UpgradeParams} from "../Params.s.sol";

contract DeployARB is Deploy {
    function setUp() public override {
        params = DeployParams({
            aaveV3Pool: 0x794a61358D6845594F94dc1DB02A252b5b4814aD,
            swapRouter02: 0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45,
            factory: 0x1F98431c8aD98523631AE4a59f267346ea31F984,
            quoterV2: 0x61fFE014bA17989E743c5F6cB21bF9697530B21e,
            usdc: 0xaf88d065e77c8cC2239327C5EDb3A432268e5831
        });
    }
}

contract UpgradeARB is Upgrade {
    function setUp() public override {
        params = UpgradeParams({
            uniswapV3Dex: 0x883Fe2FD1B1591764e30607CACFA7ECCc82FF55C,
            flashLoanLiquidation: 0x09Ab4549E340595ec824F28d291c7e1C91FaF68A
        });
    }
}
