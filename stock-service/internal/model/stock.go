package model

import "time"

// Stock 库存模型
type Stock struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID int64     `gorm:"uniqueIndex;not null" json:"product_id"`
	Quantity  int       `gorm:"not null;default:0" json:"quantity"`
	Reserved  int       `gorm:"not null;default:0" json:"reserved"`  // 预扣减数量
	Sold      int       `gorm:"not null;default:0" json:"sold"`     // 已售数量
	Version   int       `gorm:"not null;default:0" json:"version"`  // 乐观锁版本
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StockDTO 库存DTO
type StockDTO struct {
	ID        int64 `json:"id"`
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
	Reserved  int   `json:"reserved"`
	Sold      int   `json:"sold"`
	Available int   `json:"available"` // 可用库存
}

func (s *Stock) ToDTO() StockDTO {
	return StockDTO{
		ID:        s.ID,
		ProductID: s.ProductID,
		Quantity:  s.Quantity,
		Reserved:  s.Reserved,
		Sold:      s.Sold,
		Available: s.Quantity - s.Reserved,
	}
}

// StockMessage Kafka消息结构
type StockMessage struct {
	OrderID    int64  `json:"order_id"`
	UserID     int64  `json:"user_id"`
	ProductID  int64  `json:"product_id"`
	Quantity   int    `json:"quantity"`
	Action     string `json:"action"`      // reserve, confirm, cancel
	Timestamp  int64  `json:"timestamp"`
}

// StockRequest 库存操作请求
type StockRequest struct {
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

// StockResponse 库存操作响应
type StockResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Stock   int    `json:"stock,omitempty"`
}

// DeductStockRequest 扣减库存请求
type DeductStockRequest struct {
	OrderID   int64 `json:"order_id" binding:"required"`
	ProductID int64 `json:"product_id" binding:"required"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

// ConfirmStockRequest 确认库存扣减请求
type ConfirmStockRequest struct {
	OrderID   int64 `json:"order_id" binding:"required"`
	ProductID int64 `json:"product_id" binding:"required"`
}
