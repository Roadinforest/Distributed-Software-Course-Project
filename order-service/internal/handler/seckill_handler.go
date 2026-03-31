package handler

import (
	"net/http"
	"strconv"

	"order-service/internal/model"
	"order-service/internal/pkg/response"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

type SeckillHandler struct {
	service *service.SeckillService
}

func NewSeckillHandler(svc *service.SeckillService) *SeckillHandler {
	return &SeckillHandler{service: svc}
}

// Seckill 秒杀下单
func (h *SeckillHandler) Seckill(c *gin.Context) {
	var req model.SeckillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.service.Seckill(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrAlreadyOrdered:
			response.Success(c, resp) // 返回已下单信息
		case service.ErrStockNotEnough:
			response.Error(c, http.StatusBadRequest, "stock not enough")
		case service.ErrProductNotFound:
			response.Error(c, http.StatusNotFound, "product not found")
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, resp)
}

// GetOrderByID 根据订单ID查询订单
func (h *SeckillHandler) GetOrderByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "order not found")
		return
	}

	response.Success(c, order)
}

// GetOrdersByUserID 根据用户ID查询订单列表
func (h *SeckillHandler) GetOrdersByUserID(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		response.Error(c, http.StatusBadRequest, "user_id is required")
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user_id")
		return
	}

	orders, err := h.service.GetOrdersByUserID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, orders)
}

// CheckOrder 检查订单是否存在（幂等性检查）
func (h *SeckillHandler) CheckOrder(c *gin.Context) {
	userIDStr := c.Query("user_id")
	productIDStr := c.Query("product_id")

	if userIDStr == "" || productIDStr == "" {
		response.Error(c, http.StatusBadRequest, "user_id and product_id are required")
		return
	}

	userID, _ := strconv.ParseInt(userIDStr, 10, 64)
	productID, _ := strconv.ParseInt(productIDStr, 10, 64)

	order, err := h.service.CheckOrderExists(c.Request.Context(), userID, productID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if order == nil {
		response.Success(c, gin.H{"exists": false})
		return
	}

	response.Success(c, gin.H{
		"exists":   true,
		"order_id": order.ID,
		"status":   order.Status,
	})
}

// InitStock 初始化库存（测试用）
func (h *SeckillHandler) InitStock(c *gin.Context) {
	var req struct {
		ProductID int64 `json:"product_id" binding:"required"`
		Stock     int   `json:"stock" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.InitStock(c.Request.Context(), req.ProductID, req.Stock); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "stock initialized"})
}

// GetStock 获取库存
func (h *SeckillHandler) GetStock(c *gin.Context) {
	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		response.Error(c, http.StatusBadRequest, "product_id is required")
		return
	}

	productID, _ := strconv.ParseInt(productIDStr, 10, 64)
	info, err := h.service.GetStockInfo(c.Request.Context(), productID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, info)
}

// Health 健康检查
func (h *SeckillHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok"})
}
