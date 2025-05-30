// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import {Deploy} from "../Deploy.s.sol";
import {Upgrade} from "../Upgrade.s.sol";
import {DeployParams, UpgradeParams} from "../Params.s.sol";

contract DeployOP is Deploy {
    function setUp() public override {
        params = DeployParams({
            aaveV3Pool: 0x794a61358D6845594F94dc1DB02A252b5b4814aD,
            swapRouter02: 0x68b3465833fb72A70ecDF485E0e4C7bD8665Fc45,
            factory: 0x1F98431c8aD98523631AE4a59f267346ea31F984,
            quoterV2: 0x61fFE014bA17989E743c5F6cB21bF9697530B21e,
            usdc: 0x0b2C639c533813f4Aa9D7837CAf62653d097Ff85
        });
    }
}

contract UpgradeOP is Upgrade {
    function setUp() public override {
        params = UpgradeParams({
            uniswapV3Dex: 0x2243F7E7CFFC505db1D94614F2aAe8274FCdA09C,
            flashLoanLiquidation: 0xE46BFAfAEDeB9CA98424860426E0093345bf2aE1
        });
    }
}
