package model

import "time"

type PaperQuestion struct {
	ID         uint64    `gorm:"primarykey" json:"id"`
	PaperID    uint64    `gorm:"not null;index" json:"paper_id"`
	QuestionID uint64    `gorm:"not null" json:"question_id"`
	SortOrder  int       `gorm:"type:int;default:0" json:"sort_order"`
	Score      int       `gorm:"type:int;not null" json:"score"`
	CreatedAt  time.Time `json:"created_at"`

	// 关联的题目信息（用于预览，不写入 paper_questions 表）
	Question Question `gorm:"foreignKey:QuestionID" json:"-"`
}

func (PaperQuestion) TableName() string {
	return "paper_questions"
}
