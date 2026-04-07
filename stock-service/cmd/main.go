package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"stock-service/internal/bootstrap"
	"stock-service/internal/config"
	"stock-service/internal/handler"
	"stock-service/internal/router"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化应用
	app, err := bootstrap.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// 初始化处理器
	stockHandler := handler.NewStockHandler(app.StockService)

	// 设置路由
	app.Engine = router.New(stockHandler)

	// 启动Kafka Consumer
	ctx, cancel := context.WithCancel(context.Background())
	go app.Consumer.Start(ctx)

	// 启动HTTP服务
	port := cfg.App.Port
	log.Printf("Stock service starting on port %d", port)

	go func() {
		if err := app.Engine.Run(fmt.Sprintf(":%d", port)); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 取消Consumer
	cancel()

	// 关闭连接
	app.Consumer.Close()
	app.Producer.Close()

	log.Println("Server exited")
}
