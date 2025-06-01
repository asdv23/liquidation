// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@uniswap/v3-core/contracts/interfaces/IUniswapV3Pool.sol";
import "@uniswap/v3-core/contracts/interfaces/IUniswapV3Factory.sol";
import "@uniswap/v3-periphery/contracts/interfaces/IQuoterV2.sol";
import "../interfaces/IDex.sol";
import "../interfaces/IV3SwapRouter.sol";

contract UniswapV3Dex is Initializable, UUPSUpgradeable, OwnableUpgradeable, IDex {
    IV3SwapRouter public swapRouter;
    IUniswapV3Factory public factory;
    IQuoterV2 public quoter;
    address public usdc;
    uint24[] public POOL_FEES;

    event Swap(address usdc, uint256 amountOut, address receiver);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address _swapRouter, address _factory, address _quoter, address _usdc) public initializer {
        __UUPSUpgradeable_init();
        __Ownable_init(msg.sender);
        swapRouter = IV3SwapRouter(_swapRouter);
        factory = IUniswapV3Factory(_factory);
        quoter = IQuoterV2(_quoter);
        usdc = _usdc;

        // 初始化 POOL_FEES 数组
        POOL_FEES = [100, 500, 3000, 10000];
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}

    function name() external pure override returns (string memory) {
        return "UniswapV3Dex";
    }

    function swap(address tokenIn, address tokenOut, uint256 amountIn, uint256 amountOut, address receiver)
        external
        override
    {
        IERC20(tokenIn).transferFrom(msg.sender, address(this), amountIn);
        IERC20(tokenIn).approve(address(swapRouter), amountIn);

        uint256 actualAmountIn = swapTokensForExactTokens(tokenIn, tokenOut, amountOut, amountIn, msg.sender);
        if (actualAmountIn >= amountIn) revert("no usdc to get");

        // 剩余的 tokenIn 换成 usdc
        uint256 profit = 0;
        if (tokenIn != usdc) {
            profit = swapExactTokensForTokens(tokenIn, usdc, amountIn - actualAmountIn, 0, receiver);
            if (profit < 1e6) revert("got usdc is less than 1U"); // 1U = 1e6
        } else {
            // usdc no need to swap
            profit = amountIn - actualAmountIn;
            if (profit < 1e6) revert("remaining usdc is less than 1U"); // 1U = 1e6
            IERC20(usdc).transfer(receiver, profit);
        }

        emit Swap(usdc, profit, receiver);
    }

    // 寻找最佳费率池（ExactInputSingle）
    function findBestFeeExactInput(address tokenIn, address tokenOut, uint256 amountIn)
        internal
        returns (uint24 bestFee, uint256 maxAmountOut)
    {
        maxAmountOut = 0;

        uint24 selectedFee = 0;
        for (uint256 i = 0; i < POOL_FEES.length; i++) {
            uint24 fee = POOL_FEES[i];
            if (factory.getPool(tokenIn, tokenOut, fee) == address(0)) continue;

            try quoter.quoteExactInputSingle(
                IQuoterV2.QuoteExactInputSingleParams({
                    tokenIn: tokenIn,
                    tokenOut: tokenOut,
                    amountIn: amountIn,
                    fee: fee,
                    sqrtPriceLimitX96: 0 // 不设置价格限制
                })
            ) returns (uint256 amountOut, uint160, uint32, uint256) {
                if (amountOut > maxAmountOut) {
                    maxAmountOut = amountOut;
                    selectedFee = fee;
                }
            } catch {
                // 池子不存在或流动性不足，跳过
                continue;
            }
        }

        if (maxAmountOut == 0) revert("No valid input pool found");
        return (selectedFee, maxAmountOut);
    }

    // 寻找最佳费率池（ExactOutputSingle）
    function findBestFeeExactOutput(address tokenIn, address tokenOut, uint256 amountOut)
        internal
        returns (uint24 bestFee, uint256 minAmountIn)
    {
        minAmountIn = type(uint256).max;

        uint24 selectedFee = 0;
        for (uint256 i = 0; i < POOL_FEES.length; i++) {
            uint24 fee = POOL_FEES[i];
            if (factory.getPool(tokenIn, tokenOut, fee) == address(0)) continue;

            try quoter.quoteExactOutputSingle(
                IQuoterV2.QuoteExactOutputSingleParams({
                    tokenIn: tokenIn,
                    tokenOut: tokenOut,
                    amount: amountOut,
                    fee: fee,
                    sqrtPriceLimitX96: 0 // 不设置价格限制
                })
            ) returns (uint256 amountIn, uint160, uint32, uint256) {
                if (amountIn < minAmountIn) {
                    minAmountIn = amountIn;
                    selectedFee = fee;
                }
            } catch {
                continue;
            }
        }

        if (minAmountIn == type(uint256).max) revert("No valid output pool found");
        return (selectedFee, minAmountIn);
    }

    function swapExactTokensForTokens(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 amountOutMin,
        address recipient
    ) internal returns (uint256) {
        // 查找最优池子
        (uint24 bestFee, uint256 expectedAmountOut) = findBestFeeExactInput(tokenIn, tokenOut, amountIn);
        if (expectedAmountOut < amountOutMin) revert("Insufficient output amount");

        // 使用最优池子
        IV3SwapRouter.ExactInputSingleParams memory params = IV3SwapRouter.ExactInputSingleParams({
            tokenIn: tokenIn,
            tokenOut: tokenOut,
            fee: bestFee,
            recipient: recipient,
            amountIn: amountIn,
            amountOutMinimum: amountOutMin,
            sqrtPriceLimitX96: 0
        });

        return swapRouter.exactInputSingle(params);
    }

    function swapTokensForExactTokens(
        address tokenIn,
        address tokenOut,
        uint256 amountOut,
        uint256 amountInMax,
        address recipient
    ) internal returns (uint256) {
        // 查找最优池子
        (uint24 bestFee, uint256 estimatedAmountIn) = findBestFeeExactOutput(tokenIn, tokenOut, amountOut);
        if (estimatedAmountIn > amountInMax) revert("Insufficient input amount");

        // 使用最优池子
        IV3SwapRouter.ExactOutputSingleParams memory params = IV3SwapRouter.ExactOutputSingleParams({
            tokenIn: tokenIn,
            tokenOut: tokenOut,
            fee: bestFee,
            recipient: recipient,
            amountOut: amountOut,
            amountInMaximum: amountInMax,
            sqrtPriceLimitX96: 0
        });

        return swapRouter.exactOutputSingle(params);
    }
}
