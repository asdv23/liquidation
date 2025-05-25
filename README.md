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
```

### 生产环境（PostgreSQL）
```
DATABASE_URL="postgresql://user:password@localhost:5432/liquidation_bot"
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

## 已实现功能

### 1. 健康因子监控
- 多级阈值监控（清算阈值、危险阈值、健康阈值）
- 动态检查间隔（基于健康因子的渐进式检查）
- 可配置的最小/最大检查间隔（默认：200ms - 30分钟）

### 2. 数据库支持
- 使用 Prisma + SQLite 存储贷款信息
- 记录贷款健康因子历史
- 追踪清算事件和延迟
- 支持多链数据管理

### 3. 清算事件追踪
- 记录清算发现时间
- 追踪清算执行时间
- 记录清算交易哈希
- 计算清算延迟时间

### 4. 多链支持
- 支持 Base 和 Optimism 网络
- 链配置管理
- 合约连接验证
- 事件监听系统

### 5. 事件监听
- Borrow 事件监听
- Repay 事件监听
- LiquidationCall 事件监听
- 自动更新贷款状态

### 6. 配置管理
- 环境变量配置
- 链配置管理
- 阈值配置
- 检查间隔配置

## 技术栈
- NestJS
- TypeScript
- Prisma
- SQLite
- ethers.js

## 环境要求
- Node.js >= 18.18
- npm

## 安装
```bash
npm install
```

## 配置
1. 创建 `.env` 文件
2. 配置数据库连接：
```
DATABASE_URL="file:./dev.db"
```
3. 配置检查间隔（可选）：
```
MIN_CHECK_INTERVAL=200
MAX_CHECK_INTERVAL=1800000
```

## 运行
```bash
# 开发模式
npm run start:dev

# 生产模式
npm run start:prod
```

## 数据库初始化
```bash
npx prisma generate
npx prisma db push
```

## 后台服务管理

### 1. 编译项目
```bash
# 编译 TypeScript 代码
npm run build

# 生成 Prisma 客户端
npx prisma generate
```

### 2. 启动后台服务
```bash
# 使用 nohup 启动服务并将输出重定向到日志文件
nohup npm run start:prod > liquidation.log 2>&1 &

# 查看进程 ID
ps aux | grep "npm run start:prod" | grep -v grep

# 查看日志
tail -f liquidation.log
```

### 3. 重启服务
```bash
# 查找进程 ID
ps aux | grep "npm run start:prod" | grep -v grep

# 停止服务
kill <进程ID>

# 重新启动服务
nohup npm run start:prod > liquidation.log 2>&1 &
```

### 4. 更新服务
```bash
# 1. 拉取最新代码
git pull

# 2. 安装依赖
npm install

# 3. 编译项目
npm run build

# 4. 更新数据库
npx prisma generate
npx prisma db push
# 查看 SQLite 数据库信息
npx prisma studio

# 5. 重启服务
# 查找进程 ID
ps aux | grep "npm run start:prod" | grep -v grep
# 停止服务
kill <进程ID>
# 启动服务
nohup npm run start:prod > liquidation.log 2>&1 &
```

### 5. 监控服务状态
```bash
# 查看服务进程状态
ps aux | grep "npm run start:prod" | grep -v grep

# 查看服务日志
tail -f liquidation.log

# 查看服务错误日志
grep "ERROR" liquidation.log

# 查看服务内存使用
pm2 monit  # 如果使用 PM2 管理进程
```

### 6. 使用 PM2 管理服务（推荐）
```bash
# 安装 PM2
npm install -g pm2

# 启动服务
pm2 start dist/main.js --name "bot"

# 查看服务状态
pm2 status

# 查看日志
pm2 logs bot

# 重启服务
pm2 restart bot

# 停止服务
pm2 stop bot

# 删除服务
pm2 delete bot
```
