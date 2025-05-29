// SPDX-License-Identifier: MIT
pragma solidity ^0.8.22;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@uniswap/v3-periphery/contracts/interfaces/ISwapRouter.sol";
import "@uniswap/v3-core/contracts/interfaces/IUniswapV3Pool.sol";
import "../interfaces/IDex.sol";

contract UniswapV3Dex is Initializable, UUPSUpgradeable, OwnableUpgradeable, IDex {
    ISwapRouter public swapRouter;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address _swapRouter) public initializer {
        __UUPSUpgradeable_init();
        __Ownable_init(msg.sender);
        swapRouter = ISwapRouter(_swapRouter);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyOwner {}

    function name() external pure override returns (string memory) {
        return "UniswapV3Dex";
    }

    function swapExactTokensForTokens(
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 amountOutMin,
        address recipient
    ) external override returns (uint256) {
        IERC20(tokenIn).transferFrom(msg.sender, address(this), amountIn);
        IERC20(tokenIn).approve(address(swapRouter), amountIn);

        // 使用智能路由
        bytes memory path = abi.encodePacked(
            tokenIn,
            uint24(0), // 让智能路由自动选择最优费率
            tokenOut
        );

        ISwapRouter.ExactInputParams memory params = ISwapRouter.ExactInputParams({
            path: path,
            recipient: recipient,
            deadline: block.timestamp + 15 minutes,
            amountIn: amountIn,
            amountOutMinimum: amountOutMin
        });

        uint256 amountOut = swapRouter.exactInput(params);
        return amountOut;
    }

    function swapTokensForExactTokens(
        address tokenIn,
        address tokenOut,
        uint256 amountOut,
        uint256 amountInMax,
        address recipient
    ) external override returns (uint256) {
        IERC20(tokenIn).transferFrom(msg.sender, address(this), amountInMax);
        IERC20(tokenIn).approve(address(swapRouter), amountInMax);

        // 使用智能路由
        bytes memory path = abi.encodePacked(
            tokenIn,
            uint24(0), // 让智能路由自动选择最优费率
            tokenOut
        );

        ISwapRouter.ExactOutputParams memory params = ISwapRouter.ExactOutputParams({
            path: path,
            recipient: recipient,
            deadline: block.timestamp + 15 minutes,
            amountOut: amountOut,
            amountInMaximum: amountInMax
        });

        uint256 amountIn = swapRouter.exactOutput(params);

        // 返还多余的代币
        if (amountIn < amountInMax) {
            IERC20(tokenIn).transfer(msg.sender, amountInMax - amountIn);
        }

        return amountIn;
    }
}
