package handler

import (
	"net/http"

	"order-service/internal/model"
	"order-service/internal/pkg/response"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// Pay 支付订单
func (h *PaymentHandler) Pay(c *gin.Context) {
	var req model.PayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.paymentService.Pay(c.Request.Context(), req.OrderID)
	if err != nil {
		switch err {
		case service.ErrOrderNotFound:
			response.Error(c, http.StatusNotFound, "order not found")
		case service.ErrOrderAlreadyPaid:
			response.Error(c, http.StatusBadRequest, "order already paid")
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, resp)
}

// GetPaymentByOrderID 根据订单ID查询支付信息
func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	orderIDStr := c.Query("order_id")
	if orderIDStr == "" {
		response.Error(c, http.StatusBadRequest, "order_id is required")
		return
	}

	var orderID int64
	if _, err := parseStrToInt64(orderIDStr, &orderID); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order_id")
		return
	}

	payment, err := h.paymentService.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if payment == nil {
		response.Success(c, gin.H{"exists": false})
		return
	}

	response.Success(c, payment.ToDTO())
}

// Refund 退款
func (h *PaymentHandler) Refund(c *gin.Context) {
	var req model.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	err := h.paymentService.Refund(c.Request.Context(), req.OrderID)
	if err != nil {
		switch err {
		case service.ErrOrderNotFound:
			response.Error(c, http.StatusNotFound, "order not found")
		case service.ErrOrderNotPaid:
			response.Error(c, http.StatusBadRequest, "order not paid")
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(c, gin.H{"message": "refund initiated"})
}

func parseStrToInt64(s string, result *int64) (bool, error) {
	var val int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		val = val*10 + int64(c-'0')
	}
	*result = val
	return true, nil
}
