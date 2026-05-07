package model

import "time"

type OperationLog struct {
	ID        uint64    `gorm:"primarykey" json:"id"`
	UserID    uint64    `gorm:"not null;index" json:"user_id"`
	Username  string    `gorm:"type:varchar(50)" json:"username"`
	Action    string    `gorm:"type:varchar(100);not null" json:"action"`
	Target    string    `gorm:"type:varchar(100)" json:"target"`
	Detail    string    `gorm:"type:text" json:"detail"`
	IP        string    `gorm:"type:varchar(50)" json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

func (OperationLog) TableName() string {
	return "operation_logs"
}
