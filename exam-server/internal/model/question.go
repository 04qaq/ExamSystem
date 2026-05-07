package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	QuestionTypeSingle   = 1
	QuestionTypeMultiple = 2
	QuestionTypeJudge    = 3
	QuestionTypeFill     = 4
	QuestionTypeShort    = 5
)

const (
	DifficultyEasy   = 1
	DifficultyMedium = 2
	DifficultyHard   = 3
)

type Question struct {
	ID         uint64         `gorm:"primarykey" json:"id"`
	Type       int8           `gorm:"type:tinyint;not null" json:"type"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Options    string         `gorm:"type:json" json:"options"`
	Answer     string         `gorm:"type:varchar(255);not null" json:"answer"`
	Score      int            `gorm:"type:int;not null;default:5" json:"score"`
	Difficulty int8           `gorm:"type:tinyint;not null;default:1" json:"difficulty"`
	Tags       string         `gorm:"type:varchar(255)" json:"tags"`
	CreatorID  uint64         `gorm:"not null" json:"creator_id"`
	IsDeleted  int8           `gorm:"type:tinyint(1);default:0" json:"is_deleted"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Question) TableName() string {
	return "questions"
}
