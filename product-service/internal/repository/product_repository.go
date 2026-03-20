package repository

import (
	"fmt"
	"os"

	"product-service/internal/model"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db      *gorm.DB  // 主库 - 写操作
	dbRead  *gorm.DB  // 从库 - 读操作
}

func NewProductRepository(db *gorm.DB, dbRead *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db, dbRead: dbRead}
}

// FindByID 从从库读取数据(读操作)
func (r *ProductRepository) FindByID(id uint) (*model.Product, error) {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "product-service"
	}

	var product model.Product
	// 使用从库进行读操作
	err := r.dbRead.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	fmt.Printf("[%s] Read from SLAVE for product %d\n", instanceID, id)
	return &product, nil
}

// Create 创建商品(写操作)
func (r *ProductRepository) Create(product *model.Product) error {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "product-service"
	}
	fmt.Printf("[%s] Write to MASTER for creating product\n", instanceID)
	return r.db.Create(product).Error
}

// Update 更新商品(写操作)
func (r *ProductRepository) Update(product *model.Product) error {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "product-service"
	}
	fmt.Printf("[%s] Write to MASTER for updating product\n", instanceID)
	return r.db.Save(product).Error
}

// Delete 删除商品(写操作)
func (r *ProductRepository) Delete(id uint) error {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "product-service"
	}
	fmt.Printf("[%s] Write to MASTER for deleting product\n", instanceID)
	return r.db.Delete(&model.Product{}, id).Error
}

// List 从从库读取列表(读操作)
func (r *ProductRepository) List(limit, offset int) ([]model.Product, error) {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "product-service"
	}
	fmt.Printf("[%s] Read from SLAVE for listing products\n", instanceID)

	var products []model.Product
	err := r.dbRead.Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}
