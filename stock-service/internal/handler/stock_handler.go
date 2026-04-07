package handler

import (
	"net/http"
	"strconv"

	"stock-service/internal/model"
	"stock-service/internal/pkg/response"
	"stock-service/internal/service"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	stockService *service.StockService
}

func NewStockHandler(stockService *service.StockService) *StockHandler {
	return &StockHandler{stockService: stockService}
}

// InitStock 初始化库存
func (h *StockHandler) InitStock(c *gin.Context) {
	var req struct {
		ProductID int64 `json:"product_id" binding:"required"`
		Stock     int   `json:"stock" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.stockService.InitStock(c.Request.Context(), req.ProductID, req.Stock); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "stock initialized", "product_id": req.ProductID, "stock": req.Stock})
}

// GetStock 获取库存
func (h *StockHandler) GetStock(c *gin.Context) {
	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		response.Error(c, http.StatusBadRequest, "product_id is required")
		return
	}

	productID, _ := strconv.ParseInt(productIDStr, 10, 64)
	info, err := h.stockService.GetStockInfo(c.Request.Context(), productID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, info)
}

// ReserveStock 预扣减库存
func (h *StockHandler) ReserveStock(c *gin.Context) {
	var req model.DeductStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.stockService.ReserveStock(c.Request.Context(), req.OrderID, req.ProductID, req.Quantity)
	if err != nil {
		switch err {
		case service.ErrStockNotEnough:
			response.Error(c, http.StatusBadRequest, "stock not enough")
		case service.ErrStockNotExist:
			response.Error(c, http.StatusNotFound, "stock not exist")
		case service.ErrAlreadyReserved:
			response.Error(c, http.StatusBadRequest, "already reserved")
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, gin.H{"message": "stock reserved", "order_id": req.OrderID, "product_id": req.ProductID})
}

// ConfirmDeduct 确认扣减
func (h *StockHandler) ConfirmDeduct(c *gin.Context) {
	var req model.ConfirmStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.stockService.ConfirmDeduct(c.Request.Context(), req.OrderID, req.ProductID)
	if err != nil {
		switch err {
		case service.ErrReserveNotFound:
			response.Error(c, http.StatusNotFound, "reserve not found")
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, gin.H{"message": "stock confirmed", "order_id": req.OrderID, "product_id": req.ProductID})
}

// CancelReserve 取消预扣减
func (h *StockHandler) CancelReserve(c *gin.Context) {
	var req model.ConfirmStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.stockService.CancelReserve(c.Request.Context(), req.OrderID, req.ProductID)
	if err != nil {
		switch err {
		case service.ErrReserveNotFound:
			response.Error(c, http.StatusNotFound, "reserve not found")
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, gin.H{"message": "reserve canceled", "order_id": req.OrderID, "product_id": req.ProductID})
}

// Health 健康检查
func (h *StockHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok"})
}
