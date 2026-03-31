package model

import "time"

// OrderStatus 订单状态
type OrderStatus int

const (
	OrderStatusPending   OrderStatus = 0 // 待处理
	OrderStatusCreated  OrderStatus = 1 // 已创建
	OrderStatusCanceled OrderStatus = 2 // 已取消
	OrderStatusFailed   OrderStatus = 3 // 创建失败
)

// Order 订单模型
type Order struct {
	ID        int64       `gorm:"primaryKey" json:"id"`
	UserID    int64       `gorm:"index;not null" json:"user_id"`
	ProductID int64       `gorm:"index;not null" json:"product_id"`
	Quantity  int         `gorm:"default:1" json:"quantity"`
	Status    OrderStatus `gorm:"default:0" json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// OrderDTO 用于API返回
type OrderDTO struct {
	ID        int64       `json:"id"`
	UserID    int64       `json:"user_id"`
	ProductID int64       `json:"product_id"`
	Quantity  int         `json:"quantity"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

func (o *Order) ToDTO() OrderDTO {
	return OrderDTO{
		ID:        o.ID,
		UserID:    o.UserID,
		ProductID: o.ProductID,
		Quantity:  o.Quantity,
		Status:    o.Status,
		CreatedAt: o.CreatedAt,
	}
}

// SeckillMessage Kafka消息结构
type SeckillMessage struct {
	OrderID   int64  `json:"order_id"`
	UserID    int64  `json:"user_id"`
	ProductID int64  `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Timestamp int64  `json:"timestamp"`
}

// SeckillRequest 秒杀请求
type SeckillRequest struct {
	UserID    int64 `json:"user_id" binding:"required"`
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

// SeckillResponse 秒杀响应
type SeckillResponse struct {
	OrderID  int64  `json:"order_id,omitempty"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}
