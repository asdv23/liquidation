// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import "@uniswap/v3-core/contracts/interfaces/IUniswapV3Factory.sol";
import "@uniswap/v3-core/contracts/interfaces/IUniswapV3Pool.sol";
import "@uniswap/v3-periphery/contracts/interfaces/ISwapRouter.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "../interfaces/IDex.sol";

contract UniswapV3Dex is IDex {
    ISwapRouter public immutable swapRouter;
    IUniswapV3Factory public immutable factory;
    uint24 public constant DEFAULT_FEE_TIER = 3000; // 0.3%
    uint256 public constant DEFAULT_SLIPPAGE = 500; // 5% = 500/10000

    constructor(address _swapRouter, address _factory) {
        swapRouter = ISwapRouter(_swapRouter);
        factory = IUniswapV3Factory(_factory);
    }

    function name() external pure override returns (string memory) {
        return "UniswapV3";
    }

    function supportsPair(address tokenIn, address tokenOut) external view override returns (bool) {
        address pool = factory.getPool(tokenIn, tokenOut, DEFAULT_FEE_TIER);
        return pool != address(0);
    }

    function getAmountsOut(address tokenIn, address tokenOut, uint256 amountIn)
        external
        view
        override
        returns (uint256 amountOut)
    {
        address pool = factory.getPool(tokenIn, tokenOut, DEFAULT_FEE_TIER);
        require(pool != address(0), "Pool does not exist");

        IUniswapV3Pool poolContract = IUniswapV3Pool(pool);
        (uint160 sqrtPriceX96,,,,,,) = poolContract.slot0();

        // 计算预期输出金额
        // 注意：这是一个简化的计算，实际应用中需要考虑更多因素
        uint256 price = uint256(sqrtPriceX96) * uint256(sqrtPriceX96) * 1e18 / (1 << 192);
        amountOut = amountIn * price / 1e18;
    }

    function swapExactTokensForTokens(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 amountOutMin,
        address to
    ) external override returns (uint256 amountOut) {
        // 如果未指定最小输出金额，使用默认滑点计算
        if (amountOutMin == 0) {
            uint256 expectedAmountOut = this.getAmountsOut(tokenIn, tokenOut, amountIn);
            amountOutMin = expectedAmountOut * (10000 - DEFAULT_SLIPPAGE) / 10000;
        }

        // 授权 SwapRouter 使用代币
        IERC20(tokenIn).approve(address(swapRouter), amountIn);

        // 准备兑换参数
        ISwapRouter.ExactInputSingleParams memory params = ISwapRouter.ExactInputSingleParams({
            tokenIn: tokenIn,
            tokenOut: tokenOut,
            fee: DEFAULT_FEE_TIER,
            recipient: to,
            deadline: block.timestamp + 15 minutes,
            amountIn: amountIn,
            amountOutMinimum: amountOutMin,
            sqrtPriceLimitX96: 0
        });

        // 执行兑换
        amountOut = swapRouter.exactInputSingle(params);
    }

    function swapTokensForExactTokens(
        address tokenIn,
        address tokenOut,
        uint256 amountOut,
        uint256 amountInMax,
        address to
    ) external override returns (uint256 amountIn) {
        // 如果未指定最大输入金额，使用默认滑点计算
        if (amountInMax == 0) {
            uint256 expectedAmountIn = this.getAmountsOut(tokenOut, tokenIn, amountOut);
            amountInMax = expectedAmountIn * (10000 + DEFAULT_SLIPPAGE) / 10000;
        }

        // 授权 SwapRouter 使用代币
        IERC20(tokenIn).approve(address(swapRouter), amountInMax);

        // 准备兑换参数
        ISwapRouter.ExactOutputSingleParams memory params = ISwapRouter.ExactOutputSingleParams({
            tokenIn: tokenIn,
            tokenOut: tokenOut,
            fee: DEFAULT_FEE_TIER,
            recipient: to,
            deadline: block.timestamp + 15 minutes,
            amountOut: amountOut,
            amountInMaximum: amountInMax,
            sqrtPriceLimitX96: 0
        });

        // 执行兑换
        amountIn = swapRouter.exactOutputSingle(params);
    }
}
