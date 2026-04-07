package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"

	"stock-service/internal/model"
	"stock-service/internal/service"
)

type Consumer struct {
	reader       *kafka.Reader
	stockService *service.StockService
	stopCh       chan struct{}
}

func NewConsumer(brokers []string, topic, groupID string, stockService *service.StockService) *Consumer {
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
		reader:       reader,
		stockService: stockService,
		stopCh:       make(chan struct{}),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Println("Stock Consumer: starting to consume messages")

	for {
		select {
		case <-c.stopCh:
			log.Println("Stock Consumer: stopped")
			return
		case <-ctx.Done():
			log.Println("Stock Consumer: context cancelled")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				log.Printf("Stock Consumer: read message error: %v", err)
				continue
			}

			c.processMessage(ctx, msg)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	var stockMsg model.StockMessage
	if err := json.Unmarshal(msg.Value, &stockMsg); err != nil {
		log.Printf("Stock Consumer: unmarshal message failed: %v", err)
		return
	}

	log.Printf("Stock Consumer: processing action=%s, order_id=%d, product_id=%d",
		stockMsg.Action, stockMsg.OrderID, stockMsg.ProductID)

	if err := c.stockService.ProcessStockMessage(ctx, &stockMsg); err != nil {
		log.Printf("Stock Consumer: process message failed: %v", err)
		return
	}

	log.Printf("Stock Consumer: processed successfully, order_id=%d", stockMsg.OrderID)
}

func (c *Consumer) Stop() {
	close(c.stopCh)
}

func (c *Consumer) Close() error {
	c.Stop()
	return c.reader.Close()
}
