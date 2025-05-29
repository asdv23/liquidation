// SPDX-License-Identifier: MIT
pragma solidity ^0.8.10;

import "forge-std/Test.sol";
import "../src/FlashLoanLiquidation.sol";
import "../src/dex/UniswapV3Dex.sol";
import "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import "@aave/origin-v3/contracts/interfaces/IPool.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract FlashLoanLiquidationTest is Test {
    FlashLoanLiquidation public implementation;
    FlashLoanLiquidation public liquidation;
    UniswapV3Dex public dex;
    address public pool;
    address public usdc;
    address public collateralAsset;
    address public debtAsset;
    address public user;
    address public swapRouter;

    function setUp() public {
        // 部署实现合约
        implementation = new FlashLoanLiquidation();

        // 部署 DEX
        swapRouter = address(0x111);
        UniswapV3Dex dexImpl = new UniswapV3Dex();
        bytes memory dexInitData = abi.encodeWithSelector(UniswapV3Dex.initialize.selector, swapRouter);
        dex = UniswapV3Dex(address(new ERC1967Proxy(address(dexImpl), dexInitData)));

        // 设置测试环境
        pool = address(0x123); // 模拟 Aave Pool
        usdc = address(0x333); // 模拟 USDC
        collateralAsset = address(0x456); // 模拟抵押资产
        debtAsset = address(0x789); // 模拟债务资产
        user = address(0xabc); // 模拟被清算用户

        // 部署代理合约
        ERC1967Proxy proxy = new ERC1967Proxy(
            address(implementation),
            abi.encodeWithSelector(FlashLoanLiquidation.initialize.selector, pool, address(dex), usdc)
        );

        liquidation = FlashLoanLiquidation(address(proxy));
    }

    function testInitialize() public view {
        assertEq(address(liquidation.aave_v3_pool()), pool);
        assertEq(address(liquidation.dex()), address(dex));
        assertEq(liquidation.usdc(), usdc);
    }

    function testExecuteLiquidation() public {
        // 模拟健康因子检查
        vm.mockCall(
            pool,
            abi.encodeWithSelector(IPool.getUserAccountData.selector, user),
            abi.encode(0, 0, 0, 0, 0, 0.5e18) // 健康因子为0.5
        );

        // 模拟闪电贷回调
        vm.mockCall(pool, abi.encodeWithSelector(IPool.flashLoanSimple.selector), abi.encode());

        // 执行清算
        liquidation.executeLiquidation(
            collateralAsset,
            debtAsset,
            user,
            1000e18 // 清算数量
        );
    }

    function testExecuteOperation() public {
        // 模拟清算调用
        vm.mockCall(pool, abi.encodeWithSelector(IPool.liquidationCall.selector), abi.encode());

        // 模拟代币余额
        vm.mockCall(
            collateralAsset,
            abi.encodeWithSelector(IERC20.balanceOf.selector, address(liquidation)),
            abi.encode(1000e18)
        );
        // mock 闪电贷资产余额，确保足够偿还
        vm.mockCall(
            debtAsset, abi.encodeWithSelector(IERC20.balanceOf.selector, address(liquidation)), abi.encode(1010e18)
        );

        // 模拟代币授权
        vm.mockCall(collateralAsset, abi.encodeWithSelector(IERC20.approve.selector), abi.encode(true));
        vm.mockCall(debtAsset, abi.encodeWithSelector(IERC20.approve.selector), abi.encode(true));

        // 模拟 DEX 兑换
        vm.mockCall(address(dex), abi.encodeWithSelector(IDex.swapTokensForExactTokens.selector), abi.encode(500e18));
        vm.mockCall(address(dex), abi.encodeWithSelector(IDex.swapExactTokensForTokens.selector), abi.encode(0));

        // 执行闪电贷回调，伪造 msg.sender 为 pool
        bytes memory params = abi.encode(collateralAsset, user, true);
        vm.prank(pool);
        liquidation.executeOperation(
            debtAsset,
            1000e18,
            10e18, // premium
            address(this),
            params
        );
    }

    function testWithdrawToken() public {
        // 模拟代币余额
        vm.mockCall(
            collateralAsset,
            abi.encodeWithSelector(IERC20.balanceOf.selector, address(liquidation)),
            abi.encode(1000e18)
        );

        // 模拟代币转账
        vm.mockCall(collateralAsset, abi.encodeWithSelector(IERC20.transfer.selector), abi.encode(true));

        // 执行提取
        liquidation.withdrawToken(collateralAsset, 1000e18);
    }

    function test_RevertWhen_NotOwner() public {
        // 切换到非所有者地址
        vm.prank(address(0x999));

        // 执行提取（应该失败）
        vm.expectRevert(abi.encodeWithSignature("OwnableUnauthorizedAccount(address)", address(0x999)));
        liquidation.withdrawToken(collateralAsset, 1000e18);
    }
}
