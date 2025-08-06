# Go-Rest-Starter

> Production-ready Go RESTful API boilerplate with Chi, GORM, PostgreSQL, Redis, and enterprise features

[English](./README.md) | [中文](./README.zh_CN.md)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/vadxq/go-rest-starter)](https://goreportcard.com/report/github.com/vadxq/go-rest-starter)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)

## 🌟 Features

### 🚀 Core Features
- **🏭 Clean Architecture** - Three-layer architecture (Repository/Service/Handler) with comprehensive dependency injection
- **🔒 JWT Authentication** - Complete authentication system with access/refresh tokens and token blacklisting
- **👥 User Management** - Full CRUD operations with role-based access control (Admin/User roles)
- **📝 Structured Logging** - Advanced logging with trace ID, request ID, and context propagation using Go's slog
- **🚫 Rate Limiting** - IP-based request throttling with automatic cleanup
- **📊 Health Monitoring** - Comprehensive health checks with dependency monitoring and system metrics
- **🌐 Redis Cache** - Production-ready caching layer with TTL management and object serialization
- **📦 Message Queue** - Redis-based pub/sub messaging with worker pools and dead letter queue support
- **💼 Transaction Management** - GORM transaction manager with nested transaction support
- **🛡️ Security** - Multiple security layers including CORS, security headers, and input validation

### 🛠️ Middleware Stack
- **Request Context** - Trace IDs, request IDs, and user context propagation
- **Security Headers** - CSP, HSTS, X-Frame-Options, XSS Protection
- **CORS Handling** - Configurable cross-origin resource sharing
- **Panic Recovery** - Application-level panic handling with graceful error responses
- **Request Logging** - Structured request/response logging with performance metrics
- **Authentication** - JWT middleware with role-based route protection
- **Input Validation** - Comprehensive request validation using go-playground/validator

### 📈 Health & Monitoring
- **Health Endpoints** - Basic, detailed, readiness, and liveness probes
- **System Metrics** - CPU, memory, goroutine monitoring
- **Dependency Checks** - Database and Redis connection status monitoring
- **Performance Tracking** - Request counting, error tracking, and QPS monitoring
- **Kubernetes Ready** - Built-in K8s readiness and liveness probes

## Directory Structure

Design reference:

- [go project layout](https://github.com/golang-standards/project-layout)
- [go modules layout](https://go.dev/doc/modules/layout)

```md
project-root/
├── api/                          # API related files
│   └── app/                      # API app docs
│       └── docs.go               # docs.go
│       └── swagger.json          # Swagger documentation
├── cmd/                          # Main program entry
│   └── app/                      # Application
│       └── main.go               # Program entry point
├── configs/                      # Configuration files (optimized as single source)
├── deploy/                       # Deployment configurations (simplified to essential scripts)
│   └── docker/                  
│   └── k8s/                     
├── internal/                     # Internal application code
│   └── app/                      # Core application logic
│       ├── config/               # Business configurations (optimized structure)
│       ├── db/                   # Database connections (simplified to single entry)
│       ├── dto/                  # Data Transfer Objects
│       ├── handlers/             # Business handlers (simplified logic)
│       ├── injection/            # Dependency injection (optimized for lightweight DI)
│       ├── middleware/           # Middleware (grouped by functionality)
│       ├── models/               # Data models (streamlined core fields)
│       ├── repository/           # Data access layer (unified interface)
│       ├── router/               # API router
│       └── services/             # Business service layer (clear responsibility division)
├── migrations/                   # Database migration files (version controlled)
├── pkg/                          # External packages (independent reusable components)
│   ├── errors/                   # Custom error handling package
│   └── utils                     # Common utility functions
├── scripts/                      # Development and deployment scripts (simplified workflow)
├── .air.toml                     # Development hot-reload configuration
├── go.mod                        # Go module definition
└── README.md                     # Project documentation
```

## 🚀 Quick Start

### Prerequisites
- **Go 1.24+** - [Install Go](https://golang.org/doc/install)
- **PostgreSQL 12+** - [Install PostgreSQL](https://postgresql.org/download/)
- **Redis 6+** - [Install Redis](https://redis.io/download)

### Installation

```bash
# Clone the repository
git clone https://github.com/vadxq/go-rest-starter.git
cd go-rest-starter

# Install dependencies
go mod download

# Copy and configure the config file
cp configs/config.example.yaml configs/config.yaml
# Edit configs/config.yaml with your database and Redis settings
```

### Development Mode

```bash
# Run development server (with auto-reload)
./scripts/dev.sh

# Or run directly
go run cmd/app/main.go
```

### Access API Documentation

After starting the service, visit **http://localhost:7001/swagger** to view the interactive API documentation.

## 📚 API Endpoints

### 🏥 Health Check Endpoints
- `GET /health` - Basic health check with uptime
- `GET /health/detailed` - Detailed health check (includes DB and Redis status)
- `GET /health/ready` - Kubernetes readiness probe
- `GET /health/live` - Kubernetes liveness probe
- `GET /health/system` - System metrics (CPU, memory, goroutines)
- `GET /health/dependencies` - Dependency services status

### 🔐 Authentication Endpoints (Public)
- `POST /api/v1/auth/login` - User authentication
- `POST /api/v1/auth/refresh` - Refresh JWT token

### 🔒 Account Management Endpoints (Protected)
- `POST /api/v1/account/logout` - User logout (invalidates tokens)

### 👥 User Management Endpoints (Protected)
- `GET /api/v1/users` - List users with pagination and filtering
- `POST /api/v1/users` - Create new user (Admin only)
- `GET /api/v1/users/{id}` - Get user details by ID
- `PUT /api/v1/users/{id}` - Update user information
- `DELETE /api/v1/users/{id}` - Delete user (Admin only)

### 📊 System Endpoints
- `GET /version` - API version information
- `GET /status` - Service status and configuration
- `GET /metrics` - Application performance metrics

## ⚙️ Configuration

### Configuration Files
The application uses YAML configuration files with environment variable override support:

```bash
# Primary configuration
configs/config.yaml          # Main configuration (create from example)
configs/config.example.yaml  # Example configuration template
configs/config.production.yaml  # Production-specific overrides
```

### Environment Variables
All configuration values can be overridden using environment variables with `APP_` prefix:

```bash
# Server Configuration
APP_SERVER_PORT=7001
APP_SERVER_TIMEOUT=30s
APP_SERVER_READ_TIMEOUT=15s
APP_SERVER_WRITE_TIMEOUT=15s

# Database Configuration
APP_DATABASE_HOST=localhost
APP_DATABASE_PORT=5432
APP_DATABASE_USERNAME=postgres
APP_DATABASE_PASSWORD=your-password
APP_DATABASE_DBNAME=myapp
APP_DATABASE_SSLMODE=disable
APP_DATABASE_MAX_OPEN_CONNS=20
APP_DATABASE_MAX_IDLE_CONNS=5
APP_DATABASE_CONN_MAX_LIFETIME=1h

# Redis Configuration
APP_REDIS_HOST=localhost
APP_REDIS_PORT=6379
APP_REDIS_PASSWORD=""
APP_REDIS_DB=0

# JWT Configuration
APP_JWT_SECRET=your-secure-secret-key-change-in-production
APP_JWT_ACCESS_TOKEN_EXP=24h
APP_JWT_REFRESH_TOKEN_EXP=168h
APP_JWT_ISSUER=go-rest-starter

# Logging Configuration
APP_LOG_LEVEL=info
APP_LOG_FILE=logs/app.log
APP_LOG_CONSOLE=true
```

### Configuration Structure
```yaml
app:
  server:
    port: 7001
    timeout: 30s
    read_timeout: 15s
    write_timeout: 15s
  database:
    driver: postgres
    host: localhost
    port: 5432
    # ... (see config.example.yaml for full structure)
```

## 🚀 Build and Deploy

### Build Binary

```bash
# Build for current platform
go build -o app cmd/app/main.go

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o app cmd/app/main.go

# Build with version info
go build -ldflags="-s -w" -o app cmd/app/main.go
```

### Docker Deployment

```bash
# Build Docker image
docker build -t go-rest-starter -f deploy/docker/Dockerfile .

# Run with Docker Compose (includes PostgreSQL and Redis)
cd deploy/docker
docker-compose up -d

# Run container only (requires external database and Redis)
docker run -p 7001:7001 \
  -e APP_DATABASE_HOST=your-db-host \
  -e APP_REDIS_HOST=your-redis-host \
  go-rest-starter
```

### Kubernetes Deployment

```bash
# Deploy to Kubernetes
kubectl apply -f deploy/k8s/

# Check deployment status
kubectl get pods -l app=go-rest-starter
kubectl logs -f deployment/go-rest-starter
```

## 🧪 Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage report in browser
go tool cover -html=coverage.out

# Run tests with race detection
go test -race ./...

# Run tests with verbose output
go test -v ./...
```

### Test Categories

```bash
# Unit tests (services layer)
go test ./internal/app/services/

# Integration tests (if available)
go test -tags=integration ./...

# Benchmark tests
go test -bench=. ./...
```

## 🛠️ Technology Stack

### Core Framework & Libraries
- **Web Framework**: `chi/v5` - Lightweight, fast HTTP router with middleware support
- **ORM**: `GORM v1.26.1` - Feature-rich ORM with auto-migration and relations
- **Database Driver**: `gorm.io/driver/postgres` - PostgreSQL driver for GORM
- **Cache**: `redis/go-redis/v9` - Redis client with pipeline and pub/sub support

### Authentication & Security
- **JWT**: `golang-jwt/jwt/v5` - JSON Web Token implementation
- **Password Hashing**: `golang.org/x/crypto/bcrypt` - Secure password hashing
- **Input Validation**: `go-playground/validator/v10` - Struct validation with tags
- **Rate Limiting**: `golang.org/x/time/rate` - Token bucket rate limiting

### Configuration & Utilities
- **Configuration**: `spf13/viper` - Configuration management (YAML, ENV, JSON)
- **Logging**: `log/slog` - Structured logging (Go 1.21+ built-in)
- **Testing**: `stretchr/testify` - Testing toolkit with assertions and mocks

### Documentation & Development
- **API Documentation**: `swaggo/swag` - Swagger/OpenAPI 3.0 documentation generator
- **HTTP Swagger UI**: `swaggo/http-swagger/v2` - Swagger UI integration

## 🌟 Architecture & Design

### Clean Architecture Implementation
- **Handler Layer** - HTTP request handling and response formatting
- **Service Layer** - Business logic and transaction management
- **Repository Layer** - Data access and database operations
- **Dependency Injection** - Interface-based design with comprehensive DI container

### Key Design Patterns
- **Repository Pattern** - Abstract data access layer
- **Service Pattern** - Encapsulated business logic
- **Middleware Chain** - Composable request processing
- **Factory Pattern** - Component initialization
- **Observer Pattern** - Configuration watching and hot reload

### Security Features
- **JWT Authentication** - Stateless authentication with token blacklisting
- **Role-Based Access Control** - Admin/User role separation
- **Security Headers** - CSP, HSTS, X-Frame-Options, XSS protection
- **Input Validation** - Request validation with custom error messages
- **Rate Limiting** - IP-based request throttling
- **Password Security** - bcrypt hashing with configurable cost

### Performance Features
- **Connection Pooling** - Database and Redis connection management
- **Caching Layer** - Redis-based caching with TTL management
- **Structured Logging** - High-performance logging with context
- **Graceful Shutdown** - Zero-downtime deployments
- **Health Checks** - Kubernetes-ready probes

## 📝 Usage Examples

### Authentication Flow
```bash
# Login
curl -X POST http://localhost:7001/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "password"}'

# Use the returned token for authenticated requests
curl -X GET http://localhost:7001/api/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### User Management
```bash
# Create user (Admin only)
curl -X POST http://localhost:7001/api/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password", "role": "user"}'

# Get user list with pagination
curl -X GET "http://localhost:7001/api/v1/users?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Health Monitoring
```bash
# Basic health check
curl http://localhost:7001/health

# Detailed health with dependencies
curl http://localhost:7001/health/detailed

# System metrics
curl http://localhost:7001/health/system
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Go Project Layout](https://github.com/golang-standards/project-layout) - Standard Go project structure
- [Chi Router](https://github.com/go-chi/chi) - Lightweight HTTP router
- [GORM](https://gorm.io/) - The fantastic ORM library for Golang
