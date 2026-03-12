package model

import "time"

type User struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"type:varchar(128);not null" json:"-"`
	Email     string    `gorm:"type:varchar(128)" json:"email"`
	Phone     string    `gorm:"type:varchar(20);index" json:"phone"`
	Status    int8      `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
