package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"order-service/internal/kafka"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/pkg/snowflake"
)

const (
	// Redis key前缀
	stockKeyPrefix    = "seckill:stock:"     // 库存key
	idempotentKeyPrefix = "seckill:order:"   // 幂等性key
	lockKeyPrefix     = "lock:seckill:"      // 分布式锁key

	// 库存lua脚本 - 原子扣减
	stockLuaScript = `
		local stock = redis.call('GET', KEYS[1])
		if stock == false then
			return -1  -- 库存不存在
		end
		stock = tonumber(stock)
		if stock <= 0 then
			return 0  -- 库存不足
		end
		redis.call('DECR', KEYS[1])
		return 1  -- 扣减成功
	`

	// TTL
	idempotentTTL = 24 * time.Hour
	stockTTL      = 24 * time.Hour
	lockTTL       = 10 * time.Second
)

var (
	ErrStockNotEnough     = errors.New("stock not enough")
	ErrStockNotInitialized = errors.New("stock not initialized")
	ErrAlreadyOrdered     = errors.New("already ordered")
	ErrProductNotFound    = errors.New("product not found")
)

type SeckillService struct {
	redis     *redis.Client
	orderRepo *repository.OrderRepository
	producer  *kafka.Producer
}

// NewSeckillService 创建秒杀服务
func NewSeckillService(
	redisClient *redis.Client,
	orderRepo *repository.OrderRepository,
	producer *kafka.Producer,
) *SeckillService {
	return &SeckillService{
		redis:     redisClient,
		orderRepo: orderRepo,
		producer:  producer,
	}
}

// InitStock 初始化库存（从数据库加载或手动设置）
func (s *SeckillService) InitStock(ctx context.Context, productID int64, stock int) error {
	key := stockKeyPrefix + strconv.FormatInt(productID, 10)
	return s.redis.Set(ctx, key, stock, stockTTL).Err()
}

// GetStock 获取当前库存
func (s *SeckillService) GetStock(ctx context.Context, productID int64) (int, error) {
	key := stockKeyPrefix + strconv.FormatInt(productID, 10)
	val, err := s.redis.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, ErrProductNotFound
	}
	return val, err
}

// Seckill 秒杀下单
func (s *SeckillService) Seckill(ctx context.Context, req *model.SeckillRequest) (*model.SeckillResponse, error) {
	productID := req.ProductID
	userID := req.UserID
	quantity := req.Quantity

	// 1. 幂等性检查 - 同一用户同一商品只能秒杀一次
	idempotentKey := idempotentKeyPrefix + strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(productID, 10)

	// 使用SET NX检查是否已存在订单
	exists, err := s.redis.Exists(ctx, idempotentKey).Result()
	if err != nil {
		return nil, fmt.Errorf("check idempotent failed: %w", err)
	}
	if exists > 0 {
		// 已存在，获取已有订单信息
		orderIDStr, _ := s.redis.Get(ctx, idempotentKey).Result()
		orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)
		return &model.SeckillResponse{
			OrderID: orderID,
			Status:  "already_ordered",
			Message: "您已秒杀过该商品",
		}, ErrAlreadyOrdered
	}

	// 2. 库存扣减（原子操作）
	stockKey := stockKeyPrefix + strconv.FormatInt(productID, 10)

	// 使用Lua脚本保证原子性
	result, err := s.redis.Eval(ctx, stockLuaScript, []string{stockKey}).Int()
	if err != nil {
		return nil, fmt.Errorf("stock deduction failed: %w", err)
	}

	switch result {
	case -1:
		return nil, ErrProductNotFound
	case 0:
		return nil, ErrStockNotEnough
	}

	// 3. 生成订单ID（雪花算法）
	orderID, err := snowflake.GenerateID()
	if err != nil {
		// 回滚库存
		s.redis.Incr(ctx, stockKey)
		return nil, fmt.Errorf("generate order id failed: %w", err)
	}

	// 4. 设置幂等性key
	s.redis.Set(ctx, idempotentKey, strconv.FormatInt(orderID, 10), idempotentTTL)

	// 5. 发送Kafka消息异步创建订单
	seckillMsg := model.SeckillMessage{
		OrderID:   orderID,
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Timestamp: time.Now().UnixMilli(),
	}

	msgKey := strconv.FormatInt(productID, 10) + "-" + strconv.FormatInt(userID, 10)
	if err := s.producer.SendMessage(ctx, msgKey, seckillMsg); err != nil {
		// 回滚幂等性key和库存
		s.redis.Del(ctx, idempotentKey)
		s.redis.Incr(ctx, stockKey)
		return nil, fmt.Errorf("send kafka message failed: %w", err)
	}

	return &model.SeckillResponse{
		OrderID: orderID,
		Status:  "queued",
		Message: "秒杀成功，订单正在处理中",
	}, nil
}

