# 清算机器人后端服务

这是一个使用 Go 语言编写的多链清算机器人后端服务。

## 技术栈

- Go 1.21+
- Gin Web 框架
- GORM ORM
- PostgreSQL 数据库
- go-ethereum 区块链交互
- Zap 日志库
- Viper 配置管理
- Wire 依赖注入
- Swagger API 文档

## 项目结构

```
.
├── cmd/                # 主程序入口
├── config/            # 配置文件
├── internal/          # 内部包
│   ├── api/          # API 路由和处理
│   ├── core/         # 核心服务
│   ├── models/       # 数据模型
│   └── utils/        # 工具函数
├── pkg/              # 可重用的包
│   ├── blockchain/   # 区块链交互
│   └── database/     # 数据库操作
├── scripts/          # 脚本文件
└── test/            # 测试文件
```

## 快速开始

1. 安装依赖：

```bash
go mod download
```

2. 配置环境：

复制 `config/config.yaml.example` 到 `config/config.yaml` 并修改配置。

3. 运行服务：

```bash
go run cmd/main.go
```

## API 文档

启动服务后访问 `http://localhost:8080/swagger/index.html` 查看 API 文档。

## 主要功能

- 多链清算机会监控
- 自动执行清算交易
- 市场数据监控
- 风险控制
- 性能监控

## 开发指南

1. 代码风格遵循 Go 标准规范
2. 使用 `go fmt` 格式化代码
3. 编写单元测试
4. 使用 `go vet` 检查代码问题

## 部署

1. 构建二进制文件：

```bash
go build -o liquidation-bot cmd/main.go
```

2. 运行服务：

```bash
./liquidation-bot
```

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

MIT 