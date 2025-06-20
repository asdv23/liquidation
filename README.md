# Liquidation Bot

多链清算机器人，支持 Aave V3 协议的闪电贷清算。

## 项目结构

```
liquidation/
├── backend/            # Node.js 后端服务
│   ├── src/           # 源代码
│   ├── test/          # 测试
│   ├── config/        # 配置文件
│   ├── prisma/        # 数据库配置
│   ├── abi/           # 合约 ABI
│   ├── dist/          # 编译输出
│   └── node_modules/  # 依赖包
│
├── contracts/         # Solidity 智能合约
│   ├── src/          # 合约源代码
│   ├── test/         # 合约测试
│   ├── script/       # 部署脚本
│   └── lib/          # 依赖库
│
├── .github/          # GitHub 配置
├── .vscode/          # VS Code 配置
├── .gitignore        # Git 忽略文件
├── .gitmodules       # Git 子模块配置
└── README.md         # 项目说明文档
```

## 合约部署
```
forge script --broadcast \
--rpc-url <RPC-URL> \
--private-key <PRIVATE_KEY> \
--sig 'run()' \
script/deployParameters/Deploy<network>.s.sol:Deploy<network>
```

## 项目架构

本项目基于 NestJS 框架，使用 TypeScript 开发，支持多链（如 Base 和 Optimism）监听 Aave V3 合约事件，实时计算用户健康因子（health factor），并在发现不安全贷款时进行记录。

### 主要模块

- **ChainModule**：负责管理多链配置和 Provider，支持动态切换链。
- **BorrowDiscoveryModule**：监听 Aave V3 的 Borrow 和 LiquidationCall 事件，实时计算用户健康因子，并记录不安全贷款。
- **数据库模块**：使用 Prisma ORM，支持 SQLite，记录借款订单和 Token 信息。

## 多链监听

项目支持同时监听多条链（如 Base 和 Optimism）的 Aave V3 合约事件。通过 ChainService 动态获取各链的 Provider 和合约地址，确保实时监听和计算。

### 链配置
- 支持动态配置多链
- 每个链维护独立的 Provider 和合约实例
- 自动验证合约连接和代码存在性

## 借款发现模块

BorrowDiscoveryService 负责监听以下事件：

- **Borrow 事件**：
  - 记录新借款用户
  - 初始化健康因子为 1.0
  - 设置下次检查时间为当前时间
  - 创建数据库记录

- **LiquidationCall 事件**：
  - 从内存中移除被清算用户
  - 更新数据库记录（清算时间、交易哈希等）
  - 计算清算延迟时间

### 健康因子检查
- 定时检查所有活跃贷款
- 使用 multicall 批量获取用户账户数据
- 自动清理已还清债务的用户（totalDebt = 0）
- 根据健康因子动态调整检查间隔

### 内存管理
- 使用 Map 结构存储活跃贷款信息
- 每个链维护独立的贷款集合
- 缓存 Token 信息减少链上查询
- 缓存合约和 Provider 实例提高性能

## 数据库设计

使用 Prisma ORM，支持 SQLite（开发/测试环境）。

### 表结构

#### 1. 借款订单表（Loan）
| 字段名                | 类型           | 说明                   |
|----------------------|----------------|------------------------|
| id                   | Int            | 主键                   |
| chainName            | String         | 链名称                 |
| user                 | String         | 借款人地址             |
| isActive             | Boolean        | 是否活跃               |
| createdAt            | DateTime       | 创建时间               |
| updatedAt            | DateTime       | 更新时间               |
| liquidationDiscoveredAt | DateTime?    | 发现可清算的时间        |
| liquidationTxHash    | String?        | 清算交易哈希           |
| liquidationTime      | DateTime?      | 清算时间               |
| liquidator           | String?        | 清算人地址             |
| liquidationDelay     | Int?           | 从发现到清算的延迟（毫秒）|

#### 2. Token 表
| 字段名           | 类型           | 说明                   |
|------------------|----------------|------------------------|
| id               | Int            | 主键                   |
| chainName        | String         | 链名称                 |
| address          | String         | token 地址             |
| symbol           | String         | token 符号             |
| decimals         | Int            | token 精度             |
| createdAt        | DateTime       | 创建时间               |
| updatedAt        | DateTime       | 更新时间               |

## 健康因子检查机制

### 1. 多级阈值
- 清算阈值 (LIQUIDATION_THRESHOLD): 1.005
- 危险阈值 (CRITICAL_THRESHOLD): 1.01
- 健康阈值 (HEALTH_FACTOR_THRESHOLD): 1.02

