# go-rest-starter

> golang restful api starter with std lib net/http, go-chi, gorm, postgres, redis

## 目录

```md
project-root/
├── api/                          # API 相关文件（如接口定义）
│   └── v1/                       # API 版本
│       └── router.go             # API 路由定义
│       └── swagger.go            # Swagger文档路由
│       └── dto/                  # 数据传输对象
│           └── user_dto.go       # 用户相关DTO
├── cmd/                          # 主程序入口
│   └── app/                      # 应用程序
│       └── main.go               # 程序入口
├── configs/                      # 配置文件（如 YAML, JSON）
│   └── config.yaml               # 配置示例
├── deploy/                       # 部署脚本和相关配置
│   └── docker/                   # Docker 相关配置
│       └── Dockerfile            # Docker 文件
│   └── k8s/                      # Kubernetes 配置文件
│       └── deployment.yaml       # 部署配置
├── docs/                         # Swagger文档自动生成目录
├── internal/                     # 内部应用代码，不对外暴露
│   ├── app/                      # 应用核心逻辑
│   │   ├── config/               # 业务配置
│   │   ├── db/                   # 数据连接相关
│   │   ├── handlers/             # 业务handler
│   │   ├── middleware/           # 中间件
│   │   ├── models/               # 数据模型
│   │   ├── repository/           # 数据访问层
│   │   └── services/             # 业务服务层
│   └── pkg/                      # 业务的工具类和辅助函数
│       └── errors/               # 自定义错误类型
├── migrations/                   # 数据库迁移文件
│   └── 0001_init.up.sql          # 示例迁移文件
├── pkg/                          # 外部可脱离项目使用的包
│   └── utils                     # utils函数
├── scripts/                      # 脚本文件，包含初始化，ci/cd
│   └── swagger.sh                # Swagger文档生成脚本
│   └── dev.sh                    # 开发环境运行脚本
├── test/                         # 测试相关
│   └── integration/              # 集成测试
│   └── unit/                     # 单元测试
├── tmp/                          # 临时文件目录(不提交到版本库)
├── .air.toml                     # air 配置文件(热重载工具)
├── go.mod                        # Go module 文件
├── go.sum                        # Go module 校验和
├── README.md                     # 项目说明文件
└── .gitignore                    # Git 忽略文件

```

## 技术选型

- core: `chi net/http`
- Database: `postgres`
- Cache: `redis`
- ORM: `gorm`
- Logger: `zerolog`
- Config: `viper`
- Test: `testify+gomock`(可能)
- 文档: `swagger`

## 项目特性

- 分层架构: 严格的分层 Handler -> Service -> Repository 模式
- RESTful API: 符合 RESTful 规范的 API 设计
- 自定义错误处理: 统一的错误处理机制
- 中间件支持: 日志、认证、跨域等
- 配置管理: 基于 viper 的配置管理
- 数据验证: 请求数据验证
- Swagger 文档: 自动生成 API 文档

## 快速开始

### 安装依赖

```bash
# 安装Swagger工具
go install github.com/swaggo/swag/cmd/swag@latest

# 安装热重载工具
go install github.com/cosmtrek/air@latest

# 下载项目依赖
go mod download
```

### 生成Swagger文档

```bash
# 使用脚本生成
./scripts/swagger.sh
```

### 开发模式运行

```bash
# 运行开发服务器(自动热重载)
./scripts/dev.sh
```

### 访问Swagger文档

启动服务后，访问 http://localhost:7001/swagger 查看API文档。

## 构建与部署

### 构建二进制

```bash
go build -o app cmd/app/main.go
```

### Docker构建

```bash
docker build -t go-rest-starter -f deploy/docker/Dockerfile .
```

### 运行容器

```bash
docker run -p 8080:8080 go-rest-starter
```

## 项目结构说明

### API层 (api/)

处理HTTP请求和响应的转换，包括路由定义、请求验证和响应格式化。

- DTO (Data Transfer Object) 负责请求和响应的序列化/反序列化
- 路由定义清晰分组，便于维护和扩展

### 处理层 (internal/app/handlers/)

负责接收HTTP请求，调用相应的服务层方法，并构建HTTP响应。

- 处理参数提取和基本验证
- 不包含业务逻辑，只负责协调

### 服务层 (internal/app/services/)

实现业务逻辑，包括事务管理，不直接处理HTTP请求。

- 包含核心业务规则和流程
- 处理跨仓库的数据操作和事务
- 实现接口，便于单元测试和依赖注入

### 数据访问层 (internal/app/repository/)

处理数据的读写操作，与数据库交互。

- 封装所有数据库操作
- 提供接口，便于模拟测试
- 处理基本的数据转换

### 模型层 (internal/app/models/)

定义数据结构和业务实体。

- 包含数据库模型定义
- 定义模型间的关系
- 包含基本的模型方法

### 中间件 (internal/app/middleware/)

提供请求处理前后的通用功能，如日志记录、认证和授权等。

- 处理横切关注点
- 提供请求上下文增强
- 实现安全相关功能

## API文档

API文档使用Swagger自动生成，可通过访问 `/swagger` 路径查看。

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

MIT
