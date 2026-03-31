package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"order-service/internal/model"
	"order-service/internal/repository"
	"order-service/pkg/snowflake"
)

// Consumer Kafka消费者
type Consumer struct {
	reader     *kafka.Reader
	orderRepo  *repository.OrderRepository
	stopCh     chan struct{}
}

// NewConsumer 创建Kafka消费者
func NewConsumer(brokers []string, topic, groupID string, orderRepo *repository.OrderRepository) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		MaxWait:        500 * time.Millisecond,
		CommitInterval: time.Second,
	})

	return &Consumer{
		reader:    reader,
		orderRepo: orderRepo,
		stopCh:    make(chan struct{}),
	}
}

// Start 开始消费消息
func (c *Consumer) Start(ctx context.Context) {
	log.Println("Consumer: starting to consume messages")

	for {
		select {
		case <-c.stopCh:
			log.Println("Consumer: stopped")
			return
		case <-ctx.Done():
			log.Println("Consumer: context cancelled")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Consumer: read message error: %v", err)
				continue
			}

			c.processMessage(ctx, msg)
		}
	}
}

// processMessage 处理消息
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	var seckillMsg model.SeckillMessage
	if err := json.Unmarshal(msg.Value, &seckillMsg); err != nil {
		log.Printf("Consumer: unmarshal message failed: %v", err)
		return
	}

	log.Printf("Consumer: processing order_id=%d, user_id=%d, product_id=%d",
		seckillMsg.OrderID, seckillMsg.UserID, seckillMsg.ProductID)

	// 创建订单
	order := &model.Order{
		ID:        seckillMsg.OrderID,
		UserID:    seckillMsg.UserID,
		ProductID: seckillMsg.ProductID,
		Quantity:  seckillMsg.Quantity,
		Status:    model.OrderStatusCreated,
	}

	if err := c.orderRepo.Create(order); err != nil {
		log.Printf("Consumer: create order failed: %v", err)
		// 更新订单状态为失败
		c.orderRepo.UpdateStatus(seckillMsg.OrderID, model.OrderStatusFailed)
		return
	}

	log.Printf("Consumer: order created successfully, order_id=%d", seckillMsg.OrderID)
}

// Stop 停止消费
func (c *Consumer) Stop() {
	close(c.stopCh)
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	c.Stop()
	return c.reader.Close()
}

// GenerateOrderID 生成订单ID（供外部调用）
func GenerateOrderID() (int64, error) {
	return snowflake.GenerateID()
}
