package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ExamStatusInProgress = 1
	ExamStatusSubmitted  = 2
	ExamStatusGraded     = 3
)

type ExamRecord struct {
	ID              uint64         `gorm:"primarykey" json:"id"`
	UserID          uint64         `gorm:"not null;index" json:"user_id"`
	PaperID         uint64         `gorm:"not null;index" json:"paper_id"`
	StartTime       time.Time      `gorm:"not null" json:"start_time"`
	SubmitTime      *time.Time     `json:"submit_time"`
	TotalScore      int            `gorm:"type:int;default:0" json:"total_score"`
	Status          int8           `gorm:"type:tinyint;default:1" json:"status"`
	AnswersSnapshot string         `gorm:"type:json" json:"answers_snapshot"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ExamRecord) TableName() string {
	return "exam_records"
}
