# go-rest-starter

> golang restful api starter with std lib net/http, go-chi, gorm, postgres, redis

**简体中文** | [English](./README.md)

## 目录

```md
project-root/
├── api/                          # API 相关文件
│   └── v1/                       # API 版本
│       └── router.go             # API 路由定义
│       └── swagger.go            # Swagger文档路由
│       └── dto/                  # 数据传输对象
├── cmd/                          # 主程序入口
│   └── app/                      # 应用程序
│       └── main.go               # 程序入口
├── configs/                      # 配置文件（已优化为单一配置源）
├── deploy/                       # 部署配置（简化为必要的部署脚本）
│   └── docker/                  
│   └── k8s/                     
├── internal/                     # 内部应用代码
│   ├── app/                      # 应用核心逻辑
│   │   ├── config/               # 业务配置（优化配置结构）
│   │   ├── db/                   # 数据连接（简化为单一入口）
│   │   ├── handlers/             # 业务handler（简化处理逻辑）
│   │   ├── injection/            # 依赖注入（优化为更轻量的DI）
│   │   ├── middleware/           # 中间件（按功能分组）
│   │   ├── models/               # 数据模型（精简为核心字段）
│   │   ├── repository/           # 数据访问层（统一接口）
│   │   └── services/             # 业务服务层（清晰的职责划分）
│   └── pkg/                      # 内部通用工具
│       └── errors/               # 增强的错误处理
├── migrations/                   # 数据库迁移文件（版本化管理）
├── pkg/                          # 外部包（独立可复用的组件）
│   └── utils                     # 通用工具函数
├── scripts/                      # 开发和部署脚本（简化流程）
├── .air.toml                     # 开发热重载配置
├── go.mod                        # Go模块定义
└── README.md                     # 项目说明
```

## 快速开始

### 安装依赖

```bash
# 下载项目依赖
go mod download
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

### Docker构建与运行

```bash
# 构建镜像
docker build -t go-rest-starter -f deploy/docker/Dockerfile .

# 运行容器
docker run -p 7001:7001 go-rest-starter
```

## 技术选型

- Core: `chi net/http`
- Database: `postgres`
- Cache: `redis`
- ORM: `gorm`
- Logger: `zerolog`
- Config: `viper`
- Test: `testify`
- 文档: `swagger`

## 项目特性

- **精简分层**: Handler -> Service -> Repository 更清晰的职责划分
- **更好的错误处理**: 上下文感知的错误处理机制
- **优化的中间件**: 合理分组，减少性能开销
- **模块化配置**: 简化配置项，按功能分组
- **标准化响应**: 统一的API响应格式
- **增强安全性**: 精简而全面的安全防护
- **自动文档**: 集成Swagger的API文档

## 许可证

MIT
