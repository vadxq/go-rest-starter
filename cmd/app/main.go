package main

// @title Go-Rest-Starter API
// @version 1.0
// @description Go-Rest-Starter(https://github.com/vadxq/go-rest-starter) RESTful API服务，基于Go Chi、GORM、PostgreSQL和Redis
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://blog.vadxq.com
// @contact.email dxl@vadxq.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:7001
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {token}

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vadxq/go-rest-starter/internal/app"
)

func main() {
	// 创建应用实例
	application, err := app.New()
	if err != nil {
		slog.Error("创建应用失败", "error", err)
		os.Exit(1)
	}

	// 启动HTTP服务器
	serverErrCh := application.StartServer()

	// 等待信号或服务器错误
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case err := <-serverErrCh:
		slog.Error("服务器错误", "error", err)
	case sig := <-signalCh:
		slog.Info("接收到系统信号，开始优雅关闭", "signal", sig.String())
	}

	// 优雅关闭应用
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		slog.Error("应用关闭失败", "error", err)
		os.Exit(1)
	}
}
