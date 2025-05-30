// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

interface IDex {
    // 给定 tokenIn 和 amountIn，换取指定数量 AmountOut 的 tokenOut 给msg.sender，多余的 tokenIn 转成 usdc 给receiver
    function swap(address tokenIn, address tokenOut, uint256 amountIn, uint256 amountOut, address receiver) external;

    /**
     * @notice 获取 DEX 名称
     * @return name DEX 名称
     */
    function name() external view returns (string memory);
}
