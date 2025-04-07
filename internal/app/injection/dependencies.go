package injection

import (
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/internal/app/config"
	"github.com/vadxq/go-rest-starter/internal/pkg/cache"
)

// Dependencies 应用依赖容器
// 这是应用程序的核心依赖容器，遵循依赖注入模式组织各层依赖关系
type Dependencies struct {
	// 数据访问层依赖 - 负责与数据库交互
	Repositories *Repositories
	
	// 业务逻辑层依赖 - 封装核心业务规则
	Services *Services
	
	// 表现层依赖 - 处理HTTP请求和响应
	Handlers *Handlers
	
	// 应用配置 - 全局配置信息
	Config *config.AppConfig
	
	// 基础设施 - 提供底层支持
	Infrastructure struct {
		DB        *gorm.DB
		Redis     *redis.Client
		Cache     cache.Cache
		Validator *validator.Validate
		Logger    zerolog.Logger
	}
}

// NewDependencies 初始化依赖容器
// 遵循依赖倒置原则，按照从下往上的顺序初始化各层组件:
// 基础设施 -> 仓库层 -> 服务层 -> 处理器层
func NewDependencies(
	db *gorm.DB,                  // 数据库连接
	rdb *redis.Client,            // Redis客户端
	validate *validator.Validate, // 验证器
	config *config.AppConfig,     // 应用配置
	cacheInstance cache.Cache,    // 缓存实例
	logger zerolog.Logger,        // 日志记录器
) *Dependencies {
	// 创建依赖容器
	deps := &Dependencies{
		Config: config,
		Infrastructure: struct {
			DB        *gorm.DB
			Redis     *redis.Client
			Cache     cache.Cache
			Validator *validator.Validate
			Logger    zerolog.Logger
		}{
			DB:        db,
			Redis:     rdb,
			Cache:     cacheInstance,
			Validator: validate,
			Logger:    logger,
		},
	}
	
	// 1. 初始化仓库层依赖 - 数据访问层
	deps.Repositories = InitRepositories(db)
	
	// 2. 初始化服务层依赖 - 业务逻辑层
	deps.Services = InitServices(deps.Repositories, validate, db, config, cacheInstance)
	
	// 3. 初始化处理器层依赖 - 表现层
	deps.Handlers = InitHandlers(deps.Services, logger, validate)

	// 返回组装好的依赖容器
	return deps
} 
