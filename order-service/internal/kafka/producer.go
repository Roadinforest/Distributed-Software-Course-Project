package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"order-service/internal/model"
)

const (
	OrderTopic    = "order-events"
	PaymentTopic  = "payment-events"
	StockTopic    = "stock-events"
)

// Producer Kafka生产者
type Producer struct {
	orderWriter   *kafka.Writer
	paymentWriter *kafka.Writer
	stockWriter   *kafka.Writer
}

// NewProducer 创建Kafka生产者
func NewProducer(brokers []string) *Producer {
	orderWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        OrderTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	paymentWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        PaymentTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	stockWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        StockTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	return &Producer{
		orderWriter:   orderWriter,
		paymentWriter: paymentWriter,
		stockWriter:   stockWriter,
	}
}

// SendMessage 发送消息到指定topic
func (p *Producer) sendMessage(ctx context.Context, writer *kafka.Writer, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: data,
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}

	log.Printf("Producer: sent message with key=%s to topic=%s", key, writer.Topic)
	return nil
}

// SendSeckillMessage 发送秒杀订单消息
func (p *Producer) SendSeckillMessage(ctx context.Context, key string, msg model.SeckillMessage) error {
	return p.sendMessage(ctx, p.orderWriter, key, msg)
}

// SendPaymentMessage 发送支付消息
func (p *Producer) SendPaymentMessage(ctx context.Context, key string, msg model.PaymentMessage) error {
	return p.sendMessage(ctx, p.paymentWriter, key, msg)
}

// SendOrderStatusMessage 发送订单状态更新消息
func (p *Producer) SendOrderStatusMessage(ctx context.Context, key string, msg model.OrderStatusMessage) error {
	return p.sendMessage(ctx, p.orderWriter, key, msg)
}

// SendStockMessage 发送库存消息
func (p *Producer) SendStockMessage(ctx context.Context, key string, msg model.StockMessage) error {
	return p.sendMessage(ctx, p.stockWriter, key, msg)
}

// Close 关闭生产者
func (p *Producer) Close() error {
	if err := p.orderWriter.Close(); err != nil {
		return err
	}
	if err := p.paymentWriter.Close(); err != nil {
		return err
	}
	return p.stockWriter.Close()
}
