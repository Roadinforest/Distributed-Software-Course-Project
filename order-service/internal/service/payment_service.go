package service

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"

	"order-service/internal/kafka"
	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/pkg/snowflake"
)

var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrOrderAlreadyPaid  = errors.New("order already paid")
	ErrOrderNotPaid      = errors.New("order not paid")
	ErrInsufficientStock = errors.New("insufficient stock")
)

const (
	paymentKeyPrefix = "payment:order:"
	paymentTTL       = 24 * time.Hour
)

type PaymentService struct {
	redis      *redis.Client
	orderRepo  *repository.OrderRepository
	producer   *kafka.Producer
}

func NewPaymentService(
	redisClient *redis.Client,
	orderRepo *repository.OrderRepository,
	producer *kafka.Producer,
) *PaymentService {
	return &PaymentService{
		redis:     redisClient,
		orderRepo: orderRepo,
		producer:  producer,
	}
}

// Pay 支付订单
func (s *PaymentService) Pay(ctx context.Context, orderID int64) (*model.PayResponse, error) {
	// 1. 检查订单是否存在
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}

	// 2. 检查订单状态
	if order.Status == model.OrderStatusCreated || order.Status == model.OrderStatusPending {
		// 订单已支付或正在处理，检查Redis缓存
		key := paymentKeyPrefix + itoa(orderID)
		exists, _ := s.redis.Exists(ctx, key).Result()
		if exists > 0 {
			return nil, ErrOrderAlreadyPaid
		}
	}
	if order.Status == model.OrderStatusCanceled {
		return nil, errors.New("order canceled")
	}

	// 3. 生成支付ID
	paymentID, err := snowflake.GenerateID()
	if err != nil {
		return nil, err
	}

	// 4. 更新订单状态为已支付（乐观锁）
	if err := s.orderRepo.UpdateStatus(orderID, model.OrderStatusCreated); err != nil {
		return nil, err
	}

	// 5. 发送支付成功消息到Kafka
	paymentMsg := model.PaymentMessage{
		OrderID:   orderID,
		PaymentID: paymentID,
		Action:    "pay",
		Timestamp: time.Now().UnixMilli(),
	}

	msgKey := itoa(orderID)
	if err := s.producer.SendPaymentMessage(ctx, msgKey, paymentMsg); err != nil {
		// 支付消息发送失败，记录日志但不影响支付结果
		// 实际生产中需要补偿机制
	}

	// 6. 发送订单状态更新消息
	orderMsg := model.OrderStatusMessage{
		OrderID:   orderID,
		Status:    "paid",
		Timestamp: time.Now().UnixMilli(),
	}
	s.producer.SendOrderStatusMessage(ctx, itoa(orderID), orderMsg)

	// 7. 设置支付完成缓存（用于幂等性）
	key := paymentKeyPrefix + itoa(orderID)
	s.redis.Set(ctx, key, paymentID, paymentTTL)

	return &model.PayResponse{
		PaymentID: paymentID,
		Status:    "paid",
		Message:   "payment successful",
	}, nil
}

// GetPaymentByOrderID 根据订单ID获取支付信息
func (s *PaymentService) GetPaymentByOrderID(ctx context.Context, orderID int64) (*model.Payment, error) {
	// 先查缓存
	key := paymentKeyPrefix + itoa(orderID)
	paymentIDStr, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		paymentID, _ := parseInt64(paymentIDStr)
		return &model.Payment{
			ID:      paymentID,
			OrderID: orderID,
			Status:  model.PaymentStatusPaid,
		}, nil
	}

	// 查数据库
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, nil
	}

	if order.Status != model.OrderStatusCreated {
		return nil, nil
	}

	return &model.Payment{
		ID:      0,
		OrderID: orderID,
		Status:  model.PaymentStatusPaid,
	}, nil
}

// Refund 退款
func (s *PaymentService) Refund(ctx context.Context, orderID int64) error {
	// 1. 检查订单是否存在
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return ErrOrderNotFound
	}

	// 2. 检查订单是否已支付
	if order.Status != model.OrderStatusCreated {
		return ErrOrderNotPaid
	}

	// 3. 发送退款消息
	refundMsg := model.PaymentMessage{
		OrderID:   orderID,
		PaymentID: 0,
		Action:    "refund",
		Timestamp: time.Now().UnixMilli(),
	}

	msgKey := itoa(orderID)
	if err := s.producer.SendPaymentMessage(ctx, msgKey, refundMsg); err != nil {
		return err
	}

	// 4. 发送订单状态更新消息
	orderMsg := model.OrderStatusMessage{
		OrderID:   orderID,
		Status:    "refunded",
		Timestamp: time.Now().UnixMilli(),
	}
	s.producer.SendOrderStatusMessage(ctx, itoa(orderID), orderMsg)

	return nil
}

// itoa 简单int64转string
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	if negative {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}

// parseInt64 简单string转int64
func parseInt64(s string) (int64, error) {
	var result int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid number")
		}
		result = result*10 + int64(c-'0')
	}
	return result, nil
}
