package config

import (
	"os"
	"strconv"
)

type Config struct {
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
	}
	Redis struct {
		Host string
		Port int
	}
	Kafka struct {
		Brokers []string
	}
	App struct {
		Port int
	}
}

func Load() *Config {
	return &Config{
		Database: struct {
			Host     string
			Port     int
			User     string
			Password string
			DBName   string
		}{
			Host:     getEnv("STOCK_SVC_DATABASE_HOST", "mysql-stock"),
			Port:     getEnvInt("STOCK_SVC_DATABASE_PORT", 3306),
			User:     getEnv("STOCK_SVC_DATABASE_USER", "root"),
			Password: getEnv("STOCK_SVC_DATABASE_PASSWORD", "root123"),
			DBName:   getEnv("STOCK_SVC_DATABASE_DBNAME", "stock_db"),
		},
		Redis: struct {
			Host string
			Port int
		}{
			Host: getEnv("STOCK_SVC_REDIS_HOST", "redis"),
			Port: getEnvInt("STOCK_SVC_REDIS_PORT", 6379),
		},
		Kafka: struct {
			Brokers []string
		}{
			Brokers: []string{getEnv("STOCK_SVC_KAFKA_BROKERS", "kafka:9092")},
		},
		App: struct {
			Port int
		}{
			Port: getEnvInt("STOCK_SVC_APP_PORT", 8087),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
