# Go-Rest-Starter - Project Summary for Claude AI

## 📋 Project Overview

**Go-Rest-Starter** is a production-ready RESTful API boilerplate built with Go, implementing clean architecture principles and enterprise-grade features. The project serves as a foundation for building scalable web services with comprehensive authentication, user management, caching, and monitoring capabilities.

**Repository**: https://github.com/vadxq/go-rest-starter
**Language**: Go 1.24+
**Architecture**: Clean Architecture (Handler → Service → Repository)

## 🏗️ Architecture & Structure

### Directory Layout
```
go-rest-starter/
├── cmd/app/                    # Application entry point (main.go)
├── internal/app/               # Private application code
│   ├── config/                 # Configuration management
│   ├── db/                     # Database connections
│   ├── dto/                    # Data Transfer Objects
│   ├── handlers/               # HTTP handlers (presentation layer)
│   ├── services/               # Business logic layer
│   ├── repository/             # Data access layer
│   ├── models/                 # Domain models
│   ├── middleware/             # HTTP middleware
│   ├── injection/              # Dependency injection
│   └── router/                 # API routing
├── pkg/                        # Reusable packages
│   ├── cache/                  # Caching abstractions
│   ├── errors/                 # Error handling utilities
│   ├── jwt/                    # JWT utilities
│   ├── logger/                 # Structured logging
│   ├── queue/                  # Message queue management
│   ├── transaction/            # Transaction management
│   └── utils/                  # Common utilities
├── api/app/                    # API documentation (Swagger)
├── configs/                    # Configuration files
├── deploy/                     # Deployment configurations
└── migrations/                 # Database migrations
```

### Clean Architecture Layers
1. **Handler Layer** - HTTP request handling, response formatting, input validation
2. **Service Layer** - Business logic, transaction orchestration, caching
3. **Repository Layer** - Data access, database operations, query building
4. **Domain Layer** - Business entities and models

## 🚀 Core Features

### Authentication & Authorization
- **JWT-based Authentication** - Stateless authentication with access/refresh tokens
- **Token Management** - Token blacklisting, caching, automatic refresh
- **Role-Based Access Control** - Admin/User roles with route protection
- **Secure Password Handling** - bcrypt hashing with configurable cost

### User Management
- **Complete CRUD Operations** - Create, read, update, delete users
- **Pagination & Filtering** - Efficient user listing with search capabilities
- **Email Validation** - Unique email constraints and format validation
- **Role Management** - User role assignment and permission control

### Security & Middleware
- **Security Headers** - CSP, HSTS, X-Frame-Options, XSS protection
- **CORS Support** - Configurable cross-origin resource sharing
- **Rate Limiting** - IP-based request throttling with automatic cleanup
- **Input Validation** - Comprehensive request validation with custom errors
- **Request Tracing** - Trace IDs and request IDs for debugging

### Performance & Scalability
- **Redis Caching** - High-performance caching with TTL management
- **Connection Pooling** - Database and Redis connection optimization
- **Message Queues** - Redis-based pub/sub messaging with worker pools
- **Transaction Management** - GORM transactions with nested support
- **Graceful Shutdown** - Zero-downtime deployment support

### Monitoring & Health Checks
- **Multiple Health Endpoints** - Basic, detailed, readiness, liveness probes
- **System Metrics** - CPU, memory, goroutine monitoring
- **Dependency Monitoring** - Database and Redis status checks
- **Performance Tracking** - Request counting, error rates, QPS metrics
- **Kubernetes Ready** - Built-in K8s health probes

## 🛠️ Technology Stack

### Core Dependencies (from go.mod)
```go
// Web Framework & Routing
github.com/go-chi/chi/v5 v5.2.1

// Database & ORM
gorm.io/gorm v1.26.1
gorm.io/driver/postgres v1.5.11

// Authentication & Security
github.com/golang-jwt/jwt/v5 v5.2.2
golang.org/x/crypto v0.38.0
golang.org/x/time v0.12.0

// Caching & Redis
github.com/redis/go-redis/v9 v9.8.0

// Configuration & Validation
github.com/spf13/viper v1.20.1
github.com/go-playground/validator/v10 v10.26.0

// Testing & Documentation
github.com/stretchr/testify v1.10.0
github.com/swaggo/swag v1.16.4
github.com/swaggo/http-swagger/v2 v2.0.2
```

## 📚 API Endpoints

