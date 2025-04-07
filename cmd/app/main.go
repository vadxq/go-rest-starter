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
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/api"
	"github.com/vadxq/go-rest-starter/internal/app/config"
	"github.com/vadxq/go-rest-starter/internal/app/db"
	"github.com/vadxq/go-rest-starter/internal/app/injection"
	"github.com/vadxq/go-rest-starter/internal/pkg/cache"
)

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// 配置日志输出
	configPath := getConfigPath()
	setupLogger(configPath)

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("加载配置失败")
	}

	// 设置日志级别
	setLogLevel(cfg.Log.Level)
	log.Info().Str("config_path", configPath).Msg("配置加载完成")

	// 初始化应用
	app, err := initApp(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化应用失败")
	}

	// 启动HTTP服务器
	serverErrCh := startServer(app.Router, cfg.Server.Port, cfg.Server.ReadTimeout, cfg.Server.WriteTimeout)

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
}

// 初始化应用
func initApp(cfg *config.AppConfig) (*App, error) {
	log.Info().Msg("开始初始化应用...")
	
	// 初始化数据库连接
	log.Info().Msg("连接数据库...")
	database, err := db.InitDB(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}
	log.Info().Msg("数据库连接成功")

	// 初始化Redis连接
	log.Info().Msg("连接Redis...")
	redisClient, err := db.InitRedis(&cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("初始化Redis失败: %w", err)
	}
	log.Info().Msg("Redis连接成功")
	
	// 初始化缓存
	log.Info().Msg("初始化缓存...")
	cacheInstance, err := initCache(redisClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化缓存失败: %w", err)
	}
	log.Info().Msg("缓存初始化成功")

	// 初始化验证器
	validate := validator.New()

	// 初始化依赖注入系统
	log.Info().Msg("初始化依赖注入系统...")
	deps := injection.NewDependencies(
		database,       // 数据库连接
		redisClient,    // Redis客户端 
		validate,       // 验证器
		cfg,            // 应用配置
		cacheInstance,  // 缓存实例
		log.Logger,     // 日志记录器
	)
	log.Info().Msg("依赖注入系统初始化完成")

	// 创建HTTP路由器
	router := chi.NewRouter()

	// 设置API路由
	log.Info().Msg("配置API路由...")
	api.Setup(router, api.RouterConfig{
		UserHandler: deps.Handlers.UserHandler, // 用户处理器
		AuthHandler: deps.Handlers.AuthHandler, // 认证处理器
		JWTSecret:   deps.Config.JWT.Secret,    // JWT密钥
	})
	log.Info().Msg("API路由配置完成")
	
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
func startServer(router *chi.Mux, port int, readTimeout, writeTimeout time.Duration) <-chan error {
	errCh := make(chan error, 1)
	
	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	// 启动服务器
	go func() {
		log.Info().Int("port", port).Msg("HTTP服务器启动")
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
		log.Error().Err(err).Msg("服务器错误")
	case sig := <-signalCh:
		log.Info().Str("signal", sig.String()).Msg("接收到系统信号，开始优雅关闭")
	}
	
	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 关闭Redis连接
	if app.Redis != nil {
		log.Info().Msg("关闭Redis连接...")
		if err := app.Redis.Close(); err != nil {
			log.Error().Err(err).Msg("关闭Redis连接失败")
		}
	}
	
	// 关闭服务器
	log.Info().Msg("应用关闭完成")
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
		
		log.Info().Msg("使用Redis作为缓存存储")
	} else {
		log.Info().Msg("使用内存作为缓存存储")
	}
	
	return cache.NewCache(cacheOpts)
}

// 设置日志配置
func setupLogger(configPath string) {
	// 先加载配置以获取日志设置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// 如果无法加载配置，默认输出到控制台
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
		log.Error().Err(err).Msg("加载配置失败，使用默认日志设置")
		return
	}

	// 根据配置决定日志输出位置
	if cfg.Log.File != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.Log.File)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Error().Err(err).Str("dir", logDir).Msg("无法创建日志目录，使用控制台输出")
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
			return
		}

		// 创建日志文件
		logFile, err := os.OpenFile(cfg.Log.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Error().Err(err).Str("path", cfg.Log.File).Msg("无法创建日志文件，使用控制台输出")
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
			return
		}

		// 如果需要同时输出到控制台和文件
		if cfg.Log.Console {
			// 创建多输出
			multi := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stdout}, logFile)
			log.Logger = zerolog.New(multi).With().Timestamp().Logger()
		} else {
			// 只输出到文件
			log.Logger = zerolog.New(logFile).With().Timestamp().Logger()
		}
	} else {
		// 默认输出到控制台
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
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
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
