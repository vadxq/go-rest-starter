package models

import "gorm.io/gorm"

// User 用户模型
type User struct {
	gorm.Model
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	Email    string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password string `gorm:"type:varchar(100);not null" json:"-"`
	Role     string `gorm:"type:varchar(20);default:'user'" json:"role"`
}
