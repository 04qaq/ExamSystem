package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleAdmin   = 1
	RoleTeacher = 2
	RoleStudent = 3

	UserStatusActive   = 1
	UserStatusDisabled = 0
)

type User struct {
	ID           uint64         `gorm:"primarykey" json:"id"`
	Username     string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"type:varchar(255);not null" json:"-"`
	Role         int8           `gorm:"type:tinyint;not null;default:3" json:"role"`
	RealName     string         `gorm:"type:varchar(50)" json:"real_name"`
	Status       int8           `gorm:"type:tinyint;not null;default:1" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}
