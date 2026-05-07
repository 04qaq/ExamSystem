package dto

import "time"

type CreatePaperRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Duration    int       `json:"duration" binding:"required,min=1"`
	TotalScore  int       `json:"total_score" binding:"required,min=1"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
}

type UpdatePaperRequest struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Duration    int       `json:"duration" binding:"required,min=1"`
	TotalScore  int       `json:"total_score" binding:"required,min=1"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
}

type PaperQuery struct {
	Keyword  string `form:"keyword"`
	Status   int8   `form:"status"`
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
}

type PaperResponse struct {
	ID          uint64 `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
	TotalScore  int    `json:"total_score"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Status      int8   `json:"status"`
	CreatorID   uint64 `json:"creator_id"`
	QuestionCnt int    `json:"question_cnt"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type PaperListResponse struct {
	Total int64            `json:"total"`
	Items []PaperResponse  `json:"items"`
}

type AddQuestionItem struct {
	QuestionID uint64 `json:"question_id" binding:"required"`
	Score      int    `json:"score" binding:"required,min=1"`
}

type AddQuestionsRequest struct {
	Questions []AddQuestionItem `json:"questions" binding:"required,min=1"`
}

type RandomSelectRule struct {
	Type       int8 `json:"type" binding:"required,oneof=1 2 3 4 5"`
	Count      int  `json:"count" binding:"required,min=1"`
	TotalScore int  `json:"total_score" binding:"required,min=1"`
}

type RandomSelectRequest struct {
	Rules []RandomSelectRule `json:"rules" binding:"required,min=1"`
}

type PaperQuestionItem struct {
	ID         uint64 `json:"id"`
	PaperID    uint64 `json:"paper_id"`
	QuestionID uint64 `json:"question_id"`
	SortOrder  int    `json:"sort_order"`
	Score      int    `json:"score"`
	// 题目详情（预览时返回）
	QuestionType    int8   `json:"question_type,omitempty"`
	QuestionContent string `json:"question_content,omitempty"`
	QuestionOptions string `json:"question_options,omitempty"`
}

type PaperPreviewResponse struct {
	Paper     PaperResponse      `json:"paper"`
	Questions []PaperQuestionItem `json:"questions"`
}
