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
### 2. fork base(number-1)
anvil --fork-url https://base-mainnet.g.alchemy.com/v2/0aoAtW5IQvhhwLgW4wFQFbW7eM4czhOb --fork-block-number 30672597 --port 8546
export ETH_RPC_URL=http://localhost:8546
cast send --value 1ether 0xFcc65cb843f0667883f3Ac805291511c76B0B5EF --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

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
export ETH_RPC_URL=http://localhost:8546
export PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

forge script --broadcast \
--rpc-url $ETH_RPC_URL \
--private-key $PRIVATE_KEY \
--sig 'run()' \
script/deployParameters/DeployBase.s.sol:DeployBase

== Logs ==
  uniswapV3Dex 0x8A0f7Fa9ac1Afd23121f20a51F646B9Faf10E968
  flashLoanLiquidation 0x3a92c8145cb9694e2E52654707f3Fa71021fc4AC
```

### 设置后端环境, 成功则查询余额
```
<!-- 修改.env为.env.develop -->
<!-- 删除.dev.db -->
npm run db:setup
npm run start:dev

<!-- eth -->
cast call 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48 "balanceOf(address)" 0xFcc65cb843f0667883f3Ac805291511c76B0B5EF
cast call 0x40D16FC0246aD3160Ccc09B8D0D3A2cD28aE6C2f "balanceOf(address)" 0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266
cast call 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2 "balanceOf(address)" 0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266
<!-- base -->
cast call 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913 "balanceOf(address)" 0xFcc65cb843f0667883f3Ac805291511c76B0B5EF
cast call 0x4200000000000000000000000000000000000006 "balanceOf(address)" 0xFcc65cb843f0667883f3Ac805291511c76B0B5EF
cast send --value 1ether 0xFcc65cb843f0667883f3Ac805291511c76B0B5EF --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 --rpc-url http://localhost:8546
```

# wait time
backend/chart.png

S-Shaped Curve Function and Visualization
Function Definition
The piecewise function ( f(x) ) is defined as follows, where ( h ) is the input in hours, ( c_1 = 1 ) second, and ( c_2 = h \cdot 3600 ) seconds:
[f(x) =\begin{cases}1, & \text{if } x \leq 1.0005 \1 + \frac{h \cdot 3600 - 1}{1 + e^{-20(x - 1.50225)}}, & \text{if } 1.0005 < x \leq 2 \h \cdot 3600, & \text{if } x \geq 2\end{cases}]
This function:

Returns ( c_1 = 1 ) second for ( x \leq 1.0005 ).
Transitions smoothly via a sigmoid curve in ( 1.0005 < x \leq 2 ).
Returns ( c_2 = h \cdot 3600 ) seconds for ( x \geq 2 ).

Chart Visualization
The chart below visualizes ( f(x) ) for ( h = 4 ) (i.e., ( c_2 = 4 \cdot 3600 = 14400 ) seconds) over ( x \in [0, 3] ). The y-axis represents the output in seconds, and the x-axis is dimensionless.
{
  "type": "line",
  "data": {
    "datasets": [
      {
        "label": "f(x)",
        "data": [
          {"x": 0, "y": 1},
          {"x": 1.0005, "y": 1},
          {"x": 1.1, "y": 1 + 14399 / (1 + Math.exp(-20 * (1.1 - 1.50225)))},
          {"x": 1.2, "y": 1 + 14399 / (1 + Math.exp(-20 * (1.2 - 1.50225)))},
          {"x": 1.3, "y": 1 + 14399 / (1 + Math.exp(-20 * (1.3 - 1.50225)))},
          {"x": 1.4, "y": 1 + 14399 / (1 + Math.exp(-20 * (1.4 - 1.50225)))},
          {"x": 1.5, "y": 1 + 14399 / (1 + Math.exp(-20 * (1.5 - 1.50225)))},
          {"x": 1.6, "y": 1 + 14399 / (1 + Math.exp(-20 * (1.6 - 1.50225)))},
          {"x": 1.7, "y": 1 + 14399 / (1 + Math.exp(-20 * (1.7 - 1.50225)))},
          {"x": 1.8, "y": 1 + 14399 / (1 + Math.exp(-20 * (1.8 - 1.50225)))},
          {"x": 1.9, "y": 1 + 14399 / (1 + Math.exp(-20 * (1.9 - 1.50225)))},
          {"x": 2, "y": 14400},
          {"x": 3, "y": 14400}
        ],
        "borderColor": "#1e90ff",
        "backgroundColor": "#1e90ff",
        "fill": false,
        "tension": 0,
        "pointRadius": 0
      }
    ]
  },
  "options": {
    "scales": {
      "x": {
        "type": "linear",
        "title": { "display": true, "text": "x" },
        "min": 0,
        "max": 3
      },
      "y": {
        "type": "linear",
        "title": { "display": true, "text": "f(x) (seconds)" },
        "min": 0,
        "max": 15000
      }
    },
    "plugins": {
      "legend": { "display": true },
      "title": { "display": true, "text": "S-Shaped Curve with c1=1s, c2=4h" }
    }
  }
}

The chart displays:

A constant value of 1 second for ( x \leq 1.0005 ).
An S-shaped transition from 1 to 14400 seconds in ( 1.0005 < x \leq 2 ).
A constant value of 14400 seconds for ( x \geq 2 ).

