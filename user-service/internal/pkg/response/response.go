package response

import "github.com/gin-gonic/gin"

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func JSON(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Body{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
