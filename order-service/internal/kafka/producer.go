package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer Kafka生产者
type Producer struct {
	writer *kafka.Writer
	topic  string
}

// NewProducer 创建Kafka生产者
func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	return &Producer{
		writer: writer,
		topic:  topic,
	}
}

// SendMessage 发送消息
func (p *Producer) SendMessage(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: data,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write message failed: %w", err)
	}

	log.Printf("Producer: sent message with key=%s to topic=%s", key, p.topic)
	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	return p.writer.Close()
}
