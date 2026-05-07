package dto

type CreateQuestionRequest struct {
	Type       int8   `json:"type" binding:"required,oneof=1 2 3 4 5"`
	Content    string `json:"content" binding:"required"`
	Options    string `json:"options"`
	Answer     string `json:"answer" binding:"required"`
	Score      int    `json:"score" binding:"required,min=1"`
	Difficulty int8   `json:"difficulty" binding:"required,oneof=1 2 3"`
	Tags       string `json:"tags"`
}

type UpdateQuestionRequest struct {
	Type       int8   `json:"type" binding:"required,oneof=1 2 3 4 5"`
	Content    string `json:"content" binding:"required"`
	Options    string `json:"options"`
	Answer     string `json:"answer" binding:"required"`
	Score      int    `json:"score" binding:"required,min=1"`
	Difficulty int8   `json:"difficulty" binding:"required,oneof=1 2 3"`
	Tags       string `json:"tags"`
}

type QuestionQuery struct {
	Type       int8   `form:"type"`
	Difficulty int8   `form:"difficulty"`
	Tag        string `form:"tag"`
	Keyword    string `form:"keyword"`
	Page       int    `form:"page" binding:"required,min=1"`
	PageSize   int    `form:"page_size" binding:"required,min=1,max=100"`
}

type QuestionResponse struct {
	ID         uint64 `json:"id"`
	Type       int8   `json:"type"`
	Content    string `json:"content"`
	Options    string `json:"options"`
	Answer     string `json:"answer"`
	Score      int    `json:"score"`
	Difficulty int8   `json:"difficulty"`
	Tags       string `json:"tags"`
	CreatorID  uint64 `json:"creator_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type QuestionListResponse struct {
	Total int64              `json:"total"`
	Items []QuestionResponse `json:"items"`
}

type ImportQuestionItem struct {
	Type       int8   `json:"type" binding:"required,oneof=1 2 3 4 5"`
	Content    string `json:"content" binding:"required"`
	Options    string `json:"options"`
	Answer     string `json:"answer" binding:"required"`
	Score      int    `json:"score" binding:"required,min=1"`
	Difficulty int8   `json:"difficulty" binding:"required,oneof=1 2 3"`
	Tags       string `json:"tags"`
}

type ImportRequest struct {
	Questions []ImportQuestionItem `json:"questions" binding:"required,min=1,max=1000"`
}

type ImportResponse struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Fail    int `json:"fail"`
}