### 2. 动态检查间隔
基于健康因子的渐进式检查间隔：
- 低于清算阈值：最小检查间隔（默认 1 秒）
- 低于危险阈值：最小检查间隔 * 2
- 低于健康阈值：使用指数函数计算（800ms - 15分钟）
- 高于健康阈值：使用对数函数计算（15分钟 - 30分钟）

### 3. 批量检查
- 使用 multicall 合约批量获取用户账户数据
- 只检查到达检查时间的用户
- 自动清理已还清债务的用户

## 环境配置

项目支持通过环境变量配置检查间隔。

### 开发环境
```bash
DATABASE_URL="file:./dev.db"
MIN_CHECK_INTERVAL=1000    # 最小检查间隔（毫秒）
MAX_CHECK_INTERVAL=1800000 # 最大检查间隔（毫秒）
```

## 安装和运行

### 1. 安装依赖
```bash
# 安装项目依赖
npm install

# 安装全局依赖（可选）
npm install -g pm2
```

### 2. 环境配置
1. 创建 `.env` 文件：
```bash
cp .env.template .env
```

2. 配置环境变量：
```env
# 数据库配置
DATABASE_URL="file:./dev.db"

# 检查间隔配置（毫秒）
MIN_CHECK_INTERVAL=1000    # 最小检查间隔
MAX_CHECK_INTERVAL=1800000 # 最大检查间隔（30分钟）

# 链配置
BASE_RPC_URL="https://mainnet.base.org"
OPTIMISM_RPC_URL="https://mainnet.optimism.io"
```

### 3. 数据库初始化
```bash
# 生成 Prisma 客户端
npx prisma generate

# 更新数据库结构
npx prisma db push

# 查看数据库（可选）
npx prisma studio
```

### 4. 运行服务

#### 开发模式
```bash
# 启动开发服务器
npm run start:dev

# 或者使用 watch 模式
npm run start:dev:watch
```

#### 生产模式
```bash
# 编译项目
npm run build

# 使用 PM2 启动服务
pm2 start dist/main.js --name "bot" --log-date-format "YYYY-MM-DD HH:mm:ss" --merge-logs

# 查看服务状态
pm2 status

# 查看日志
pm2 logs bot

## 日志轮转
pm2 install pm2-logrotate

# 设置最大日志大小为 50MB
pm2 set pm2-logrotate:max_size 100M

# 保留最近 10 个轮转日志文件
pm2 set pm2-logrotate:retain 10

# 启用日志压缩
pm2 set pm2-logrotate:compress true

# 设置日志文件名时间格式为 YYYY-MM-DD
pm2 set pm2-logrotate:dateFormat YYYY-MM-DD

# 设置每天凌晨 2 点轮转
pm2 set pm2-logrotate:rotateInterval '0 2 * * *'

# 设置检查日志大小的频率为 60 秒
pm2 set pm2-logrotate:workerInterval 60

# 设置时区为 Asia/Shanghai（中国时区，视需求调整）
pm2 set pm2-logrotate:TZ Asia/Shanghai
```

### 5. 服务管理

#### 使用 PM2
```bash
# 重启服务
pm2 restart bot

# 停止服务
pm2 stop bot

# 删除服务
pm2 delete bot

# 查看资源使用
pm2 monit
```

#### 使用系统服务
```bash
# 使用 nohup 启动
nohup npm run start:prod > liquidation.log 2>&1 &

# 查看进程
ps aux | grep "npm run start:prod"

# 停止服务
kill <进程ID>
```

### 6. 日志管理
```bash
tail -n 1000 ~/.pm2/logs/bot-out.log

# 清空所有日志
pm2 flush
```

### 7. 更新服务
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

