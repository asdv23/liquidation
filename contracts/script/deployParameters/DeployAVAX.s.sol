// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import {Deploy} from "../Deploy.s.sol";
import {Upgrade} from "../Upgrade.s.sol";
import {DeployParams, UpgradeParams} from "../Params.s.sol";

contract DeployAVAX is Deploy {
    function setUp() public override {
        params = DeployParams({
            aaveV3Pool: 0x794a61358D6845594F94dc1DB02A252b5b4814aD,
            swapRouter02: 0xbb00FF08d01D300023C629E8fFfFcb65A5a578cE,
            factory: 0x740b1c1de25031C31FF4fC9A62f554A55cdC1baD,
            quoterV2: 0xbe0F5544EC67e9B3b2D979aaA43f18Fd87E6257F,
            usdc: 0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E
        });
    }
}

contract UpgradeAVAX is Upgrade {
    function setUp() public override {
        params = UpgradeParams({
            uniswapV3Dex: 0x2243F7E7CFFC505db1D94614F2aAe8274FCdA09C,
            flashLoanLiquidation: 0xE46BFAfAEDeB9CA98424860426E0093345bf2aE1
        });
    }
}
