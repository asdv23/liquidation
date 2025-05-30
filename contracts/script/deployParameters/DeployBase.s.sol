// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import {Deploy} from "../Deploy.s.sol";
import {Upgrade} from "../Upgrade.s.sol";
import {DeployParams, UpgradeParams} from "../Params.s.sol";

contract DeployBase is Deploy {
    function setUp() public override {
        params = DeployParams({
            aaveV3Pool: 0xA238Dd80C259a72e81d7e4664a9801593F98d1c5,
            swapRouter02: 0x2626664c2603336E57B271c5C0b26F421741e481,
            factory: 0x33128a8fC17869897dcE68Ed026d694621f6FDfD,
            quoterV2: 0x3d4e44Eb1374240CE5F1B871ab261CD16335B76a,
            usdc: 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913
        });
    }
}

contract UpgradeBase is Upgrade {
    function setUp() public override {
        params = UpgradeParams({
            uniswapV3Dex: 0x2243F7E7CFFC505db1D94614F2aAe8274FCdA09C,
            flashLoanLiquidation: 0xE46BFAfAEDeB9CA98424860426E0093345bf2aE1
        });
    }
}
