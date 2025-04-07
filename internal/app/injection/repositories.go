package injection

import (
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/internal/app/repository"
)

// Repositories 所有仓库的集合
// 包含所有数据访问层对象，负责与数据源交互
type Repositories struct {
	// 用户数据访问对象
	UserRepo repository.UserRepository
	
	// 可以在此添加更多仓库...
	// ProductRepo repository.ProductRepository
	// OrderRepo repository.OrderRepository
}

// InitRepositories 初始化所有仓库
// 这是依赖注入的第一层，负责创建所有数据访问对象
func InitRepositories(db *gorm.DB) *Repositories {
	// 参数验证
	if db == nil {
		log.Fatal().Msg("数据库连接不能为空")
	}
	
	// 创建所有仓库实例
	userRepo := repository.NewUserRepository(db)
	
	// 返回仓库集合
	return &Repositories{
		UserRepo: userRepo,
	}
} 