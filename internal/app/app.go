package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	"github.com/vadxq/go-rest-starter/pkg/logger"
)

// App 应用结构体
type App struct {
	DB        *gorm.DB
	Redis     *redis.Client
	Router    *chi.Mux
	Cache     cache.Cache
	Validator *validator.Validate
	Deps      *injection.Dependencies
	Server    *http.Server
	Config    *config.AppConfig
	logger    *slog.Logger
}

// New 创建新的应用实例
func New() (*App, error) {
	// 配置日志输出
	configPath := getConfigPath()
	programLevel := setupLogger(configPath)

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	// 设置日志级别
	setLogLevel(cfg.Log.Level, programLevel)
	slog.Info("配置加载完成", "config_path", configPath)

	// 创建应用实例
	app := &App{
		Config: cfg,
		logger: slog.Default(),
	}

	// 初始化应用
	if err := app.initialize(); err != nil {
		return nil, fmt.Errorf("初始化应用失败: %w", err)
	}

	return app, nil
}

// initialize 初始化应用组件
func (app *App) initialize() error {
	slog.Info("开始初始化应用...")

	// 初始化数据库连接
	if err := app.initDatabase(); err != nil {
		return fmt.Errorf("初始化数据库失败: %w", err)
	}

	// 初始化Redis连接
	if err := app.initRedis(); err != nil {
		return fmt.Errorf("初始化Redis失败: %w", err)
	}

	// 初始化缓存
	if err := app.initCache(); err != nil {
		return fmt.Errorf("初始化缓存失败: %w", err)
	}

	// 初始化验证器
	app.Validator = validator.New()

	// 初始化依赖注入
	if err := app.initDependencies(); err != nil {
		return fmt.Errorf("初始化依赖注入失败: %w", err)
	}

	// 初始化路由
	if err := app.initRouter(); err != nil {
		return fmt.Errorf("初始化路由失败: %w", err)
	}

	slog.Info("应用初始化完成")
	return nil
}

// initDatabase 初始化数据库连接
func (app *App) initDatabase() error {
	slog.Info("连接数据库...")
	
	database, err := db.InitDB(&app.Config.Database)
	if err != nil {
		return err
	}
	
	app.DB = database
	slog.Info("数据库连接成功")
	return nil
}

// initRedis 初始化Redis连接
func (app *App) initRedis() error {
	slog.Info("连接Redis...")
	
	redisClient, err := db.InitRedis(&app.Config.Redis)
	if err != nil {
		return err
	}
	
	app.Redis = redisClient
	slog.Info("Redis连接成功")
	return nil
}

// initCache 初始化缓存
func (app *App) initCache() error {
	slog.Info("初始化缓存...")
	
	// 缓存服务必须依赖Redis
	if app.Redis == nil {
		slog.Warn("Redis未配置，缓存服务将不可用")
		return nil
	}
	
	cacheOpts := cache.Options{
		DefaultExpiration: 10 * time.Minute,
		CleanupInterval:   5 * time.Minute,
		RedisAddress:      fmt.Sprintf("%s:%d", app.Config.Redis.Host, app.Config.Redis.Port),
		RedisPassword:     app.Config.Redis.Password,
		RedisDB:           app.Config.Redis.DB,
	}

	slog.Info("使用Redis作为缓存存储")
	
	cacheInstance, err := cache.NewCache(cacheOpts)
	if err != nil {
		slog.Error("初始化Redis缓存失败", "error", err)
		// 缓存不是必需的，可以继续运行
		return nil
	}
	
	app.Cache = cacheInstance
	slog.Info("缓存初始化成功")
	return nil
}

// initDependencies 初始化依赖注入
func (app *App) initDependencies() error {
	slog.Info("初始化依赖注入系统...")
	
	// 创建结构化日志器
	structuredLogger, err := logger.NewLogger(&logger.LogConfig{
		Level:   app.Config.Log.Level,
		File:    app.Config.Log.File,
		Console: app.Config.Log.Console,
	})
	if err != nil {
		return fmt.Errorf("创建结构化日志器失败: %w", err)
	}
	
	deps := injection.NewDependencies(
		app.DB,
		app.Redis,
		app.Validator,
		app.Config,
		app.Cache,
		structuredLogger,
	)
	
	app.Deps = deps
	slog.Info("依赖注入系统初始化完成")
	return nil
}

