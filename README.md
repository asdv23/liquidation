# 多链清算机器人

## 项目架构

本项目基于 NestJS 框架，使用 TypeScript 开发，支持多链（如 Base 和 Optimism）监听 Aave V3 合约事件，实时计算用户健康因子（health factor），并在发现不安全贷款时进行记录和清算。

### 主要模块

- **ChainModule**：负责管理多链配置和 Provider，支持动态切换链。
- **BorrowDiscoveryModule**：监听 Aave V3 的 Borrow、Repay 和 LiquidationCall 事件，实时计算用户健康因子，并记录不安全贷款。
- **数据库模块**：使用 Prisma ORM，支持 PostgreSQL 和 SQLite（开发/测试环境），记录借款订单和健康因子变动历史。

## 多链监听

项目支持同时监听多条链（如 Base 和 Optimism）的 Aave V3 合约事件。通过 ChainService 动态获取各链的 Provider 和合约地址，确保实时监听和计算。

## 借款发现模块

BorrowDiscoveryService 负责监听以下事件：

- **Borrow 事件**：记录新借款用户，并检查其健康因子。
- **Repay 事件**：用户主动偿还贷款时，从活跃借款列表中删除。
- **LiquidationCall 事件**：用户被清算时，从活跃借款列表中删除，并记录清算时间。

同时，定时轮询所有活跃借款用户的健康因子，确保及时发现不安全的贷款。

### 渐进式轮询
为了实现最大程度的节省 rpc 调用次数，对于健康因子非常大的借款，没有必要频繁检测其变化，但对于健康因子逼近清算阈值的借款，需要非常频繁的检测其变化，以便其一出现就立刻可以发送清算交易来获得清算激励。

## 数据库设计

使用 Prisma ORM，支持 PostgreSQL 和 SQLite（开发/测试环境）。

### 表结构

#### 1. 借款订单表（loans）
| 字段名           | 类型           | 说明                   |
|------------------|----------------|------------------------|
| id               | SERIAL PRIMARY KEY | 主键                |
| chain            | VARCHAR(32)    | 链名（如 base/op）     |
| userAddress      | VARCHAR(64)    | 借款用户地址           |
| borrowTxHash     | VARCHAR(66)    | 借款交易哈希           |
| borrowTime       | TIMESTAMP      | 借款时间               |
| repayTxHash      | VARCHAR(66)    | 偿还交易哈希（可空）    |
| repayTime        | TIMESTAMP      | 偿还时间（可空）        |
| liquidateTxHash  | VARCHAR(66)    | 清算交易哈希（可空）    |
| liquidateTime    | TIMESTAMP      | 清算时间（可空）        |
| lastHealthFactor | NUMERIC        | 最近一次检测到的健康因子 |
| status           | VARCHAR(16)    | 状态（active/closed/liquidated）|

#### 2. 健康因子变动历史表（loan_health_history）
| 字段名           | 类型           | 说明                   |
|------------------|----------------|------------------------|
| id               | SERIAL PRIMARY KEY | 主键                |
| loanId           | INTEGER        | loans.id 外键           |
| checkTime        | TIMESTAMP      | 检查时间               |
| healthFactor     | NUMERIC        | 检查时的健康因子        |

## 环境配置

项目支持通过环境变量动态切换数据库类型和轮询间隔。

### 开发环境（SQLite）
```
DATABASE_URL="file:memory:"
POLLING_INTERVAL=300000  # 5分钟
```

### 生产环境（PostgreSQL）
```
DATABASE_URL="postgresql://user:password@localhost:5432/liquidation_bot"
POLLING_INTERVAL=300000  # 5分钟
```

## 启动项目

1. 安装依赖：
   ```bash
   npm install
   ```

2. 配置环境变量：
   - 复制 `.env.template` 为 `.env`，并根据需要修改配置。

3. 初始化数据库：
   ```bash
   npx prisma migrate dev --name init
   ```

4. 启动服务：
   ```bash
   npm run start:dev
   ```

## 后续计划

- 集成清算模块，自动发送清算交易。
- 增加监控和告警功能，实时通知不安全贷款。
- 优化数据库查询，支持复杂统计和分析。
