package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"product-service/internal/config"
	"product-service/internal/model"
	"product-service/internal/repository"
)

const (
	// 缓存key前缀
	cacheKeyPrefix = "product:detail:"

	// 分布式锁key前缀
	lockKeyPrefix = "lock:product:"

	// 锁过期时间
	lockExpire = 10 * time.Second

	// 锁等待超时
	lockWaitTimeout = 3 * time.Second
)

type ProductService struct {
	repo   *repository.ProductRepository
	redis  *redis.Client
	config *config.Config
}

func NewProductService(repo *repository.ProductRepository, redisClient *redis.Client, cfg *config.Config) *ProductService {
	return &ProductService{
		repo:   repo,
		redis:  redisClient,
		config: cfg,
	}
}

// GetDBReadStats 获取从库状态信息
func (s *ProductService) GetDBReadStats() string {
	return "Read from MySQL Slave (Read Replica)"
}

// GetProductByID 获取商品详情 - 使用Cache Aside模式
// 处理了缓存穿透、缓存击穿、缓存雪崩问题
func (s *ProductService) GetProductByID(ctx context.Context, id uint) (*model.ProductDTO, error) {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "product-service"
	}

	// 1. 缓存穿透防护：参数合法性校验
	if id == 0 {
		return nil, fmt.Errorf("invalid product id: %d", id)
	}

	cacheKey := cacheKeyPrefix + fmt.Sprint(id)

	// 2. 先从缓存读取
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中
		var product model.ProductDTO
		if json.Unmarshal([]byte(cached), &product) == nil {
			fmt.Printf("[%s] Cache HIT for product %d\n", instanceID, id)
			return &product, nil
		}
	}

	fmt.Printf("[%s] Cache MISS for product %d\n", instanceID, id)

	// 3. 缓存未命中，尝试获取分布式锁防止缓存击穿
	lockKey := lockKeyPrefix + fmt.Sprint(id)

	// 尝试获取锁
	locked, err := s.redis.SetNX(ctx, lockKey, "1", lockExpire).Result()
	if err != nil {
		// 锁获取失败，短暂等待后重试读取缓存
		time.Sleep(100 * time.Millisecond)
		cached, retryErr := s.redis.Get(ctx, cacheKey).Result()
		if retryErr == nil {
			var product model.ProductDTO
			if json.Unmarshal([]byte(cached), &product) == nil {
				return &product, nil
			}
		}
		// 仍失败，继续查询数据库
	}

	var product *model.Product
	if locked {
		// 获取锁成功，从数据库查询
		product, err = s.repo.FindByID(id)
		if err != nil {
			// 数据库查询失败
			if err.Error() == "record not found" {
				// 缓存穿透防护：空值缓存，TTL较短
				s.setNullCache(ctx, cacheKey)
			}
			// 释放锁
			s.redis.Del(ctx, lockKey)
			return nil, err
		}

		// 4. 写入缓存 - 带随机TTL防止雪崩
		s.setProductCache(ctx, cacheKey, product)

		// 释放锁
		s.redis.Del(ctx, lockKey)
	} else {
		// 未获取到锁，说明有其他请求正在加载，等待后重试缓存
		// 等待锁超时后尝试读取
		waitCtx, cancel := context.WithTimeout(ctx, lockWaitTimeout)
		defer cancel()

		for waitCtx.Err() == nil {
			cached, err := s.redis.Get(ctx, cacheKey).Result()
			if err == nil {
				var productDTO model.ProductDTO
				if json.Unmarshal([]byte(cached), &productDTO) == nil {
					return &productDTO, nil
				}
			}
			time.Sleep(50 * time.Millisecond)
		}

		// 等待超时，直接查询数据库
		product, err = s.repo.FindByID(id)
		if err != nil {
			return nil, err
		}
	}

	dto := product.ToDTO()
	return &dto, nil
}

// setProductCache 设置商品缓存 - 带随机TTL防止雪崩
func (s *ProductService) setProductCache(ctx context.Context, key string, product *model.Product) {
	data, err := json.Marshal(product.ToDTO())
	if err != nil {
		fmt.Printf("Failed to marshal product: %v\n", err)
		return
	}

	// 计算TTL：基础TTL + 随机抖动（防止雪崩）
	ttl := s.config.Cache.TTL
	if s.config.Cache.EnableRandom {
		// 添加0-120秒的随机延迟
		ttl += rand.Intn(120)
	}

	// 最大TTL限制
	if ttl > s.config.Cache.MaxTTL {
		ttl = s.config.Cache.MaxTTL
	}

	s.redis.Set(ctx, key, string(data), time.Duration(ttl)*time.Second)
}

// setNullCache 设置空值缓存 - 防止缓存穿透
func (s *ProductService) setNullCache(ctx context.Context, key string) {
	// 使用较短的TTL缓存空值
	nullData, _ := json.Marshal(model.ProductDTO{ID: 0})
	s.redis.Set(ctx, key, string(nullData), time.Duration(s.config.Cache.NullTTL)*time.Second)
}

// UpdateProduct 更新商品 - 清除缓存
func (s *ProductService) UpdateProduct(product *model.Product) error {
	// 先更新数据库
	if err := s.repo.Update(product); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := cacheKeyPrefix + fmt.Sprint(product.ID)
	s.redis.Del(context.Background(), cacheKey)

	return nil
}

// DeleteProduct 删除商品 - 清除缓存
func (s *ProductService) DeleteProduct(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := cacheKeyPrefix + fmt.Sprint(id)
	s.redis.Del(context.Background(), cacheKey)

	return nil
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(product *model.Product) error {
	return s.repo.Create(product)
}

// ListProducts 获取商品列表
func (s *ProductService) ListProducts(limit, offset int) ([]model.ProductDTO, error) {
	products, err := s.repo.List(limit, offset)
	if err != nil {
		return nil, err
	}

	var dtos []model.ProductDTO
	for _, p := range products {
		dtos = append(dtos, p.ToDTO())
	}
	return dtos, nil
}

// PreloadHotProducts 预热热点数据 - 防止雪崩
func (s *ProductService) PreloadHotProducts(ids []uint) {
	ctx := context.Background()
	for _, id := range ids {
		product, err := s.repo.FindByID(id)
		if err != nil {
			continue
		}
		cacheKey := cacheKeyPrefix + fmt.Sprint(id)
		s.setProductCache(ctx, cacheKey, product)
	}
}
