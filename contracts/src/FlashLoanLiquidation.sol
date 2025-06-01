// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@aave/origin-v3/contracts/interfaces/IPool.sol";
import "@aave/origin-v3/contracts/interfaces/IPoolAddressesProvider.sol";
import "@aave/origin-v3/contracts/interfaces/IPoolDataProvider.sol";
import "@aave/origin-v3/contracts/misc/flashloan/interfaces/IFlashLoanSimpleReceiver.sol";
import "./interfaces/IDex.sol";
import "forge-std/console2.sol";

contract FlashLoanLiquidation is Initializable, UUPSUpgradeable, OwnableUpgradeable, IFlashLoanSimpleReceiver {
    IPool public aave_v3_pool;
    IDex public dex;

    event Liquidation(
        address indexed user,
        address indexed asset,
        uint256 amount,
        address indexed collateralAsset,
        uint256 collateralAmount
    );

    event SwapWithAggregator(
        address indexed target,
        address indexed profitToken,
        uint256 profit,
        address indexed collateralAsset,
        uint256 collateralBalance
    );

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address _aave_v3_pool, address _dex) public initializer {
        __UUPSUpgradeable_init();
        __Ownable_init(msg.sender);
        aave_v3_pool = IPool(_aave_v3_pool);
        dex = IDex(_dex);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}

    /**
     * @notice Aave V3闪电贷回调函数
     * @param asset 借入的资产地址
     * @param amount 借入的数量
     * @param premium 闪电贷费用
     * @param _initiator 发起者地址
     * @param params 额外参数
     */
    function executeOperation(address asset, uint256 amount, uint256 premium, address _initiator, bytes calldata params)
        external
        override
        returns (bool)
    {
        if (msg.sender != address(aave_v3_pool)) revert("Caller must be aave v3 pool");
        if (IERC20(asset).balanceOf(address(this)) < amount) revert("Insufficient balance to liquidate debt");

        // remove warning
        _initiator;

        // 解码参数
        (address collateralAsset, address user, bytes memory data) = abi.decode(params, (address, address, bytes));

        // 执行清算
        IERC20(asset).approve(address(aave_v3_pool), amount);
        // debtToCover parameter can be set to  uint(-1) and the protocol will proceed with the highest possible liquidation allowed by the close factor.
        aave_v3_pool.liquidationCall(collateralAsset, asset, user, type(uint256).max, false /* bool receiveAToken */ );

        // 获取抵押品数量
        uint256 collateralAmount = IERC20(collateralAsset).balanceOf(address(this));
        console2.log("collateralAmount", collateralAmount);
        uint256 amountToRepay = amount + premium;
        if (data.length > 0) {
            _swapWithAggregator(collateralAsset, collateralAmount, data);
        } else {
            IERC20(collateralAsset).approve(address(dex), collateralAmount);
            dex.swap(collateralAsset, asset, collateralAmount, amountToRepay, owner());
        }

        // 用于偿还闪电贷的数量
        if (IERC20(asset).balanceOf(address(this)) < amountToRepay) revert("Insufficient balance to repay flash loan");
        IERC20(asset).approve(address(aave_v3_pool), amountToRepay);

        emit Liquidation(user, asset, amount, collateralAsset, collateralAmount);
        return true;
    }

    function _swapWithAggregator(address collateralAsset, uint256 collateralAmount, bytes memory params) internal {
        (address usdc, address target, bytes memory data) = abi.decode(params, (address, address, bytes));

        // 记录开始时间
        uint256 startTime = block.timestamp;
        // 设置超时时间为 5 分钟
        uint256 timeout = 5 minutes;

        // 获取 usdc 余额
        uint256 usdcBalanceBefore = IERC20(usdc).balanceOf(address(this));

        // 调用 aggregator 合约前检查是否超时
        if (block.timestamp > startTime + timeout) revert("Operation timeout");

        // 调用 aggregator 合约
        IERC20(collateralAsset).approve(target, collateralAmount);
        (bool success, bytes memory result) = target.call(data);
        if (!success) {
            if (result.length > 0) {
                assembly {
                    let result_size := mload(result)
                    revert(add(result, 32), result_size)
                }
            } else {
                revert("swap with aggregator failed with empty result");
            }
        }

        // 检查是否超时
        if (block.timestamp > startTime + timeout) revert("Operation timeout");

        // check profit
        uint256 usdcBalanceAfter = IERC20(usdc).balanceOf(address(this));
        uint256 profit = usdcBalanceAfter - usdcBalanceBefore;
        uint256 collateralBalance = IERC20(collateralAsset).balanceOf(address(this));
        if (profit < 1e5) revert("got usdc is less than 0.1 USDC"); // 0.1USDC = 1e5

        // transfer to owner
        IERC20(usdc).transfer(owner(), usdcBalanceAfter);
        IERC20(collateralAsset).transfer(owner(), collateralBalance);

        emit SwapWithAggregator(target, usdc, profit, collateralAsset, collateralBalance);
    }

    /**
     * @notice 提取合约中的代币
     * @param token 代币地址
     * @param amount 数量
     */
    function withdrawToken(address token, uint256 amount) external onlyOwner {
        IERC20(token).transfer(owner(), amount);
    }

    // 实现 IFlashLoanSimpleReceiver 接口的 ADDRESSES_PROVIDER 和 POOL
    function ADDRESSES_PROVIDER() external view override returns (IPoolAddressesProvider) {
        try aave_v3_pool.ADDRESSES_PROVIDER() returns (IPoolAddressesProvider provider) {
            return provider;
        } catch {
            return IPoolAddressesProvider(address(0));
        }
    }

    function POOL() external view override returns (IPool) {
        return aave_v3_pool;
    }

    /**
     * @notice 执行闪电贷清算，使用Aggregator
     * @param collateralAsset 抵押品资产地址
     * @param debtAsset 债务资产地址
     * @param user 用户地址
     * @param data aggregator 额外参数
     */
    function executeLiquidation(address collateralAsset, address debtAsset, address user, bytes calldata data)
        external
        onlyOwner
    {
        // 先检查用户的健康因子是否小于 1，小于 1 再执行闪电贷进行清算
        (,,,,, uint256 healthFactor) = aave_v3_pool.getUserAccountData(user);
        if (healthFactor >= 1e18) revert("Health factor is greater than 1");

        // 实时获取债务数量，链下计算的不准确
        (, uint256 currentStableDebt, uint256 currentVariableDebt,,,,,,) =
            IPoolDataProvider(this.ADDRESSES_PROVIDER().getPoolDataProvider()).getUserReserveData(debtAsset, user);
        uint256 totalDebtAmount = currentStableDebt + currentVariableDebt;

        bytes memory params = abi.encode(collateralAsset, user, data);
        aave_v3_pool.flashLoanSimple(address(this), debtAsset, totalDebtAmount, params, 0);
    }
}
