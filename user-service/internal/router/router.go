package router

import (
	"log"
	"net/http"
	"os"
	"time"

	"user-service/internal/consul"
	"user-service/internal/handler"
	"user-service/internal/middleware"
	jwtpkg "user-service/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func New(userHandler *handler.UserHandler, jwtManager *jwtpkg.Manager) *gin.Engine {
	r := gin.Default()

	// 健康检查端点
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Consul配置动态监听端点
	r.GET("/api/v1/config", func(c *gin.Context) {
		key := c.Query("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
			return
		}

		// 从环境变量或context获取consul client
		consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
		if consulAddr == "" {
			consulAddr = "localhost:8500"
		}

		client, err := consul.NewConsulClient(consulAddr)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		defer client.Close()

		value, err := client.GetConfig("config/" + key)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"key": key, "value": string(value)})
	})

	// API路由组
	api := r.Group("/api/v1")
	// 应用限流中间件
	api.Use(middleware.GlobalRateLimitMiddleware)
	// 应用熔断中间件
	api.Use(middleware.CircuitBreakerMiddleware)

	users := api.Group("/users")
	{
		users.POST("/register", userHandler.Register)
		users.POST("/login", userHandler.Login)

		authed := users.Group("")
		authed.Use(middleware.JWTAuth(jwtManager))
		{
			authed.GET("/profile", userHandler.Profile)
			authed.PUT("/profile", userHandler.UpdateProfile)
		}
	}

	// Consul服务注册
	go registerServiceWithConsul()

	return r
}

// registerServiceWithConsul 将服务注册到Consul
func registerServiceWithConsul() {
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr == "" {
		consulAddr = "localhost:8500"
	}

	client, err := consul.NewConsulClient(consulAddr)
	if err != nil {
		log.Printf("consul client created failed: %v, will retry later", err)
		// 重试注册
		time.Sleep(5 * time.Second)
		go registerServiceWithConsul()
		return
	}
	defer client.Close()

	port := 8081
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "user-service-1"
	}

	// 从环境变量获取端口
	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		var p int
		if _, err := os.Stdout.WriteString(envPort); err == nil {
			p = 8081
		}
		if p > 0 {
			port = p
		}
	}

	// 注册服务
	err = client.RegisterService(
		"user-service",
		port,
		[]string{"go", "gin", "v1"},
		"/healthz",
	)
	if err != nil {
		log.Printf("register service to consul failed: %v, will retry later", err)
		time.Sleep(5 * time.Second)
		go registerServiceWithConsul()
		return
	}

	log.Printf("service registered to consul: %s", instanceID)

	// 模拟配置动态更新
	go watchDynamicConfig(client)
}

// watchDynamicConfig 监听配置变化
func watchDynamicConfig(client *consul.ConsulClient) {
	err := client.WatchConfig("config/service/rate-limit", func(value []byte) {
		if len(value) > 0 {
			log.Printf("dynamic config updated: rate-limit = %s", string(value))
			// 这里可以动态更新限流配置
		}
	})
	if err != nil {
		log.Printf("watch config failed: %v", err)
	}
}
