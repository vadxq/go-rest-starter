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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/internal/app/config"
	"github.com/vadxq/go-rest-starter/internal/app/db"
	"github.com/vadxq/go-rest-starter/internal/app/injection"
	api "github.com/vadxq/go-rest-starter/internal/app/router"
	"github.com/vadxq/go-rest-starter/pkg/cache"
)

var programLevel = new(slog.LevelVar) // 全局日志级别控制器

func main() {
	// 初始化日志级别为 Info，setupLogger 会根据配置覆盖
	programLevel.Set(slog.LevelInfo)

	// 配置日志输出
	configPath := getConfigPath()
	setupLogger(configPath) // setupLogger 内部会设置 slog.SetDefault

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}

	// 设置日志级别 (基于配置文件)
	setLogLevel(cfg.Log.Level) // 会更新 programLevel
	slog.Info("配置加载完成", "config_path", configPath)

	// 初始化应用
	app, err := initApp(cfg)
	if err != nil {
		slog.Error("初始化应用失败", "error", err)
		os.Exit(1)
	}

	// 启动HTTP服务器
	serverErrCh := startServer(app, cfg.Server.Port, cfg.Server.ReadTimeout, cfg.Server.WriteTimeout)

	// 处理优雅关闭
	shutdownApp(app, serverErrCh)
}

// App 应用结构体
type App struct {
	DB        *gorm.DB
	Redis     *redis.Client
	Router    *chi.Mux
	Cache     cache.Cache
	Validator *validator.Validate
	Deps      *injection.Dependencies
	Server    *http.Server
}

// 初始化应用
func initApp(cfg *config.AppConfig) (*App, error) {
	slog.Info("开始初始化应用...")

	// 初始化数据库连接
	slog.Info("连接数据库...")
	database, err := db.InitDB(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}
	slog.Info("数据库连接成功")

	// 初始化Redis连接
	slog.Info("连接Redis...")
	redisClient, err := db.InitRedis(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("初始化Redis失败: %w", err)
	}
	slog.Info("Redis连接成功")

	// 初始化缓存
	slog.Info("初始化缓存...")
	cacheInstance, err := initCache(redisClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化缓存失败: %w", err)
	}
	slog.Info("缓存初始化成功")

	// 初始化验证器
	validate := validator.New()

	// 初始化依赖注入系统
	slog.Info("初始化依赖注入系统...")
	// 此处需要确保 slog.Default() 返回的是已配置的 logger
	// setupLogger 中已通过 slog.SetDefault 设置
	deps := injection.NewDependencies(
		database,       // 数据库连接
		redisClient,    // Redis客户端
		validate,       // 验证器
		cfg,            // 应用配置
		cacheInstance,  // 缓存实例
		slog.Default(), // 日志记录器 (slog.Logger)
	)
	slog.Info("依赖注入系统初始化完成")

	// 创建HTTP路由器
	router := chi.NewRouter()

	// 设置API路由
	slog.Info("配置API路由...")
	api.Setup(router, api.RouterConfig{
		UserHandler: deps.Handlers.UserHandler, // 用户处理器
		AuthHandler: deps.Handlers.AuthHandler, // 认证处理器
		JWTSecret:   deps.Config.JWT.Secret,    // JWT密钥
	})
	slog.Info("API路由配置完成")

	return &App{
		DB:        database,
		Redis:     redisClient,
		Router:    router,
		Cache:     cacheInstance,
		Validator: validate,
		Deps:      deps,
	}, nil
}

// 启动HTTP服务器
func startServer(app *App, port int, readTimeout, writeTimeout time.Duration) <-chan error {
	errCh := make(chan error, 1)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      app.Router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// 保存服务器实例
	app.Server = server

	// 启动服务器
	go func() {
		slog.Info("HTTP服务器启动", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP服务器错误: %w", err)
		}
	}()

	return errCh
}

