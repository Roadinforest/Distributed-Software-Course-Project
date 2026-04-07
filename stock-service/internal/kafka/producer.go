package kafka

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"

	"stock-service/internal/model"
)

const (
	StockTopic = "stock-events"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        StockTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	return &Producer{writer: writer}
}

// SendMessage 发送库存消息
func (p *Producer) SendMessage(ctx context.Context, key string, msg model.StockMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: data,
	})
}

// SendReserveMessage 发送库存预扣减消息
func (p *Producer) SendReserveMessage(ctx context.Context, orderID, productID int64, quantity int) error {
	msg := model.StockMessage{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Action:    "reserve",
		Timestamp: time.Now().UnixMilli(),
	}
	key := strconv.FormatInt(productID, 10) + "-" + strconv.FormatInt(orderID, 10)
	return p.SendMessage(ctx, key, msg)
}

// SendConfirmMessage 发送库存确认扣减消息
func (p *Producer) SendConfirmMessage(ctx context.Context, orderID, productID int64) error {
	msg := model.StockMessage{
		OrderID:   orderID,
		ProductID: productID,
		Action:    "confirm",
		Timestamp: time.Now().UnixMilli(),
	}
	key := strconv.FormatInt(productID, 10) + "-" + strconv.FormatInt(orderID, 10)
	return p.SendMessage(ctx, key, msg)
}

// SendCancelMessage 发送库存取消预扣减消息
func (p *Producer) SendCancelMessage(ctx context.Context, orderID, productID int64) error {
	msg := model.StockMessage{
		OrderID:   orderID,
		ProductID: productID,
		Action:    "cancel",
		Timestamp: time.Now().UnixMilli(),
	}
	key := strconv.FormatInt(productID, 10) + "-" + strconv.FormatInt(orderID, 10)
	return p.SendMessage(ctx, key, msg)
}

// Close 关闭producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