// GetOrderByID 根据订单ID查询订单
func (s *SeckillService) GetOrderByID(ctx context.Context, orderID int64) (*model.OrderDTO, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	dto := order.ToDTO()
	return &dto, nil
}

// GetOrdersByUserID 根据用户ID查询订单列表
func (s *SeckillService) GetOrdersByUserID(ctx context.Context, userID int64) ([]model.OrderDTO, error) {
	orders, err := s.orderRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	var dtos []model.OrderDTO
	for _, o := range orders {
		dtos = append(dtos, o.ToDTO())
	}
	return dtos, nil
}

// CheckOrderExists 检查订单是否存在（用于幂等性检查）
func (s *SeckillService) CheckOrderExists(ctx context.Context, userID, productID int64) (*model.OrderDTO, error) {
	idempotentKey := idempotentKeyPrefix + strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(productID, 10)

	orderIDStr, err := s.redis.Get(ctx, idempotentKey).Result()
	if err == redis.Nil {
		// Redis中没有，查询数据库
		order, err := s.orderRepo.FindByUserIDAndProductID(userID, productID)
		if err != nil {
			return nil, nil // 没有订单
		}
		dto := order.ToDTO()
		return &dto, nil
	}
	if err != nil {
		return nil, err
	}

	orderID, _ := strconv.ParseInt(orderIDStr, 10, 64)
	return s.GetOrderByID(ctx, orderID)
}

// SetIdempotentKey 设置幂等性key（用于测试或手动同步）
func (s *SeckillService) SetIdempotentKey(ctx context.Context, userID, productID, orderID int64) error {
	key := idempotentKeyPrefix + strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(productID, 10)
	return s.redis.Set(ctx, key, strconv.FormatInt(orderID, 10), idempotentTTL).Err()
}

// GetIdempotentOrder 获取幂等性key对应的订单ID
func (s *SeckillService) GetIdempotentOrder(ctx context.Context, userID, productID int64) (int64, error) {
	key := idempotentKeyPrefix + strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(productID, 10)
	orderIDStr, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(orderIDStr, 10, 64)
}

// SetStock 设置库存（用于测试）
func (s *SeckillService) SetStock(ctx context.Context, productID int64, stock int) error {
	key := stockKeyPrefix + strconv.FormatInt(productID, 10)
	return s.redis.Set(ctx, key, stock, stockTTL).Err()
}

// SeckillWithStockCheck 秒杀时检查库存并扣减（返回订单是否创建成功）
func (s *SeckillService) SeckillWithStockCheck(ctx context.Context, productID, userID int64, quantity int) (*model.Order, error) {
	// 1. 幂等性检查
	existingOrder, _ := s.CheckOrderExists(ctx, userID, productID)
	if existingOrder != nil {
		return nil, ErrAlreadyOrdered
	}

	// 2. 扣减库存
	stockKey := stockKeyPrefix + strconv.FormatInt(productID, 10)
	result, err := s.redis.Eval(ctx, stockLuaScript, []string{stockKey}).Int()
	if err != nil {
		return nil, err
	}
	if result <= 0 {
		return nil, ErrStockNotEnough
	}

	// 3. 生成订单ID
	orderID, err := snowflake.GenerateID()
	if err != nil {
		s.redis.Incr(ctx, stockKey)
		return nil, err
	}

	// 4. 创建订单
	order := &model.Order{
		ID:        orderID,
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Status:    model.OrderStatusCreated,
	}

	if err := s.orderRepo.Create(order); err != nil {
		s.redis.Incr(ctx, stockKey)
		return nil, err
	}

	// 5. 设置幂等性key
	s.SetIdempotentKey(ctx, userID, productID, orderID)

	return order, nil
}

// StockInfo 库存信息（用于测试和监控）
type StockInfo struct {
	ProductID int64 `json:"product_id"`
	Stock     int   `json:"stock"`
}

// GetStockInfo 获取库存信息
func (s *SeckillService) GetStockInfo(ctx context.Context, productID int64) (*StockInfo, error) {
	stock, err := s.GetStock(ctx, productID)
	if err != nil && err != ErrProductNotFound {
		return nil, err
	}
	return &StockInfo{
		ProductID: productID,
		Stock:     stock,
	}, nil
}

// OrderInfo 订单信息
type OrderInfo struct {
	OrderID  int64  `json:"order_id"`
	UserID   int64  `json:"user_id"`
	ProductID int64 `json:"product_id"`
	Status   string `json:"status"`
}

// MarshalJSON 序列化订单状态
func (s *SeckillService) MarshalOrderStatus(status model.OrderStatus) string {
	switch status {
	case model.OrderStatusPending:
		return "pending"
	case model.OrderStatusCreated:
		return "created"
	case model.OrderStatusCanceled:
		return "canceled"
	case model.OrderStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ParseSeckillResponse 解析秒杀响应
func ParseSeckillResponse(resp *model.SeckillResponse) string {
	data, _ := json.Marshal(resp)
	return string(data)
}
