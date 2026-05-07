package dto

type AvailablePaperResponse struct {
	ID          uint64  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Duration    int     `json:"duration"`
	TotalScore  int     `json:"total_score"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	MyRecordID  *uint64 `json:"my_record_id,omitempty"` // 进行中的考试记录ID
}

type ExamQuestionItem struct {
	SortOrder int    `json:"sort_order"`
	QuestionID uint64 `json:"question_id"`
	Type      int8   `json:"type"`
	Content   string `json:"content"`
	Options   string `json:"options"`   // 已打乱选项
	Score     int    `json:"score"`
}

type StartExamResponse struct {
	RecordID        uint64             `json:"record_id"`
	Questions       []ExamQuestionItem `json:"questions"`
	RemainingSeconds int               `json:"remaining_seconds"`
	TotalScore      int                `json:"total_score"`
}

type SaveAnswerRequest struct {
	QuestionID uint64 `json:"question_id" binding:"required"`
	Answer     string `json:"answer"`
}

type SubmitExamRequest struct {
	TabSwitchCount int `json:"tab_switch_count"` // 切屏次数
}

type SubmitResponse struct {
	TotalScore    int    `json:"total_score"`
	CorrectCount  int    `json:"correct_count"`
	WrongCount    int    `json:"wrong_count"`
	SubmittedAt   string `json:"submitted_at"`
}

type ExamRecordResponse struct {
	ID         uint64  `json:"id"`
	PaperTitle string  `json:"paper_title"`
	PaperID    uint64  `json:"paper_id"`
	TotalScore int     `json:"total_score"`
	Status     int8    `json:"status"`
	StartTime  string  `json:"start_time"`
	SubmitTime *string `json:"submit_time"`
}

type ExamRecordListResponse struct {
	Total int64                `json:"total"`
	Items []ExamRecordResponse `json:"items"`
}

type AnswerDetailItem struct {
	QuestionID       uint64 `json:"question_id"`
	Content          string `json:"content"`
	Type             int8   `json:"type"`
	Options          string `json:"options"`
	Score            int    `json:"score"`
	ProvidedAnswer   string `json:"provided_answer"`
	CorrectAnswer    string `json:"correct_answer"`
	IsCorrect        *int8  `json:"is_correct"` // nil=主观题待批阅
	ScoreGained      int    `json:"score_gained"`
	Comment          string `json:"comment"`
}

type ExamRecordDetailResponse struct {
	ID         uint64             `json:"id"`
	PaperTitle string             `json:"paper_title"`
	PaperID    uint64             `json:"paper_id"`
	TotalScore int                `json:"total_score"`
	Status     int8               `json:"status"`
	StartTime  string             `json:"start_time"`
	SubmitTime *string            `json:"submit_time"`
	Details    []AnswerDetailItem `json:"details"`
}
