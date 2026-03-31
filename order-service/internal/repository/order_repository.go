package repository

import (
	"fmt"
	"os"

	"order-service/internal/model"

	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create 创建订单
func (r *OrderRepository) Create(order *model.Order) error {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "order-service"
	}
	fmt.Printf("[%s] Creating order: %d for user %d, product %d\n", instanceID, order.ID, order.UserID, order.ProductID)
	return r.db.Create(order).Error
}

// FindByID 根据ID查询订单
func (r *OrderRepository) FindByID(id int64) (*model.Order, error) {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "order-service"
	}
	fmt.Printf("[%s] Finding order by ID: %d\n", instanceID, id)

	var order model.Order
	err := r.db.First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByUserID 根据用户ID查询订单列表
func (r *OrderRepository) FindByUserID(userID int64) ([]model.Order, error) {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "order-service"
	}
	fmt.Printf("[%s] Finding orders by user ID: %d\n", instanceID, userID)

	var orders []model.Order
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// FindByUserIDAndProductID 根据用户ID和商品ID查询订单（用于幂等性检查）
func (r *OrderRepository) FindByUserIDAndProductID(userID, productID int64) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateStatus 更新订单状态
func (r *OrderRepository) UpdateStatus(orderID int64, status model.OrderStatus) error {
	return r.db.Model(&model.Order{}).Where("id = ?", orderID).Update("status", status).Error
}
