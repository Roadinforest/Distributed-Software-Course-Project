package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"stock-service/internal/model"
	"stock-service/internal/repository"
)

const (
	// Redis key前缀
	stockKeyPrefix    = "stock:quantity:"     // 库存数量key
	reservedKeyPrefix = "stock:reserved:"     // 预扣减key
	idempotentKeyPrefix = "stock:order:"      // 幂等性key

	// Lua脚本 - 库存预扣减（原子操作）
	reserveStockLua = `
		local stock_key = KEYS[1]
		local reserved_key = KEYS[2]
		local quantity = tonumber(ARGV[1])
		local order_id = ARGV[2]
		local ttl = tonumber(ARGV[3])

		-- 检查幂等性
		local exists = redis.call('EXISTS', reserved_key)
		if exists == 1 then
			return -2  -- 已存在
		end

		-- 检查库存
		local stock = redis.call('GET', stock_key)
		if stock == false then
			return -1  -- 库存不存在
		end
		stock = tonumber(stock)
		if stock < quantity then
			return 0  -- 库存不足
		end

		-- 扣减库存
		redis.call('DECRBY', stock_key, quantity)
		-- 设置预扣减记录
		redis.call('HSET', reserved_key, order_id, quantity)
		redis.call('EXPIRE', reserved_key, ttl)

		return 1  -- 成功
	`

	// Lua脚本 - 确认扣减
	confirmDeductLua = `
		local stock_key = KEYS[1]
		local reserved_key = KEYS[2]
		local order_id = ARGV[1]

		-- 检查预扣减记录
		local quantity = redis.call('HGET', reserved_key, order_id)
		if quantity == false then
			return -1  -- 没有预扣减记录
		end

		-- 删除预扣减记录
		redis.call('HDEL', reserved_key, order_id)

		return tonumber(quantity)  -- 返回扣减数量
	`

	// Lua脚本 - 取消预扣减（回滚）
	cancelReserveLua = `
		local stock_key = KEYS[1]
		local reserved_key = KEYS[2]
		local order_id = ARGV[1]

		-- 检查预扣减记录
		local quantity = redis.call('HGET', reserved_key, order_id)
		if quantity == false then
			return -1  -- 没有预扣减记录
		end
		quantity = tonumber(quantity)

		-- 回滚库存
		redis.call('INCRBY', stock_key, quantity)
		-- 删除预扣减记录
		redis.call('HDEL', reserved_key, order_id)

		return quantity  -- 返回回滚数量
	`

	stockTTL      = 24 * time.Hour
	reservedTTL   = 30 * time.Minute
	idempotentTTL = 24 * time.Hour
)

var (
	ErrStockNotEnough   = errors.New("stock not enough")
	ErrStockNotExist    = errors.New("stock not exist")
	ErrAlreadyReserved  = errors.New("already reserved")
	ErrReserveNotFound  = errors.New("reserve not found")
)

type StockService struct {
	redis      *redis.Client
	stockRepo  *repository.StockRepository
}

func NewStockService(redisClient *redis.Client, stockRepo *repository.StockRepository) *StockService {
	return &StockService{
		redis:     redisClient,
		stockRepo: stockRepo,
	}
}

// InitStock 初始化库存
func (s *StockService) InitStock(ctx context.Context, productID int64, stock int) error {
	key := stockKeyPrefix + strconv.FormatInt(productID, 10)
	return s.redis.Set(ctx, key, stock, stockTTL).Err()
}

// GetStock 获取当前库存
func (s *StockService) GetStock(ctx context.Context, productID int64) (int, error) {
	key := stockKeyPrefix + strconv.FormatInt(productID, 10)
	val, err := s.redis.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, ErrStockNotExist
	}
	return val, err
}

// ReserveStock 预扣减库存（基于Redis原子操作）
func (s *StockService) ReserveStock(ctx context.Context, orderID, productID int64, quantity int) error {
	stockKey := stockKeyPrefix + strconv.FormatInt(productID, 10)
	reservedKey := reservedKeyPrefix + strconv.FormatInt(productID, 10)

	// 使用Lua脚本保证原子性
	result, err := s.redis.Eval(ctx, reserveStockLua,
		[]string{stockKey, reservedKey},
		quantity, strconv.FormatInt(orderID, 10), int(reservedTTL.Seconds())).Int()

	if err != nil {
		return fmt.Errorf("reserve stock failed: %w", err)
	}

	switch result {
	case -2:
		return ErrAlreadyReserved
	case -1:
		return ErrStockNotExist
	case 0:
		return ErrStockNotEnough
	}

	return nil
}

