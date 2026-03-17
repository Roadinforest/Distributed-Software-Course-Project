package bootstrap

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"product-service/internal/config"
	"product-service/internal/model"
)

const (
	maxDBConnectRetries = 20
	dbRetryInterval     = 2 * time.Second
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
}

func Initialize(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	db, err := connectMySQLWithRetry(cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("connect mysql failed: %w", err)
	}

	if err := db.AutoMigrate(&model.Product{}); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("warning: redis ping failed: %v", err)
	}

	return &App{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
	}, nil
}

func connectMySQLWithRetry(dsn string) (*gorm.DB, error) {
	var lastErr error

	for i := 1; i <= maxDBConnectRetries; i++ {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			lastErr = err
			log.Printf("mysql connection attempt %d/%d failed: %v", i, maxDBConnectRetries, err)
			time.Sleep(dbRetryInterval)
			continue
		}

		sqlDB, err := db.DB()
		if err != nil {
			lastErr = err
			log.Printf("mysql db handle attempt %d/%d failed: %v", i, maxDBConnectRetries, err)
			time.Sleep(dbRetryInterval)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		pingErr := sqlDB.PingContext(ctx)
		cancel()
		if pingErr != nil {
			lastErr = pingErr
			log.Printf("mysql ping attempt %d/%d failed: %v", i, maxDBConnectRetries, pingErr)
			time.Sleep(dbRetryInterval)
			continue
		}

		return db, nil
	}

	return nil, lastErr
}
