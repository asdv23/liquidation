package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"liquidation-bot/config"
	"liquidation-bot/internal/api"
	"liquidation-bot/internal/core"
	"liquidation-bot/pkg/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// initLogger 初始化日志记录器
func initLogger() (*zap.Logger, error) {
	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		// TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "chain",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		// zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	// 创建日志记录器
	logger := zap.New(core /*zap.AddCaller(),*/, zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, nil
}

func main() {
	// 初始化日志
	logger, err := initLogger()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 加载配置
	cfg, err := config.NewConfig(logger, "config/config.yaml")
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// 初始化数据库连接
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 初始化 Gin 路由
	gin.SetMode(gin.ReleaseMode)
	router := gin.New() // 使用 gin.New() 而不是 gin.Default()，因为我们要使用自定义的日志中间件

	// 初始化服务
	services := core.NewServices(db, cfg, logger)

	// 注册路由
	api.RegisterRoutes(router, services)

	// 启动服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}