// ConfirmDeduct 确认扣减（删除预扣减记录）
func (s *StockService) ConfirmDeduct(ctx context.Context, orderID, productID int64) error {
	stockKey := stockKeyPrefix + strconv.FormatInt(productID, 10)
	reservedKey := reservedKeyPrefix + strconv.FormatInt(productID, 10)
	idempotentKey := idempotentKeyPrefix + strconv.FormatInt(orderID, 10)

	// 使用Lua脚本确认扣减
	quantity, err := s.redis.Eval(ctx, confirmDeductLua,
		[]string{stockKey, reservedKey},
		strconv.FormatInt(orderID, 10)).Int()

	if err != nil {
		return fmt.Errorf("confirm deduct failed: %w", err)
	}

	if quantity < 0 {
		return ErrReserveNotFound
	}

	// 删除幂等性key
	s.redis.Del(ctx, idempotentKey)

	return nil
}

// CancelReserve 取消预扣减（回滚库存）
func (s *StockService) CancelReserve(ctx context.Context, orderID, productID int64) error {
	stockKey := stockKeyPrefix + strconv.FormatInt(productID, 10)
	reservedKey := reservedKeyPrefix + strconv.FormatInt(productID, 10)
	idempotentKey := idempotentKeyPrefix + strconv.FormatInt(orderID, 10)

	// 使用Lua脚本取消预扣减
	quantity, err := s.redis.Eval(ctx, cancelReserveLua,
		[]string{stockKey, reservedKey},
		strconv.FormatInt(orderID, 10)).Int()

	if err != nil {
		return fmt.Errorf("cancel reserve failed: %w", err)
	}

	if quantity < 0 {
		return ErrReserveNotFound
	}

	// 删除幂等性key
	s.redis.Del(ctx, idempotentKey)

	return nil
}

// GetReservedQuantity 获取预扣减数量
func (s *StockService) GetReservedQuantity(ctx context.Context, productID int64, orderID int64) (int, error) {
	reservedKey := reservedKeyPrefix + strconv.FormatInt(productID, 10)
	quantity, err := s.redis.HGet(ctx, reservedKey, strconv.FormatInt(orderID, 10)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return quantity, err
}

// StockInfo 库存信息
type StockInfo struct {
	ProductID       int64 `json:"product_id"`
	TotalStock      int   `json:"total_stock"`
	ReservedStock   int   `json:"reserved_stock"`
	AvailableStock  int   `json:"available_stock"`
}

// GetStockInfo 获取库存详细信息
func (s *StockService) GetStockInfo(ctx context.Context, productID int64) (*StockInfo, error) {
	stock, err := s.GetStock(ctx, productID)
	if err != nil && err != ErrStockNotExist {
		return nil, err
	}

	reservedKey := reservedKeyPrefix + strconv.FormatInt(productID, 10)
	reserved, _ := s.redis.HLen(ctx, reservedKey).Result()

	info := &StockInfo{
		ProductID:       productID,
		TotalStock:      stock,
		ReservedStock:   int(reserved),
		AvailableStock:  stock - int(reserved),
	}

	return info, nil
}

// SyncStockToRedis 从数据库同步库存到Redis
func (s *StockService) SyncStockToRedis(ctx context.Context, productID int64) error {
	stock, err := s.stockRepo.FindByProductID(productID)
	if err != nil {
		return err
	}

	key := stockKeyPrefix + strconv.FormatInt(productID, 10)
	available := stock.Quantity - stock.Reserved

	return s.redis.Set(ctx, key, available, stockTTL).Err()
}

// ProcessStockMessage 处理库存消息（Kafka消费者调用）
func (s *StockService) ProcessStockMessage(ctx context.Context, msg *model.StockMessage) error {
	switch msg.Action {
	case "reserve":
		return s.ReserveStock(ctx, msg.OrderID, msg.ProductID, msg.Quantity)
	case "confirm":
		return s.ConfirmDeduct(ctx, msg.OrderID, msg.ProductID)
	case "cancel":
		return s.CancelReserve(ctx, msg.OrderID, msg.ProductID)
	default:
		return fmt.Errorf("unknown action: %s", msg.Action)
	}
}