// initRouter 初始化路由
func (app *App) initRouter() error {
	slog.Info("配置API路由...")
	
	router := chi.NewRouter()
	
	api.Setup(router, api.RouterConfig{
		UserHandler:   app.Deps.Handlers.UserHandler,
		AuthHandler:   app.Deps.Handlers.AuthHandler,
		HealthHandler: app.Deps.Handlers.HealthHandler,
		JWTSecret:     app.Deps.Config.JWT.Secret,
	})
	
	app.Router = router
	slog.Info("API路由配置完成")
	return nil
}

// StartServer 启动HTTP服务器
func (app *App) StartServer() <-chan error {
	errCh := make(chan error, 1)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Server.Port),
		Handler:      app.Router,
		ReadTimeout:  app.Config.Server.ReadTimeout,
		WriteTimeout: app.Config.Server.WriteTimeout,
	}

	app.Server = server

	// 启动服务器
	go func() {
		slog.Info("HTTP服务器启动", "port", app.Config.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("HTTP服务器错误: %w", err)
		}
	}()

	return errCh
}

// Shutdown 优雅关闭应用
func (app *App) Shutdown(ctx context.Context) error {
	slog.Info("开始优雅关闭应用...")
	
	// 使用channel收集错误
	errChan := make(chan error, 3)
	
	// 并发关闭各个组件
	go func() {
		if app.Server != nil {
			slog.Info("关闭HTTP服务器...")
			errChan <- app.Server.Shutdown(ctx)
		} else {
			errChan <- nil
		}
	}()
	
	go func() {
		if app.DB != nil {
			slog.Info("关闭数据库连接...")
			if sqlDB, err := app.DB.DB(); err == nil {
				errChan <- sqlDB.Close()
			} else {
				errChan <- err
			}
		} else {
			errChan <- nil
		}
	}()
	
	go func() {
		if app.Redis != nil {
			slog.Info("关闭Redis连接...")
			errChan <- app.Redis.Close()
		} else {
			errChan <- nil
		}
	}()
	
	// 等待所有关闭操作完成
	var hasError bool
	for i := 0; i < 3; i++ {
		if err := <-errChan; err != nil {
			slog.Error("关闭组件失败", "error", err)
			hasError = true
		}
	}
	
	if hasError {
		slog.Warn("应用关闭时出现错误")
	} else {
		slog.Info("应用优雅关闭完成")
	}
	
	return nil
}

// 设置日志配置
func setupLogger(configPath string) *slog.LevelVar {
	programLevel := new(slog.LevelVar)
	programLevel.Set(slog.LevelInfo)

	cfg, err := config.LoadConfig(configPath)
	var output io.Writer = os.Stdout
	handlerOptions := &slog.HandlerOptions{
		AddSource: true,
		Level:     programLevel,
	}

	if err != nil {
		slog.Warn("加载日志配置失败，使用默认控制台文本输出", "error", err)
	} else {
		if cfg.Log.File != "" {
			currentDate := time.Now().Format("2006-01-02")
			logDir := filepath.Dir(cfg.Log.File)
			logFileName := filepath.Base(cfg.Log.File)
			ext := filepath.Ext(logFileName)
			nameWithoutExt := strings.TrimSuffix(logFileName, ext)
			newLogFileName := fmt.Sprintf("%s-%s%s", nameWithoutExt, currentDate, ext)
			logFilePath := filepath.Join(logDir, newLogFileName)

			if mkDirErr := os.MkdirAll(logDir, 0755); mkDirErr != nil {
				slog.Error("无法创建日志目录，使用控制台输出", "dir", logDir, "error", mkDirErr)
			} else {
				logFile, openErr := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
				if openErr != nil {
					slog.Error("无法创建日志文件，使用控制台输出", "path", logFilePath, "error", openErr)
				} else {
					if cfg.Log.Console {
						output = io.MultiWriter(os.Stdout, logFile)
						slog.Info("日志将同时输出到控制台和文件", "file", logFilePath)
					} else {
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
	return programLevel
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
func setLogLevel(level string, programLevel *slog.LevelVar) {
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