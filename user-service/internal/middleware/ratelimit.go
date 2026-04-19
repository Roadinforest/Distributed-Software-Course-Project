package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiterConfig 限流器配置
type RateLimiterConfig struct {
	RequestsPerSecond int           // 每秒允许的请求数
	BurstSize         int           // 突发容量
	CleanupInterval   time.Duration // 清理间隔
}

// DefaultRateLimiterConfig 默认限流配置
var DefaultRateLimiterConfig = RateLimiterConfig{
	RequestsPerSecond: 100,
	BurstSize:         200,
	CleanupInterval:   5 * time.Minute,
}

// TokenBucket 令牌桶
type TokenBucket struct {
	tokens     float64
	maxTokens  float64
	rate        float64
	lastUpdate  time.Time
	mu          sync.Mutex
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	return &TokenBucket{
		tokens:    float64(burst),
		maxTokens: float64(burst),
		rate:      rate,
		lastUpdate: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.lastUpdate = now

	// 添加令牌
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// IPRateLimiter 基于IP的限流器
type IPRateLimiter struct {
	limiters map[string]*TokenBucket
	mu       sync.RWMutex
	config   RateLimiterConfig
}

// NewIPRateLimiter 创建IP限流器
func NewIPRateLimiter(cfg RateLimiterConfig) *IPRateLimiter {
	rl := &IPRateLimiter{
		limiters: make(map[string]*TokenBucket),
		config:   cfg,
	}

	// 启动清理协程
	go rl.cleanup()

	return rl
}

// GetLimiter 获取IP对应的限流器
func (rl *IPRateLimiter) GetLimiter(ip string) *TokenBucket {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查
	if limiter, exists = rl.limiters[ip]; exists {
		return limiter
	}

	limiter = NewTokenBucket(float64(rl.config.RequestsPerSecond), rl.config.BurstSize)
	rl.limiters[ip] = limiter
	return limiter
}

// cleanup 定期清理过期的限流器
func (rl *IPRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, limiter := range rl.limiters {
			limiter.mu.Lock()
			if now.Sub(limiter.lastUpdate) > rl.config.CleanupInterval*2 {
				delete(rl.limiters, ip)
			}
			limiter.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware 创建限流中间件
func RateLimitMiddleware(cfg RateLimiterConfig) gin.HandlerFunc {
	limiter := NewIPRateLimiter(cfg)

	return func(c *gin.Context) {
		ip := c.ClientIP()
	bucket := limiter.GetLimiter(ip)

		if !bucket.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "rate limit exceeded, please try again later",
				"retry_after": "1s",
			})
			return
		}

		c.Next()
	}
}

// GlobalRateLimitMiddleware 全局限流中间件实例
var GlobalRateLimitMiddleware gin.HandlerFunc

func init() {
	GlobalRateLimitMiddleware = RateLimitMiddleware(DefaultRateLimiterConfig)
}
