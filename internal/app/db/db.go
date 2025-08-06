package db

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vadxq/go-rest-starter/internal/app/config"
)

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// 生产环境优化：调整日志级别
	logLevel := logger.Warn
	if cfg.Driver == "development" {
		logLevel = logger.Info
	}
	
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		PrepareStmt:            true,  // 预编译语句，提升性能
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库连接失败: %w", err)
	}

	// 生产环境连接池优化
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns) 
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxLifetime / 2) // 空闲连接超时

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库ping失败: %w", err)
	}

	return db, nil
}

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 10,           // 连接池大小
		MinIdleConns: 5,        // 最小空闲连接
		MaxRetries: 2,          // 最大重试次数
		DialTimeout: 5 * time.Second,
		ReadTimeout: 3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	return rdb, nil
}
