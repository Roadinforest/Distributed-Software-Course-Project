package model

import "time"

// PaymentStatus 支付状态
type PaymentStatus int

const (
	PaymentStatusPending   PaymentStatus = 0 // 待支付
	PaymentStatusPaid      PaymentStatus = 1 // 已支付
	PaymentStatusRefunded  PaymentStatus = 2 // 已退款
	PaymentStatusFailed    PaymentStatus = 3 // 支付失败
)

// Payment 支付模型
type Payment struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	OrderID   int64          `gorm:"uniqueIndex;not null" json:"order_id"`
	Amount    int64          `gorm:"not null" json:"amount"`      // 支付金额(分)
	Status    PaymentStatus  `gorm:"default:0" json:"status"`
	PayTime   *time.Time     `json:"pay_time,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// PaymentDTO 支付DTO
type PaymentDTO struct {
	ID      int64         `json:"id"`
	OrderID int64         `json:"order_id"`
	Amount  int64         `json:"amount"`
	Status  PaymentStatus `json:"status"`
	PayTime *time.Time    `json:"pay_time,omitempty"`
}

func (p *Payment) ToDTO() PaymentDTO {
	return PaymentDTO{
		ID:      p.ID,
		OrderID: p.OrderID,
		Amount:  p.Amount,
		Status:  p.Status,
		PayTime: p.PayTime,
	}
}

// PaymentMessage Kafka支付消息
type PaymentMessage struct {
	OrderID   int64  `json:"order_id"`
	PaymentID int64  `json:"payment_id"`
	Action    string `json:"action"` // pay, refund, confirm
	Timestamp int64  `json:"timestamp"`
}

// PayRequest 支付请求
type PayRequest struct {
	OrderID int64 `json:"order_id" binding:"required"`
}

// PayResponse 支付响应
type PayResponse struct {
	PaymentID int64  `json:"payment_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	OrderID int64 `json:"order_id" binding:"required"`
}

// OrderStatusMessage 订单状态消息
type OrderStatusMessage struct {
	OrderID   int64  `json:"order_id"`
	Status    string `json:"status"` // pending, created, paid, canceled, completed
	Timestamp int64  `json:"timestamp"`
}