# 5. 重启服务
pm2 restart bot
```

## 后续计划

### 1. 清算模块
- 闪电贷不支持 debtToken 清算，需要使用 usdc 清算
- go 重写 nodejs 服务
- odos 请求响应时间打印
- 链下计算太慢导致来不及提交交易？

### 2. 监控和告警
- 实现监控面板
  - 显示活跃贷款数量
  - 显示健康因子分布
  - 显示清算机会
- 告警系统
  - 邮件通知
  - Telegram 机器人
  - 自定义 webhook


## 技术栈

### 核心框架
- NestJS：后端框架
- TypeScript：开发语言
- Prisma：ORM 框架
- SQLite：数据库

### 区块链相关
- ethers.js：以太坊交互
- Aave V3：借贷协议
- Multicall：批量查询

### 开发工具
- Node.js：运行环境
- npm：包管理
- Git：版本控制

### 部署和监控
- PM2：进程管理
- Docker：容器化
- 日志系统：内置

## 环境要求

### 系统要求
- Node.js >= 18.18
- npm >= 9.0
- 4GB+ RAM
- 20GB+ 磁盘空间

### 网络要求
- 稳定的网络连接
- 支持 WebSocket
- 低延迟 RPC 节点

### 开发环境
- VSCode（推荐）
- Git
- SQLite 客户端

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
MIN_CHECK_INTERVAL=1000
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

## 故障排除

### 1. 常见问题

#### 数据库问题
```bash
# 数据库连接失败
Error: P1001: Can't reach database server

解决方案：
1. 检查 DATABASE_URL 配置
2. 确保数据库文件权限正确
3. 尝试重新初始化数据库
```

#### RPC 问题
```bash
# RPC 连接超时
Error: timeout exceeded

解决方案：
1. 检查网络连接
2. 更换 RPC 节点
3. 增加超时时间
```

#### 内存问题
```bash
# 内存使用过高
FATAL ERROR: Ineffective mark-compacts near heap limit

解决方案：
1. 增加 Node.js 内存限制
2. 优化内存使用
3. 定期重启服务
```

### 2. 日志分析

#### 错误日志
```bash
# 查看最近的错误
grep "ERROR" liquidation.log | tail -n 50

# 查看特定链的错误
grep "ERROR.*\[base\]" liquidation.log

# 查看健康因子相关错误
grep "ERROR.*health factor" liquidation.log
```

#### 性能日志
```bash
# 查看检查间隔
grep "Next check" liquidation.log

# 查看批量检查结果
grep "Checking health factors" liquidation.log

# 查看内存使用
grep "Memory usage" liquidation.log
```

## 性能优化

### 1. RPC 优化

#### 批量请求
- 使用 multicall 合约
- 合并相似请求
- 实现请求队列

#### 连接优化
- 使用 WebSocket 连接
- 实现连接池
- 自动重连机制

### 2. 内存优化

#### 缓存策略
- 实现 LRU 缓存
- 定期清理过期数据
- 限制缓存大小

#### 数据结构
- 优化 Map 结构
- 使用 WeakMap
- 实现数据分片

### 3. 数据库优化

#### 查询优化
- 使用索引
- 优化查询语句
- 实现查询缓存

#### 存储优化
- 定期清理历史数据
- 压缩数据库
- 实现数据归档

## 贡献指南

### 1. 开发流程

#### 代码规范
- 使用 TypeScript
- 遵循 ESLint 规则
- 编写单元测试

#### 提交规范
```bash
# 提交格式
<type>(<scope>): <subject>

# 类型说明
feat: 新功能
fix: 修复
docs: 文档
style: 格式
refactor: 重构
test: 测试
chore: 构建
```

### 2. 测试规范

#### 单元测试
```bash
# 运行测试
npm run test

# 运行特定测试
npm run test -- -t "health factor"

# 测试覆盖率
npm run test:cov
```

#### 集成测试
```bash
# 运行集成测试
npm run test:e2e

# 测试特定模块
npm run test:e2e -- -t "borrow"
```

### 3. 文档规范

#### 代码注释
```typescript
/**
 * 函数说明
 * @param {string} param1 - 参数1说明
 * @returns {Promise<void>} 返回值说明
 */
```

#### 文档更新
- 更新 README.md
- 更新 API 文档
- 更新注释

## 安全说明

### 1. 私钥安全
- 不要提交私钥
- 使用环境变量
- 加密存储

### 2. 合约安全
- 验证合约地址
- 检查合约代码
- 限制调用权限

### 3. 数据安全
- 加密敏感数据
- 定期备份
- 访问控制

## 版本历史

### v0.1.0 (2024-03-xx)
- 初始版本
- 基础功能实现
- 多链支持

### 待发布
- 清算模块
- 监控面板
- 性能优化

## 许可证

MIT License

## 联系方式

- 项目维护者：[维护者名称]
- 邮箱：[邮箱地址]
- Telegram：[Telegram 链接]

# Add a new chain
1. DeployXXX.s.sol
2. DeployContracts
3. Update .env
4. testTheGraph.js



# 清算
a=b
a!=b
  a in flash
  a not int flash

total
partial
