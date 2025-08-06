package handlers

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"log/slog"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *slog.Logger
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(db *gorm.DB, redis *redis.Client, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// HealthStatus 健康状态结构
type HealthStatus struct {
	Status     string            `json:"status"`
	Timestamp  time.Time         `json:"timestamp"`
	Services   map[string]string `json:"services"`
	Version    string            `json:"version"`
	Uptime     string            `json:"uptime,omitempty"`
}

var startTime = time.Now()

// Health 基础健康检查
// @Summary 健康检查
// @Description 检查应用基本状态
// @Tags health
// @Produce json
// @Success 200 {object} HealthStatus
// @Router /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		Services:  make(map[string]string),
	}

	RespondJSON(w, http.StatusOK, status)
}

// DetailedHealth 详细健康检查
// @Summary 详细健康检查
// @Description 检查应用及其依赖服务的详细状态
// @Tags health
// @Produce json
// @Success 200 {object} HealthStatus
// @Success 503 {object} HealthStatus "服务不可用"
// @Router /health/detailed [get]
func (h *HealthHandler) DetailedHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
		Services:  make(map[string]string),
	}

	// 检查数据库连接
	dbStatus := h.checkDatabase(ctx)
	status.Services["database"] = dbStatus

	// 检查Redis连接
	redisStatus := h.checkRedis(ctx)
	status.Services["redis"] = redisStatus

	// 确定整体状态
	if dbStatus != "healthy" || redisStatus != "healthy" {
		status.Status = "unhealthy"
		RespondJSON(w, http.StatusServiceUnavailable, status)
		return
	}

	RespondJSON(w, http.StatusOK, status)
}

// Ready 就绪检查
// @Summary 就绪检查
// @Description 检查应用是否准备好接收请求
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Success 503 {object} map[string]interface{} "服务未就绪"
// @Router /ready [get]
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	ready := true
	checks := make(map[string]interface{})

	// 检查数据库
	if h.checkDatabase(ctx) != "healthy" {
		ready = false
		checks["database"] = "not ready"
	} else {
		checks["database"] = "ready"
	}

	// 检查Redis
	if h.checkRedis(ctx) != "healthy" {
		ready = false
		checks["redis"] = "not ready"
	} else {
		checks["redis"] = "ready"
	}

	response := map[string]interface{}{
		"ready":     ready,
		"timestamp": time.Now(),
		"checks":    checks,
	}

	if ready {
		RespondJSON(w, http.StatusOK, response)
	} else {
		RespondJSON(w, http.StatusServiceUnavailable, response)
	}
}

// Live 存活检查
// @Summary 存活检查
// @Description 检查应用是否存活（基础响应能力）
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /live [get]
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"alive":     true,
		"timestamp": time.Now(),
	}
	RespondJSON(w, http.StatusOK, response)
}

// checkDatabase 检查数据库连接状态
func (h *HealthHandler) checkDatabase(ctx context.Context) string {
	if h.db == nil {
		return "unavailable"
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		h.logger.Error("获取数据库连接失败", "error", err)
		return "error"
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		h.logger.Error("数据库ping失败", "error", err)
		return "unhealthy"
	}

	return "healthy"
}

// checkRedis 检查Redis连接状态
func (h *HealthHandler) checkRedis(ctx context.Context) string {
	if h.redis == nil {
		return "unavailable"
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		h.logger.Error("Redis ping失败", "error", err)
		return "unhealthy"
	}

	return "healthy"
}

// Readiness K8s就绪探针
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	h.Ready(w, r)
}

// Liveness K8s存活探针
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	h.Live(w, r)
}

// SystemInfo 系统信息
func (h *HealthHandler) SystemInfo(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	systemInfo := map[string]interface{}{
		"runtime": map[string]interface{}{
			"version":    runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"cpu_count":  runtime.NumCPU(),
			"goos":       runtime.GOOS,
			"goarch":     runtime.GOARCH,
		},
		"memory": map[string]interface{}{
			"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
			"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
			"sys_mb":         float64(m.Sys) / 1024 / 1024,
			"num_gc":         m.NumGC,
			"heap_alloc_mb":  float64(m.HeapAlloc) / 1024 / 1024,
			"heap_sys_mb":    float64(m.HeapSys) / 1024 / 1024,
		},
		"application": map[string]interface{}{
			"version": "1.0.0",
			"uptime":  time.Since(startTime).String(),
			"started": startTime.Format(time.RFC3339),
		},
		"timestamp": time.Now().Unix(),
	}
	
	RespondJSON(w, http.StatusOK, systemInfo)
}

// CheckDependencies 检查所有依赖服务
func (h *HealthHandler) CheckDependencies(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	type DependencyStatus struct {
		Name         string        `json:"name"`
		Status       string        `json:"status"`
		ResponseTime time.Duration `json:"response_time_ms"`
		Error        string        `json:"error,omitempty"`
	}
	
	var dependencies []DependencyStatus
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// 检查数据库
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		status := "healthy"
		var errMsg string
		
		if h.db != nil {
			sqlDB, err := h.db.DB()
			if err != nil {
				status = "error"
				errMsg = err.Error()
			} else if err := sqlDB.PingContext(ctx); err != nil {
				status = "unhealthy"
				errMsg = err.Error()
			}
		} else {
			status = "unavailable"
		}
		
		mu.Lock()
		dependencies = append(dependencies, DependencyStatus{
			Name:         "postgresql",
			Status:       status,
			ResponseTime: time.Since(start) / time.Millisecond,
			Error:        errMsg,
		})
		mu.Unlock()
	}()
	
	// 检查Redis
	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		status := "healthy"
		var errMsg string
		
		if h.redis != nil {
			if err := h.redis.Ping(ctx).Err(); err != nil {
				status = "unhealthy"
				errMsg = err.Error()
			}
		} else {
			status = "unavailable"
		}
		
		mu.Lock()
		dependencies = append(dependencies, DependencyStatus{
			Name:         "redis",
			Status:       status,
			ResponseTime: time.Since(start) / time.Millisecond,
			Error:        errMsg,
		})
		mu.Unlock()
	}()
	
	wg.Wait()
	
	// 确定整体状态
	overallStatus := "healthy"
	for _, dep := range dependencies {
		if dep.Status != "healthy" {
			overallStatus = "degraded"
			if dep.Status == "unhealthy" || dep.Status == "error" {
				overallStatus = "unhealthy"
				break
			}
		}
	}
	
	response := map[string]interface{}{
		"status":       overallStatus,
		"dependencies": dependencies,
		"timestamp":    time.Now().Unix(),
	}
	
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	RespondJSON(w, statusCode, response)
}