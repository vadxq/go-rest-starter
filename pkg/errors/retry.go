package errors

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts     int           // 最大重试次数
	InitialDelay    time.Duration // 初始延迟
	MaxDelay        time.Duration // 最大延迟
	Multiplier      float64       // 延迟倍数
	RandomizeFactor float64       // 随机因子（0-1之间）
	RetryIf         func(error) bool // 判断是否需要重试的函数
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:     3,
	InitialDelay:    100 * time.Millisecond,
	MaxDelay:        10 * time.Second,
	Multiplier:      2.0,
	RandomizeFactor: 0.1,
	RetryIf:         IsRetryable,
}

// RetryableFunc 可重试的函数类型
type RetryableFunc func() error

// RetryableWithContextFunc 带上下文的可重试函数类型
type RetryableWithContextFunc func(context.Context) error

// Retry 执行带重试的函数
func Retry(fn RetryableFunc, config *RetryConfig) error {
	return RetryWithContext(context.Background(), func(ctx context.Context) error {
		return fn()
	}, config)
}

// RetryWithContext 执行带上下文和重试的函数
func RetryWithContext(ctx context.Context, fn RetryableWithContextFunc, config *RetryConfig) error {
	if config == nil {
		config = &DefaultRetryConfig
	}

	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// 检查上下文是否已取消
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		// 执行函数
		err := fn(ctx)
		if err == nil {
			return nil // 成功
		}

		lastErr = err

		// 检查是否应该重试
		if config.RetryIf != nil && !config.RetryIf(err) {
			return err // 不可重试的错误
		}

		// 如果是最后一次尝试，直接返回错误
		if attempt == config.MaxAttempts-1 {
			break
		}

		// 计算延迟时间
		delay := calculateDelay(attempt, config)

		// 等待或直到上下文取消
		select {
		case <-time.After(delay):
			// 继续下一次重试
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		}
	}

	return &RetryError{
		LastError: lastErr,
		Attempts:  config.MaxAttempts,
	}
}

// calculateDelay 计算重试延迟
func calculateDelay(attempt int, config *RetryConfig) time.Duration {
	// 指数退避
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))
	
	// 添加随机抖动
	if config.RandomizeFactor > 0 {
		randomFactor := 1.0 + (rand.Float64()*2-1)*config.RandomizeFactor
		delay *= randomFactor
	}
	
	// 确保不超过最大延迟
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}
	
	return time.Duration(delay)
}

// RetryError 重试错误
type RetryError struct {
	LastError error
	Attempts  int
}

func (e *RetryError) Error() string {
	return fmt.Sprintf("operation failed after %d attempts: %v", e.Attempts, e.LastError)
}

func (e *RetryError) Unwrap() error {
	return e.LastError
}

// IsRetryable 判断错误是否可重试
func IsRetryable(err error) bool {
	// 如果是HTTP状态码相关的错误，根据状态码判断
	// 5xx错误、429、408可以重试
	// 这里简化处理，实际使用时可根据具体错误类型判断
	
	// 默认某些错误可重试
	return false
}

// ExponentialBackoff 指数退避重试
func ExponentialBackoff(fn RetryableFunc) error {
	config := &RetryConfig{
		MaxAttempts:     5,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        30 * time.Second,
		Multiplier:      2.0,
		RandomizeFactor: 0.2,
		RetryIf:         IsRetryable,
	}
	return Retry(fn, config)
}

// LinearBackoff 线性退避重试
func LinearBackoff(fn RetryableFunc) error {
	config := &RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    1 * time.Second,
		MaxDelay:        5 * time.Second,
		Multiplier:      1.0,
		RandomizeFactor: 0,
		RetryIf:         IsRetryable,
	}
	return Retry(fn, config)
}

// RetryWithFixedDelay 固定延迟重试
func RetryWithFixedDelay(fn RetryableFunc, delay time.Duration, attempts int) error {
	config := &RetryConfig{
		MaxAttempts:     attempts,
		InitialDelay:    delay,
		MaxDelay:        delay,
		Multiplier:      1.0,
		RandomizeFactor: 0,
		RetryIf:         IsRetryable,
	}
	return Retry(fn, config)
}

// CircuitBreaker 断路器
type CircuitBreaker struct {
	maxFailures      int
	resetTimeout     time.Duration
	halfOpenRequests int
	
	failures         int
	lastFailureTime  time.Time
	state            CircuitState
}

// CircuitState 断路器状态
type CircuitState int

const (
	// StateClosed 关闭状态（正常）
	StateClosed CircuitState = iota
	// StateOpen 打开状态（熔断）
	StateOpen
	// StateHalfOpen 半开状态（尝试恢复）
	StateHalfOpen
)

// NewCircuitBreaker 创建断路器
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:      maxFailures,
		resetTimeout:     resetTimeout,
		halfOpenRequests: 1,
		state:           StateClosed,
	}
}

// Execute 执行函数（带断路器保护）
func (cb *CircuitBreaker) Execute(fn RetryableFunc) error {
	// 检查断路器状态
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) < cb.resetTimeout {
			return &CircuitOpenError{
				ResetAt: cb.lastFailureTime.Add(cb.resetTimeout),
			}
		}
		// 尝试进入半开状态
		cb.state = StateHalfOpen
		cb.halfOpenRequests = 1
	}

	// 执行函数
	err := fn()
	
	if err != nil {
		cb.recordFailure()
		return err
	}
	
	cb.recordSuccess()
	return nil
}

// recordFailure 记录失败
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailureTime = time.Now()
	
	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
	}
}

// recordSuccess 记录成功
func (cb *CircuitBreaker) recordSuccess() {
	if cb.state == StateHalfOpen {
		cb.halfOpenRequests--
		if cb.halfOpenRequests <= 0 {
			cb.state = StateClosed
			cb.failures = 0
		}
	}
}

// CircuitOpenError 断路器打开错误
type CircuitOpenError struct {
	ResetAt time.Time
}

func (e *CircuitOpenError) Error() string {
	return fmt.Sprintf("circuit breaker is open, will reset at %v", e.ResetAt)
}