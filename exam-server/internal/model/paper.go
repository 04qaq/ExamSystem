package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	PaperStatusDraft     = 0
	PaperStatusPublished = 1
	PaperStatusClosed    = 2
)

type Paper struct {
	ID          uint64         `gorm:"primarykey" json:"id"`
	Title       string         `gorm:"type:varchar(100);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Duration    int            `gorm:"type:int;not null" json:"duration"`
	TotalScore  int            `gorm:"type:int;not null" json:"total_score"`
	StartTime   time.Time      `gorm:"not null" json:"start_time"`
	EndTime     time.Time      `gorm:"not null" json:"end_time"`
	Status      int8           `gorm:"type:tinyint;default:0" json:"status"`
	CreatorID   uint64         `gorm:"not null" json:"creator_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Paper) TableName() string {
	return "papers"
}
