package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Cache    CacheConfig    `mapstructure:"cache"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Charset  string `mapstructure:"charset"`

	// 读写分离配置 - 从库
	ReadHost     string `mapstructure:"read_host"`
	ReadPort     int    `mapstructure:"read_port"`
	ReadUser     string `mapstructure:"read_user"`
	ReadPassword string `mapstructure:"read_password"`
	ReadDBName   string `mapstructure:"read_dbname"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type CacheConfig struct {
	TTL           int  `mapstructure:"ttl"`
	MaxTTL        int  `mapstructure:"max_ttl"`
	NullTTL       int  `mapstructure:"null_ttl"`
	EnableRandom  bool `mapstructure:"enable_random"`
}

func (d DatabaseConfig) DSN() string {
	charset := d.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", d.User, d.Password, d.Host, d.Port, d.DBName, charset)
}

// ReadDSN 返回从库(读库)的DSN
func (d DatabaseConfig) ReadDSN() string {
	// 如果没有配置从库，使用主库
	if d.ReadHost == "" {
		return d.DSN()
	}
	charset := d.Charset
	if charset == "" {
		charset = "utf8mb4"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", d.ReadUser, d.ReadPassword, d.ReadHost, d.ReadPort, d.ReadDBName, charset)
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.SetEnvPrefix("PRODUCT_SVC")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	return &cfg, nil
}
