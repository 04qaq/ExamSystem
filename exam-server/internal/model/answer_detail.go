package model

import "time"

type AnswerDetail struct {
	ID              uint64    `gorm:"primarykey" json:"id"`
	ExamRecordID    uint64    `gorm:"not null;index" json:"exam_record_id"`
	QuestionID      uint64    `gorm:"not null" json:"question_id"`
	ProvidedAnswer  string    `gorm:"type:text" json:"provided_answer"`
	IsCorrect       *int8     `gorm:"type:tinyint" json:"is_correct"`
	ScoreGained     int       `gorm:"type:int;default:0" json:"score_gained"`
	Comment         string    `gorm:"type:varchar(255)" json:"comment"`
	CreatedAt       time.Time `json:"created_at"`
}

func (AnswerDetail) TableName() string {
	return "answer_details"
}