// 处理应用优雅关闭
func shutdownApp(app *App, serverErrCh <-chan error) {
	// 创建信号通道
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case err := <-serverErrCh:
		slog.Error("服务器错误", "error", err)
	case sig := <-signalCh:
		slog.Info("接收到系统信号，开始优雅关闭", "signal", sig.String())
	}

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭HTTP服务器
	if app.Server != nil {
		slog.Info("关闭HTTP服务器...")
		if err := app.Server.Shutdown(ctx); err != nil {
			slog.Error("关闭HTTP服务器失败", "error", err)
		}
	}

	// 关闭数据库连接
	if app.DB != nil {
		slog.Info("关闭数据库连接...")
		sqlDB, err := app.DB.DB()
		if err != nil {
			slog.Error("获取底层SQL DB失败", "error", err)
		} else {
			if err := sqlDB.Close(); err != nil {
				slog.Error("关闭数据库连接失败", "error", err)
			}
		}
	}

	// 关闭Redis连接
	if app.Redis != nil {
		slog.Info("关闭Redis连接...")
		if err := app.Redis.Close(); err != nil {
			slog.Error("关闭Redis连接失败", "error", err)
		}
	}

	slog.Info("应用关闭完成")
}

// 初始化缓存
func initCache(redisClient *redis.Client, cfg *config.AppConfig) (cache.Cache, error) {
	cacheOpts := cache.Options{
		DefaultExpiration: 10 * time.Minute,
		CleanupInterval:   5 * time.Minute,
	}

	// 如果Redis可用，则使用Redis作为缓存
	if redisClient != nil {
		cacheOpts.RedisAddress = fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
		cacheOpts.RedisPassword = cfg.Redis.Password
		cacheOpts.RedisDB = cfg.Redis.DB

		slog.Info("使用Redis作为缓存存储")
	} else {
		slog.Info("使用内存作为缓存存储")
	}

	return cache.NewCache(cacheOpts)
}

// 设置日志配置
func setupLogger(configPath string) {
	cfg, err := config.LoadConfig(configPath)
	var output io.Writer = os.Stdout // 默认输出到控制台
	handlerOptions := &slog.HandlerOptions{
		AddSource: true,         // 添加源码位置信息
		Level:     programLevel, // 使用全局动态级别
	}

	if err != nil {
		slog.Warn("加载日志配置失败，使用默认控制台文本输出", "error", err)
		// 继续使用默认的 programLevel (Info) 和 os.Stdout
	} else {
		// 根据配置决定日志输出位置
		if cfg.Log.File != "" {
			// 获取当前日期
			currentDate := time.Now().Format("2006-01-02")

			// 构建日志文件路径：将日期添加到文件名中
			logDir := filepath.Dir(cfg.Log.File)
			logFileName := filepath.Base(cfg.Log.File)
			ext := filepath.Ext(logFileName)
			nameWithoutExt := strings.TrimSuffix(logFileName, ext)
			newLogFileName := fmt.Sprintf("%s-%s%s", nameWithoutExt, currentDate, ext)
			logFilePath := filepath.Join(logDir, newLogFileName)

			// 确保日志目录存在
			if mkDirErr := os.MkdirAll(logDir, 0755); mkDirErr != nil {
				slog.Error("无法创建日志目录，使用控制台输出", "dir", logDir, "error", mkDirErr)
			} else {
				logFile, openErr := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
				if openErr != nil {
					slog.Error("无法创建日志文件，使用控制台输出", "path", logFilePath, "error", openErr)
				} else {
					if cfg.Log.Console { // 同时输出到控制台和文件
						output = io.MultiWriter(os.Stdout, logFile)
						slog.Info("日志将同时输出到控制台和文件", "file", logFilePath)
					} else { // 只输出到文件
						output = logFile
						slog.Info("日志将输出到文件", "file", logFilePath)
					}
				}
			}
		} else {
			slog.Info("日志将输出到控制台")
		}
	}

	logger := slog.New(slog.NewTextHandler(output, handlerOptions))
	slog.SetDefault(logger)
}

// 获取配置文件路径
func getConfigPath() string {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	return configPath
}

// 设置日志级别
func setLogLevel(level string) {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		slog.Warn("未知的日志级别，将使用Info级别", "configured_level", level)
		l = slog.LevelInfo
	}
	programLevel.Set(l)
	slog.Info("日志级别设置为", "level", l.String())
}
