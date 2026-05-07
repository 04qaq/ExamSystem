package dto

// PendingGradeRecord 待批阅记录项
type PendingGradeRecord struct {
	RecordID         uint64 `json:"record_id"`
	PaperID          uint64 `json:"paper_id"`
	PaperTitle       string `json:"paper_title"`
	StudentName      string `json:"student_name"`
	StudentID        uint64 `json:"student_id"`
	SubmitTime       string `json:"submit_time"`
	SubjectiveCount  int    `json:"subjective_count"`
	GradedCount      int    `json:"graded_count"`
	TotalScore       int    `json:"total_score"`
	Status           int8   `json:"status"`
}

// PendingGradeListResponse 待批阅列表
type PendingGradeListResponse struct {
	Total int64                 `json:"total"`
	Items []PendingGradeRecord `json:"items"`
}

// GradeDetailItem 批阅题目项
type GradeDetailItem struct {
	DetailID       uint64 `json:"detail_id"`
	QuestionID     uint64 `json:"question_id"`
	Type           int8   `json:"type"`
	Content        string `json:"content"`
	Options        string `json:"options"`
	Score          int    `json:"score"`
	ProvidedAnswer string `json:"provided_answer"`
	CorrectAnswer  string `json:"correct_answer"`
	IsCorrect      *int8  `json:"is_correct"`
	ScoreGained    int    `json:"score_gained"`
	Comment        string `json:"comment"`
}

// PendingGradeDetailResponse 批阅详情
type PendingGradeDetailResponse struct {
	RecordID    uint64            `json:"record_id"`
	PaperTitle  string            `json:"paper_title"`
	StudentName string            `json:"student_name"`
	TotalScore  int               `json:"total_score"`
	Status      int8              `json:"status"`
	SubmitTime  string            `json:"submit_time"`
	Details     []GradeDetailItem `json:"details"`
}

// GradeItem 单个批阅项
type GradeItem struct {
	DetailID     uint64 `json:"detail_id" binding:"required"`
	ScoreGained  int    `json:"score_gained"`
	Comment      string `json:"comment"`
}

// GradeBatchRequest 批量批阅请求
type GradeBatchRequest struct {
	Grades []GradeItem `json:"grades" binding:"required"`
}

// QuestionStatItem 题目统计
type QuestionStatItem struct {
	QuestionID   uint64  `json:"question_id"`
	Content      string  `json:"content"`
	Type         int8    `json:"type"`
	TotalCount   int     `json:"total_count"`
	CorrectCount int     `json:"correct_count"`
	CorrectRate  float64 `json:"correct_rate"`
	AvgScore     float64 `json:"avg_score"`
	FullScore    int     `json:"full_score"`
}

// PaperStatistics 试卷统计
type PaperStatistics struct {
	PaperID        uint64  `json:"paper_id"`
	PaperTitle     string  `json:"paper_title"`
	TotalStudents  int     `json:"total_students"`
	SubmittedCount int     `json:"submitted_count"`
	AvgScore       float64 `json:"avg_score"`
	MaxScore       int     `json:"max_score"`
	MinScore       int     `json:"min_score"`
	PassCount      int     `json:"pass_count"`
	FailCount      int     `json:"fail_count"`
	PassRate       float64 `json:"pass_rate"`
}

// StatisticsResponse 完整统计
type StatisticsResponse struct {
	Paper         PaperStatistics   `json:"paper"`
	QuestionStats []QuestionStatItem `json:"question_stats"`
	ScoreDist     map[string]int     `json:"score_distribution"`
}
