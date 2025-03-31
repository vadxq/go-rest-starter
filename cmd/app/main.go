package main

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

	v1 "github.com/vadxq/go-rest-starter/api/v1"
	"github.com/vadxq/go-rest-starter/internal/app/config"
	"github.com/vadxq/go-rest-starter/internal/app/db"
	"github.com/vadxq/go-rest-starter/internal/app/handlers"
	"github.com/vadxq/go-rest-starter/internal/app/repository"
	"github.com/vadxq/go-rest-starter/internal/app/services"
)

func main() {
	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// 配置日志输出
	setupLogger(getConfigPath())

	// 加载配置
	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("加载配置失败")
	}

	// 设置日志级别
	setLogLevel(cfg.Log.Level)

	// 初始化数据库连接
	database, err := db.InitDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化数据库失败")
	}

	// 初始化Redis连接
	redisClient, err := db.InitRedis(&cfg.Redis)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化Redis失败")
	}
	defer redisClient.Close()

	// 初始化验证器
	validate := validator.New()

	// 初始化依赖
	deps := initDependencies(database, redisClient, validate)

	// 创建路由
	router := chi.NewRouter()

	// 设置API路由
	v1.SetupRoutes(router, deps.userHandler)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 启动服务器
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-sigChan

		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal().Msg("优雅关闭超时，强制退出")
			}
		}()

		log.Info().Msg("正在关闭服务器...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal().Err(err).Msg("服务器关闭失败")
		}
		serverStopCtx()
	}()

	// 启动服务器
	log.Info().Int("port", cfg.Server.Port).Msg("服务器启动")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("服务器启动失败")
	}

	// 等待服务器完全关闭
	<-serverCtx.Done()
	log.Info().Msg("服务器已关闭")
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

// 依赖注入结构
type dependencies struct {
	userHandler *handlers.UserHandler
}

// 初始化所有依赖
func initDependencies(db *gorm.DB, rdb *redis.Client, validate *validator.Validate) *dependencies {
	// 初始化仓库（Repository 层）
	userRepo := repository.NewUserRepository(db) // 数据访问层，处理数据库操作

	// 初始化服务（Service 层）
	userService := services.NewUserService(userRepo, validate, db) // 业务逻辑层，包含事务管理

	// 初始化处理器（Handler 层）
	userHandler := handlers.NewUserHandler(userService, log.Logger, validate) // HTTP 请求处理层

	return &dependencies{
		userHandler: userHandler, // 最终注入到路由层
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
