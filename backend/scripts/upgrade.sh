#!/bin/bash

# 加载环境变量
source .env

# 检查必要的环境变量
if [ -z "$PRIVATE_KEY" ]; then
    echo "Error: PRIVATE_KEY is not set in .env file"
    exit 1
fi

if [ -z "$FLASH_LOAN_LIQUIDATION" ]; then
    echo "Error: FLASH_LOAN_LIQUIDATION is not set in .env file"
    exit 1
fi

# 编译合约
echo "Compiling contracts..."
forge build

# 执行升级脚本
echo "Running upgrade script..."
forge script scripts/Upgrade.s.sol:UpgradeScript \
    --rpc-url $BASE_RPC_URL \
    --broadcast \
    --verify \
    -vvvv

echo "Upgrade script completed" 