### Public Routes (No Authentication)
- `POST /api/v1/auth/login` - User authentication
- `POST /api/v1/auth/refresh` - JWT token refresh
- Health check endpoints (`/health`, `/health/detailed`, `/ready`, `/live`)

### Protected Routes (JWT Required)
- `POST /api/v1/account/logout` - User logout (token invalidation)
- `GET /api/v1/users` - List users with pagination
- `POST /api/v1/users` - Create user (Admin only)
- `GET /api/v1/users/{id}` - Get user details
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user (Admin only)

### System Endpoints
- `GET /version` - API version information
- `GET /status` - Service status
- `GET /metrics` - Performance metrics
- `GET /swagger/*` - API documentation UI

## ⚙️ Configuration

### Configuration Files
- `configs/config.example.yaml` - Example configuration template
- `configs/config.yaml` - Main configuration (created from example)
- `configs/config.production.yaml` - Production overrides

### Environment Variables
All configuration values support environment variable overrides with `APP_` prefix:
- `APP_SERVER_PORT` - Server port (default: 7001)
- `APP_DATABASE_*` - Database configuration
- `APP_REDIS_*` - Redis configuration
- `APP_JWT_*` - JWT settings
- `APP_LOG_*` - Logging configuration

## 🗄️ Database

### Models
- **User Model** - Core user entity with email, password, role fields
- **Migration Support** - SQL migration files in `migrations/` directory
- **Connection Management** - Configurable connection pooling and timeouts

### Repository Pattern
- Interface-based repository design for testability
- GORM integration with automatic migrations
- Transaction support with rollback capabilities

## 🔧 Development & Testing

### Development Scripts
- `./scripts/dev.sh` - Development server with auto-reload
- `./scripts/setup.sh` - Initial project setup
- `./scripts/swagger.sh` - Generate API documentation

### Testing Support
- Unit test foundation with testify framework
- Service layer testing with mocks
- Integration test support (tagged builds)
- Benchmark testing capabilities

### Build & Deployment
- Docker support with multi-stage builds
- Kubernetes deployment configurations
- Docker Compose for local development
- Production-ready binary builds

## 🔍 Key Implementation Details

### Dependency Injection
- Comprehensive DI container in `internal/app/injection/`
- Interface-based design for all major components
- Lifecycle management for application components

### Error Handling
- Custom error types with HTTP status mapping
- Structured error responses with context
- Panic recovery with graceful degradation

### Logging
- Structured logging with Go's built-in `slog`
- Request tracing with correlation IDs
- Configurable log levels and outputs
- File rotation and console logging support

### Caching Strategy
- Redis-only caching (no memory cache to avoid leaks)
- TTL-based expiration with smart invalidation
- Object serialization with JSON marshaling
- Cache-aside pattern implementation

### Security Measures
- JWT token blacklisting with Redis storage
- Password hashing with bcrypt and salt
- Security headers for common vulnerabilities
- Input sanitization and validation

## 🚦 Common Development Tasks

### Adding New Endpoints
1. Define DTOs in `internal/app/dto/`
2. Create handler methods in `internal/app/handlers/`
3. Add business logic in `internal/app/services/`
4. Implement data access in `internal/app/repository/`
5. Register routes in `internal/app/router/`
6. Add Swagger annotations for documentation

### Database Changes
1. Create migration files in `migrations/app/`
2. Update models in `internal/app/models/`
3. Modify repository interfaces and implementations
4. Update service layer logic if needed

### Adding Middleware
1. Create middleware in `internal/app/middleware/`
2. Register in router configuration
3. Add configuration options if needed

## 📊 Performance Characteristics

### Optimizations
- Connection pooling for database and Redis
- Efficient pagination with offset/limit
- Caching for frequently accessed data
- Structured logging for minimal overhead

### Scalability Considerations
- Stateless design for horizontal scaling
- Redis for session and cache storage
- Health checks for load balancer integration
- Graceful shutdown for rolling deployments

## 🎯 Best Practices Implemented

### Code Quality
- Clean architecture separation
- Interface-driven development
- Comprehensive error handling
- Input validation on all endpoints
- Structured logging throughout

### Security
- JWT-based stateless authentication
- Role-based access control
- Security headers implementation
- Input sanitization and validation
- Secure password handling

### Operations
- Health check endpoints
- Metrics collection
- Configuration management
- Graceful shutdown
- Docker containerization

This project serves as an excellent foundation for building production-ready REST APIs in Go, with comprehensive features and enterprise-grade architecture patterns.