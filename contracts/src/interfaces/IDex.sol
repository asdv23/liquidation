// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

interface IDex {
    /**
     * @notice 指定输入金额执行代币兑换
     * @param tokenIn 输入代币地址
     * @param tokenOut 输出代币地址
     * @param amountIn 输入金额
     * @param amountOutMin 最小输出金额
     * @param to 接收地址
     * @return amountOut 实际输出金额
     */
    function swapExactTokensForTokens(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 amountOutMin,
        address to
    ) external returns (uint256 amountOut);

    /**
     * @notice 指定输出金额执行代币兑换
     * @param tokenIn 输入代币地址
     * @param tokenOut 输出代币地址
     * @param amountOut 期望输出金额
     * @param amountInMax 最大输入金额
     * @param to 接收地址
     * @return amountIn 实际输入金额
     */
    function swapTokensForExactTokens(
        address tokenIn,
        address tokenOut,
        uint256 amountOut,
        uint256 amountInMax,
        address to
    ) external returns (uint256 amountIn);

    /**
     * @notice 获取 DEX 名称
     * @return name DEX 名称
     */
    function name() external view returns (string memory);
}
