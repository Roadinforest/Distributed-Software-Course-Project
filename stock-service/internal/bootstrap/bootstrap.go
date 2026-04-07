package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"stock-service/internal/config"
	"stock-service/internal/kafka"
	"stock-service/internal/repository"
	"stock-service/internal/service"
)

type App struct {
	Engine       *gin.Engine
	StockService *service.StockService
	Producer     *kafka.Producer
	Consumer     *kafka.Consumer
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func NewApp(cfg *config.Config) (*App, error) {
	// 初始化MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("connect to mysql failed: %w", err)
	}

	// 初始化Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		DB:   0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis failed: %w", err)
	}

	// 初始化Repository
	stockRepo := repository.NewStockRepository(db)

	// 初始化数据库
	if err := stockRepo.InitDatabase(); err != nil {
		return nil, fmt.Errorf("init database failed: %w", err)
	}

	// 初始化Service
	stockService := service.NewStockService(redisClient, stockRepo)

	// 初始化Kafka Producer
	producer := kafka.NewProducer(cfg.Kafka.Brokers)

	// 初始化Kafka Consumer
	consumer := kafka.NewConsumer(cfg.Kafka.Brokers, kafka.StockTopic, "stock-service-group", stockService)

	return &App{
		StockService: stockService,
		Producer:     producer,
		Consumer:     consumer,
	}, nil
}
