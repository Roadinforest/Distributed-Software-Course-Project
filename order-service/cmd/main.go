package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"order-service/internal/bootstrap"
	"order-service/internal/handler"
	"order-service/internal/kafka"
	"order-service/internal/repository"
	"order-service/internal/router"
	"order-service/internal/service"
)

func main() {
	app, err := bootstrap.Initialize("./config")
	if err != nil {
		log.Fatalf("initialize app failed: %v", err)
	}

	// 创建仓储
	orderRepo := repository.NewOrderRepository(app.DB)

	// 创建Kafka生产者
	producer := kafka.NewProducer(app.Config.Kafka.Brokers)
	defer producer.Close()

	// 创建秒杀服务
	seckillService := service.NewSeckillService(app.Redis, orderRepo, producer)

	// 创建支付服务
	paymentService := service.NewPaymentService(app.Redis, orderRepo, producer)

	// 创建处理器
	seckillHandler := handler.NewSeckillHandler(seckillService)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// 创建路由
	r := router.New(seckillHandler, paymentHandler)

	// 启动Kafka消费者
	consumer := kafka.NewConsumer(
		app.Config.Kafka.Brokers,
		app.Config.Kafka.Topic,
		app.Config.Kafka.GroupID,
		orderRepo,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumer.Start(ctx)
	defer consumer.Close()

	// 启动HTTP服务
	addr := fmt.Sprintf(":%d", app.Config.Server.Port)
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "order-service"
	}
	log.Printf("starting server instance=%s addr=%s", instanceID, addr)
	log.Printf("kafka brokers=%s topic=%s", app.Config.Kafka.BrokerList(), app.Config.Kafka.Topic)

	// 优雅关闭
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatalf("start server failed: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	cancel()
	log.Println("server stopped")
}
