# 如何验证清算功能正常
## 成功清算
1. 找到主网一笔已清算交易，往前搜索到一个区块高度能清算此笔交易
2. 本地从这个区块高度 fork 网络
3. 部署清算合约，并将合约地址更新到配置文件
4. 数据库中插入这笔待清算的交易
5. 程序运行，执行清算成功

## 失败清算
1. 找到主网一笔已清算交易，往前搜索到一个区块高度不能清算此笔交易
2. 本地从这个区块高度 fork 网络
3. 部署清算合约，并将合约地址更新到配置文件
4. 数据库中插入这笔待清算的交易
5. 程序运行，执行清算失败，仅扣除少量 gas 费。

## 示例 1

### 1. 找到一笔已清算交易
https://basescan.org/tx/0xf71a2ce14c968faf3ea01ff2d3ce78d3df36a47b01eb2608dc2f69c98a325178
```
LiquidationCall - height:30672598
8453:0xa238dd80c259a72e81d7e4664a9801593f98d1c5
{
    "collateralAsset":"0x4200000000000000000000000000000000000006"
    "debtAsset":"0x833589fcd6edb6e08f4c7c32d4f71b54bda02913"
    "user":"0xe9526c721d489464079acb7568c1304af8687298"
    "debtToCover":"120857040"
    "liquidatedCollateralAmount":"50929511049751630"
    "liquidator":"0x888888887a487f209e31a692b227d8d1ff9070ba"
    "receiveAToken":false
}
```
### 2. fork base
anvil --fork-url https://base-mainnet.g.alchemy.com/v2/0aoAtW5IQvhhwLgW4wFQFbW7eM4czhOb --fork-block-number 30672597 --port 8546

#### 验证可以清算 - 查询链上健康因子
```
curl --location 'http://localhost:8546' \
--header 'Content-Type: application/json' \
--data '{
    "jsonrpc": "2.0",
    "method": "eth_call",
    "params": [
        {
            "data": "0xbf92857c000000000000000000000000e9526c721d489464079acb7568c1304af8687298",
            "to": "0xA238Dd80C259a72e81d7e4664a9801593F98d1c5"
        },
        "latest"
    ],
    "id": 1
}'

{"jsonrpc":"2.0","id":1,"result":"0x00000000000000000000000000000000000000000000000000000003624a5d5300000000000000000000000000000000000000000000000000000002d046b0ef0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000206c0000000000000000000000000000000000000000000000000000000000001f400000000000000000000000000000000000000000000000000dda85f7ab3168ad"}
```

#### 验证可以清算 - 解码，并检查 healthFactor＜ 1
```
# 结果复制并修改testHealthFactor.js 
% node backend/test/testHealthFactor.js

totalCollateralBase: 145.33942611 USD
totalDebtBase: 120.84228335 USD
availableBorrowsBase: 0.0 USD
currentLiquidationThreshold: 8300 bps
ltv: 8000 bps
healthFactor: 0.998257566191544493
```

### 部署合约
Private Key: 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
Address: 0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266
```
cd contracts
forge script --broadcast \
--rpc-url http://localhost:8546 \
--private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
--sig 'run()' \
script/deployParameters/DeployBase.s.sol:DeployBase

== Logs ==
  params.aaveV3Pool 0xA238Dd80C259a72e81d7e4664a9801593F98d1c5
  params.usdc 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913
  params.swapRouter02 0x2626664c2603336E57B271c5C0b26F421741e481
  params.uniswapV3Dex 0x37767d8102966577A4f5c7930e0657C592E5061b
  params.flashLoanLiquidation 0x1E5fc0875e2646562Cf694d992182CBb96033Ce4
```
