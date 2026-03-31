package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, gin.H{
		"code": -1,
		"message": message,
	})
}
