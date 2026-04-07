package router

import (
	"net/http"
	"os"

	"stock-service/internal/handler"

	"github.com/gin-gonic/gin"
)

func New(stockHandler *handler.StockHandler) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		instanceID := os.Getenv("INSTANCE_ID")
		if instanceID == "" {
			instanceID = "stock-service"
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "instance": instanceID})
	})

	// API路由
	api := r.Group("/api/v1")
	{
		stock := api.Group("/stock")
		{
			stock.POST("/init", stockHandler.InitStock)
			stock.GET("/info", stockHandler.GetStock)
			stock.POST("/reserve", stockHandler.ReserveStock)
			stock.POST("/confirm", stockHandler.ConfirmDeduct)
			stock.POST("/cancel", stockHandler.CancelReserve)
		}
	}

	return r
}
