# 依赖注入（Dependency Injection）

本包实现了一个简单而高效的依赖注入系统，用于管理应用程序的各层依赖关系。

## 设计原则

1. **模块化**: 依赖按层次清晰划分，便于维护和扩展
2. **单一职责**: 每个注入文件只负责一层的依赖管理
3. **低耦合**: 各层之间通过接口交互，而非具体实现
4. **显式依赖**: 所有依赖关系都明确声明，没有隐藏的依赖
5. **清晰结构**: 依赖关系组织明确，避免扁平化结构导致的混乱

## 包结构

```
injection/
├── dependencies.go  # 主依赖注入入口
├── repositories.go  # 仓库层依赖
├── services.go      # 服务层依赖
└── handlers.go      # 处理器层依赖
```

## 依赖关系图

```
Handler -> Service -> Repository -> DB/Redis
```

## 依赖组织结构

```go
type Dependencies struct {
    Repositories *Repositories  // 仓库层依赖
    Services     *Services      // 服务层依赖
    Handlers     *Handlers      // 处理器层依赖
    Config       *config.AppConfig  // 配置信息
}
```

## 使用方法

在应用程序的入口点（例如main.go）中，通过以下方式初始化所有依赖：

```go
// 初始化依赖
deps := injection.NewDependencies(
    database,      // 数据库连接
    redisClient,   // Redis客户端
    validate,      // 验证器
    config,        // 应用配置
    cacheInstance, // 缓存实例
    logger,        // 日志记录器
)

// 使用初始化的依赖
api.Setup(router, api.RouterConfig{
    UserHandler: deps.Handlers.UserHandler,
    AuthHandler: deps.Handlers.AuthHandler,
    JWTSecret:   deps.Config.JWT.Secret,
})
```

## 扩展指南

### 添加新的仓库

1. 在`repository`包中定义新的仓库接口和实现
2. 在`repositories.go`中的`Repositories`结构体添加新字段
3. 在`InitRepositories`函数中初始化新的仓库实例

例如：

```go
// 定义新仓库
type ProductRepository interface {
    // 方法定义...
}

// Repositories结构体添加字段
type Repositories struct {
    UserRepo    repository.UserRepository
    ProductRepo repository.ProductRepository  // 新增字段
}

// 初始化函数中添加
func InitRepositories(db *gorm.DB) *Repositories {
    return &Repositories{
        UserRepo:    repository.NewUserRepository(db),
        ProductRepo: repository.NewProductRepository(db),  // 初始化新仓库
    }
}
```

### 添加新的服务

1. 在`services`包中定义新的服务接口和实现
2. 在`services.go`中的`Services`结构体添加新字段
3. 在`InitServices`函数中初始化新的服务实例

### 添加新的处理器

1. 在`handlers`包中定义新的处理器
2. 在`handlers.go`中的`Handlers`结构体添加新字段
3. 在`InitHandlers`函数中初始化新的处理器实例

## 最佳实践

1. 始终通过接口而非具体类型进行依赖注入
2. 保持单向依赖关系，避免循环依赖
3. 对于需要共享的配置或工具，通过依赖注入传递，而非全局变量
4. 在测试中，可以轻松替换任何依赖为mock实现
5. 保持依赖关系层次清晰，遵循依赖倒置原则 