package router

import (
	"net/http"
	"os"

	"product-service/internal/handler"

	"github.com/gin-gonic/gin"
)

func New(productHandler *handler.ProductHandler) *gin.Engine {
	r := gin.Default()

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		instanceID := os.Getenv("INSTANCE_ID")
		if instanceID == "" {
			instanceID = "product-service"
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "instance": instanceID})
	})

	// API路由
	api := r.Group("/api/v1")
	{
		products := api.Group("/products")
		{
			products.GET("", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.POST("", productHandler.CreateProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}
	}

	return r
}
