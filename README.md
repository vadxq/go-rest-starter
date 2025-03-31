# go-rest-starter

> golang restful api starter with std lib net/http, go-chi, gorm, postgres, redis

## 目录

```md
project-root/
├── api/                          # API 相关文件（如接口定义）
│   └── v1/                       # API 版本
│       └── handler.go            # 处理请求的逻辑
│       └── router.go             # API 路由定义
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
├── migrations/                   # 数据库迁移文件
│   └── 0001_init.up.sql          # 示例迁移文件
├── pkg/                          # 外部可脱离项目使用的包
│   └── utils                     # utils函数
├── scripts/                      # 脚本文件，包含初始化，ci/cd
│   └── setup.sh                  # 设置脚本
├── test/                         # 测试相关
│   └── integration/              # 集成测试
│   └── unit/                     # 单元测试
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

## go install

```bash
# 安装二进制文件
go install xxx

# 下载包
go get xxx

# 扫描你的项目并更新文件，并下载文件 go.mod中提到的所有依赖项go.mod
go mod tidy

# 仅从 下载依赖项go.mod，而不修改go.mod文件
go mod download
```

## 运行

```bash
go run cmd/app/main.go
```
