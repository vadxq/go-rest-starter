# go-rest-starter

> golang restful api starter with std lib net/http, go-chi, gorm, postgres, redis

[English](./README.md) | [中文](./README.zh_CN.md)

## Directory Structure

```md
project-root/
├── api/                          # API related files
│   └── v1/                       # API version
│       └── router.go             # API route definitions
│       └── swagger.go            # Swagger documentation routes
│       └── dto/                  # Data Transfer Objects
├── cmd/                          # Main program entry
│   └── app/                      # Application
│       └── main.go               # Program entry point
├── configs/                      # Configuration files (optimized as single source)
├── deploy/                       # Deployment configurations (simplified to essential scripts)
│   └── docker/                  
│   └── k8s/                     
├── internal/                     # Internal application code
│   ├── app/                      # Core application logic
│   │   ├── config/               # Business configurations (optimized structure)
│   │   ├── db/                   # Database connections (simplified to single entry)
│   │   ├── handlers/             # Business handlers (simplified logic)
│   │   ├── injection/            # Dependency injection (optimized for lightweight DI)
│   │   ├── middleware/           # Middleware (grouped by functionality)
│   │   ├── models/               # Data models (streamlined core fields)
│   │   ├── repository/           # Data access layer (unified interface)
│   │   └── services/             # Business service layer (clear responsibility division)
│   └── pkg/                      # Internal common utilities
│       └── errors/               # Enhanced error handling
├── migrations/                   # Database migration files (version controlled)
├── pkg/                          # External packages (independent reusable components)
│   └── utils                     # Common utility functions
├── scripts/                      # Development and deployment scripts (simplified workflow)
├── .air.toml                     # Development hot-reload configuration
├── go.mod                        # Go module definition
└── README.md                     # Project documentation
```

## Quick Start

### Install Dependencies

```bash
# Download project dependencies
go mod download
```

### Development Mode

```bash
# Run development server (with auto-reload)
./scripts/dev.sh
```

### Access Swagger Documentation

After starting the service, visit http://localhost:7001/swagger to view the API documentation.

## Build and Deploy

### Build Binary

```bash
go build -o app cmd/app/main.go
```

### Docker Build and Run

```bash
# Build image
docker build -t go-rest-starter -f deploy/docker/Dockerfile .

# Run container
docker run -p 7001:7001 go-rest-starter
```

## Technology Stack

- Core: `chi net/http`
- Database: `postgres`
- Cache: `redis`
- ORM: `gorm`
- Logger: `zerolog`
- Config: `viper`
- Test: `testify`
- Documentation: `swagger`

## Project Features

- **Streamlined Layering**: Handler -> Service -> Repository with clearer responsibility division
- **Better Error Handling**: Context-aware error handling mechanism
- **Optimized Middleware**: Reasonable grouping, reduced performance overhead
- **Modular Configuration**: Simplified configuration items, grouped by functionality
- **Standardized Responses**: Unified API response format
- **Enhanced Security**: Streamlined yet comprehensive security measures
- **Auto Documentation**: Integrated Swagger API documentation

## License

MIT
