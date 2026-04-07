package repository

import (
	"context"
	"errors"

	"stock-service/internal/model"

	"gorm.io/gorm"
)

var (
	ErrStockNotFound = errors.New("stock not found")
	ErrStockNotEnough = errors.New("stock not enough")
)

type StockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) *StockRepository {
	return &StockRepository{db: db}
}

// Create 创建库存记录
func (r *StockRepository) Create(stock *model.Stock) error {
	return r.db.Create(stock).Error
}

// FindByProductID 根据商品ID查找库存
func (r *StockRepository) FindByProductID(productID int64) (*model.Stock, error) {
	var stock model.Stock
	err := r.db.Where("product_id = ?", productID).First(&stock).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrStockNotFound
	}
	return &stock, err
}

// UpdateQuantity 更新库存数量（乐观锁）
func (r *StockRepository) UpdateQuantity(ctx context.Context, productID int64, quantity, reserved, sold, version int) error {
	result := r.db.WithContext(ctx).Model(&model.Stock{}).
		Where("product_id = ? AND version = ?", productID, version).
		Updates(map[string]interface{}{
			"quantity": quantity,
			"reserved": reserved,
			"sold":     sold,
			"version":  version + 1,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("update failed: version conflict")
	}
	return nil
}

// IncrementReserved 增加预扣减数量
func (r *StockRepository) IncrementReserved(ctx context.Context, productID int64, quantity int) error {
	result := r.db.WithContext(ctx).Model(&model.Stock{}).
		Where("product_id = ? AND (quantity - reserved) >= ?", productID, quantity).
		Update("reserved", gorm.Expr("reserved + ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrStockNotEnough
	}
	return nil
}

// ConfirmSold 确认销售（reserved转sold）
func (r *StockRepository) ConfirmSold(ctx context.Context, productID int64, quantity int) error {
	result := r.db.WithContext(ctx).Model(&model.Stock{}).
		Where("product_id = ? AND reserved >= ?", productID, quantity).
		Updates(map[string]interface{}{
			"reserved": gorm.Expr("reserved - ?", quantity),
			"sold":     gorm.Expr("sold + ?", quantity),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("confirm sold failed: insufficient reserved stock")
	}
	return nil
}

// CancelReserved 取消预扣减（释放库存）
func (r *StockRepository) CancelReserved(ctx context.Context, productID int64, quantity int) error {
	result := r.db.WithContext(ctx).Model(&model.Stock{}).
		Where("product_id = ? AND reserved >= ?", productID, quantity).
		Updates(map[string]interface{}{
			"reserved": gorm.Expr("reserved - ?", quantity),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("cancel reserved failed: insufficient reserved stock")
	}
	return nil
}

// InitDatabase 初始化数据库
func (r *StockRepository) InitDatabase() error {
	return r.db.AutoMigrate(&model.Stock{})
}
