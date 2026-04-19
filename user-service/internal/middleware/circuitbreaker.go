package middleware

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
)

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Name        string        // 熔断器名称
	MaxRequests uint32        // 熔断打开前的最大请求数
	Interval    time.Duration // 统计周期
	Timeout     time.Duration // 熔断开放时间
}

// DefaultCircuitBreakerConfig 默认熔断器配置
var DefaultCircuitBreakerConfig = CircuitBreakerConfig{
	Name:        "user-service",
	MaxRequests: 3,
	Interval:    10 * time.Second,
	Timeout:     30 * time.Second,
}

// NewCircuitBreaker 创建熔断器中间件
func NewCircuitBreaker(cfg CircuitBreakerConfig) gin.HandlerFunc {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 错误率超过50%或连续失败超过3次时打开熔断
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.ConsecutiveFailures >= 3 || (counts.Requests >= 10 && failureRatio >= 0.5)
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("circuit breaker state changed: name=%s from=%s to=%s", name, from, to)
		},
	})

	return func(c *gin.Context) {
		result, err := cb.Execute(func() (interface{}, error) {
			c.Next()
			if c.Writer.Status() >= 500 {
				return nil, errors.New("server error")
			}
			return nil, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "service unavailable, circuit breaker opened",
				"error":   err.Error(),
			})
			return
		}

		_ = result
	}
}

// CircuitBreakerMiddleware 全局熔断中间件实例
var CircuitBreakerMiddleware gin.HandlerFunc

func init() {
	CircuitBreakerMiddleware = NewCircuitBreaker(DefaultCircuitBreakerConfig)
}
