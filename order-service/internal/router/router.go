package router

import (
	"net/http"
	"os"

	"order-service/internal/handler"

	"github.com/gin-gonic/gin"
)

func New(seckillHandler *handler.SeckillHandler, paymentHandler *handler.PaymentHandler) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		instanceID := os.Getenv("INSTANCE_ID")
		if instanceID == "" {
			instanceID = "order-service"
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "instance": instanceID})
	})

	// API路由
	api := r.Group("/api/v1")
	{
		seckill := api.Group("/seckill")
		{
			seckill.POST("/order", seckillHandler.Seckill)
			seckill.GET("/orders/:id", seckillHandler.GetOrderByID)
			seckill.GET("/orders", seckillHandler.GetOrdersByUserID)
			seckill.GET("/orders/check", seckillHandler.CheckOrder)
			seckill.POST("/stock/init", seckillHandler.InitStock)
			seckill.GET("/stock", seckillHandler.GetStock)
		}

		// 支付相关路由
		payment := api.Group("/payment")
		{
			payment.POST("/pay", paymentHandler.Pay)
			payment.GET("/info", paymentHandler.GetPaymentByOrderID)
			payment.POST("/refund", paymentHandler.Refund)
		}
	}

	return r
}
