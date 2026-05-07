package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"exam-server/internal/cache"
	"exam-server/internal/database"
	"exam-server/internal/dto"
	"exam-server/internal/model"
	"exam-server/internal/repository"
	"exam-server/pkg/errcode"
)

type ExamService struct {
	examRepo      *repository.ExamRepo
	paperRepo     *repository.PaperRepo
	pqRepo        *repository.PaperQuestionRepo
	questionRepo  *repository.QuestionRepo
}

func NewExamService() *ExamService {
	return &ExamService{
		examRepo:     repository.NewExamRepo(),
		paperRepo:    repository.NewPaperRepo(),
		pqRepo:       repository.NewPaperQuestionRepo(),
		questionRepo: repository.NewQuestionRepo(),
	}
}

// GetAvailablePapers 获取可考试卷列表
func (s *ExamService) GetAvailablePapers(userID uint64) ([]dto.AvailablePaperResponse, int, error) {
	// 查询所有已发布试卷
	var papers []model.Paper
	now := time.Now()

	query := database.DB.Model(&model.Paper{}).
		Where("status = ? AND start_time <= ? AND end_time >= ?",
			model.PaperStatusPublished, now, now).
		Order("start_time ASC")
	if err := query.Find(&papers).Error; err != nil {
		return nil, errcode.UnknownError, err
	}

	// 查询学生已提交的试卷
	submittedMap, err := s.examRepo.GetSubmittedPaperIDsByUser(userID)
	if err != nil {
		submittedMap = make(map[uint64]uint64)
	}

	// 查询学生进行中的考试
	var inProgressRecords []model.ExamRecord
	database.DB.Model(&model.ExamRecord{}).
		Where("user_id = ? AND status = ?", userID, model.ExamStatusInProgress).
		Find(&inProgressRecords)
	inProgressMap := make(map[uint64]uint64)
	for _, r := range inProgressRecords {
		inProgressMap[r.PaperID] = r.ID
	}

	resp := make([]dto.AvailablePaperResponse, 0, len(papers))
	for _, p := range papers {
		item := dto.AvailablePaperResponse{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			Duration:    p.Duration,
			TotalScore:  p.TotalScore,
			StartTime:   p.StartTime.Format("2006-01-02 15:04:05"),
			EndTime:     p.EndTime.Format("2006-01-02 15:04:05"),
		}
		if recordID, ok := inProgressMap[p.ID]; ok {
			item.MyRecordID = &recordID
		}
		// 已提交的也标记
		if recordID, ok := submittedMap[p.ID]; ok {
			item.MyRecordID = &recordID
		}
		resp = append(resp, item)
	}

	return resp, errcode.Success, nil
}

// StartExam 开始考试
func (s *ExamService) StartExam(paperID uint64, userID uint64) (*dto.StartExamResponse, int, error) {
	paper, err := s.paperRepo.FindByID(paperID)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.Status != model.PaperStatusPublished {
		return nil, errcode.PaperNotPublished, errors.New("试卷未发布")
	}

	now := time.Now()
	if now.Before(paper.StartTime) || now.After(paper.EndTime) {
		return nil, errcode.PaperTimeInvalid, errors.New("不在考试有效时间内")
	}

	// 检查是否有进行中的考试
	existing, _ := s.examRepo.FindInProgressByUserAndPaper(userID, paperID)
	if existing != nil {
		// 返回进行中的考试，继续答题
		return s.buildExamResponse(existing, paper)
	}

	// 获取试卷题目
	questions, err := s.pqRepo.GetQuestionsWithDetail(paperID)
	if err != nil {
		return nil, errcode.UnknownError, err
	}
	if len(questions) == 0 {
		return nil, errcode.PaperNoQuestions, errors.New("试卷中无题目")
	}

	// 创建考试记录
	record := &model.ExamRecord{
		UserID:          userID,
		PaperID:         paperID,
		StartTime:       now,
		Status:          model.ExamStatusInProgress,
		AnswersSnapshot: "{}",
	}
	if err := s.examRepo.CreateRecord(record); err != nil {
		return nil, errcode.UnknownError, err
	}

	return s.buildExamResponseFromQuestions(record, paper, questions)
}

// SaveAnswer 保存答案
func (s *ExamService) SaveAnswer(recordID uint64, userID uint64, req dto.SaveAnswerRequest) (int, error) {
	record, err := s.examRepo.FindRecordByID(recordID)
	if err != nil {
		return errcode.ExamRecordNotFound, errors.New("考试记录不存在")
	}
	if record.UserID != userID {
		return errcode.NoPermission, errors.New("无权限")
	}
	if record.Status != model.ExamStatusInProgress {
		return errcode.ExamAlreadySubmitted, errors.New("试卷已提交")
	}

	// 写入缓存
	cache.GlobalCache.Set(recordID, req.QuestionID, req.Answer)

	// 直接持久化到 DB（双写保证不丢）
	detail := &model.AnswerDetail{
		ExamRecordID:   recordID,
		QuestionID:     req.QuestionID,
		ProvidedAnswer: req.Answer,
	}
	if err := s.examRepo.UpsertAnswer(detail); err != nil {
		return errcode.UnknownError, err
	}

	return errcode.Success, nil
}

// SubmitExam 提交试卷并自动判分
func (s *ExamService) SubmitExam(recordID uint64, userID uint64, req dto.SubmitExamRequest) (*dto.SubmitResponse, int, error) {
	record, err := s.examRepo.FindRecordByID(recordID)
	if err != nil {
		return nil, errcode.ExamRecordNotFound, errors.New("考试记录不存在")
	}
	if record.UserID != userID {
		return nil, errcode.NoPermission, errors.New("无权限")
	}
	if record.Status != model.ExamStatusInProgress {
		return nil, errcode.ExamAlreadySubmitted, errors.New("试卷已提交")
	}

	// 先刷缓存到 DB
	cache.GlobalCache.FlushRecord(recordID)

	// 获取所有答案
	details, err := s.examRepo.GetAnswerDetails(recordID)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	// 获取试卷题目（含正确答案）
	questions, err := s.pqRepo.GetQuestionsWithDetail(record.PaperID)
	if err != nil {
		return nil, errcode.UnknownError, err
	}
	qMap := make(map[uint64]model.PaperQuestion)
	for _, q := range questions {
		qMap[q.QuestionID] = q
	}

	// 逐题判分
	now := time.Now()
	correctCount := 0
	wrongCount := 0
	totalScore := 0

	for i := range details {
		pq, ok := qMap[details[i].QuestionID]
		if !ok {
			continue
		}
		q := pq.Question

		// 客观题自动判分
		if q.Type == model.QuestionTypeSingle || q.Type == model.QuestionTypeMultiple || q.Type == model.QuestionTypeJudge {
			correct := gradeQuestion(q.Type, details[i].ProvidedAnswer, q.Answer)
			if correct {
				details[i].ScoreGained = pq.Score
				details[i].IsCorrect = int8Ptr(1)
				correctCount++
			} else {
				details[i].ScoreGained = 0
				details[i].IsCorrect = int8Ptr(0)
				wrongCount++
			}
		} else {
			// 主观题标记为待批阅
			details[i].ScoreGained = 0
			details[i].IsCorrect = nil
		}
		totalScore += details[i].ScoreGained

		// 更新到数据库
		database.DB.Model(&model.AnswerDetail{}).
			Where("id = ?", details[i].ID).
			Updates(map[string]interface{}{
				"score_gained": details[i].ScoreGained,
				"is_correct":   details[i].IsCorrect,
			})
	}

	// 生成答案快照（记录切屏次数等信息）
	snapshotData := fmt.Sprintf(`{"tab_switch_count":%d}`, req.TabSwitchCount)

	// 更新考试记录
	record.Status = model.ExamStatusSubmitted
	record.SubmitTime = &now
	record.TotalScore = totalScore
	record.AnswersSnapshot = snapshotData
	if err := s.examRepo.UpdateRecord(record); err != nil {
		return nil, errcode.UnknownError, err
	}

	return &dto.SubmitResponse{
		TotalScore:   totalScore,
		CorrectCount: correctCount,
		WrongCount:   wrongCount,
		SubmittedAt:  now.Format("2006-01-02 15:04:05"),
	}, errcode.Success, nil
}

// GetRecords 获取考试记录列表
func (s *ExamService) GetRecords(userID uint64, page, pageSize int) (*dto.ExamRecordListResponse, int, error) {
	records, total, err := s.examRepo.ListRecordsByUser(userID, page, pageSize)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	// 获取试卷标题
	items := make([]dto.ExamRecordResponse, len(records))
	for i, r := range records {
		paper, _ := s.paperRepo.FindByID(r.PaperID)
		paperTitle := ""
		if paper != nil {
			paperTitle = paper.Title
		}
		items[i] = dto.ExamRecordResponse{
			ID:         r.ID,
			PaperTitle: paperTitle,
			PaperID:    r.PaperID,
			TotalScore: r.TotalScore,
			Status:     r.Status,
			StartTime:  r.StartTime.Format("2006-01-02 15:04:05"),
		}
		if r.SubmitTime != nil {
			t := r.SubmitTime.Format("2006-01-02 15:04:05")
			items[i].SubmitTime = &t
		}
	}

	return &dto.ExamRecordListResponse{
		Total: total,
		Items: items,
	}, errcode.Success, nil
}

// GetRecordDetail 获取成绩详情
func (s *ExamService) GetRecordDetail(recordID uint64, userID uint64) (*dto.ExamRecordDetailResponse, int, error) {
	record, err := s.examRepo.FindRecordByID(recordID)
	if err != nil {
		return nil, errcode.ExamRecordNotFound, errors.New("考试记录不存在")
	}
	if record.UserID != userID {
		return nil, errcode.NoPermission, errors.New("无权限")
	}

	paper, _ := s.paperRepo.FindByID(record.PaperID)
	paperTitle := ""
	if paper != nil {
		paperTitle = paper.Title
	}

	// 获取答题详情
	details, err := s.examRepo.GetAnswerDetails(recordID)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	// 获取题目信息
	questions, _ := s.pqRepo.GetQuestionsWithDetail(record.PaperID)
	qMap := make(map[uint64]model.PaperQuestion)
	for _, q := range questions {
		qMap[q.QuestionID] = q
	}

	itemDetails := make([]dto.AnswerDetailItem, len(details))
	for i, d := range details {
		item := dto.AnswerDetailItem{
			QuestionID:     d.QuestionID,
			ProvidedAnswer: d.ProvidedAnswer,
			IsCorrect:      d.IsCorrect,
			ScoreGained:    d.ScoreGained,
			Comment:        d.Comment,
		}
		if pq, ok := qMap[d.QuestionID]; ok {
			q := pq.Question
			item.Content = q.Content
			item.Type = q.Type
			item.Options = q.Options
			item.Score = pq.Score
			item.CorrectAnswer = q.Answer
		}
		itemDetails[i] = item
	}

	resp := &dto.ExamRecordDetailResponse{
		ID:         record.ID,
		PaperTitle: paperTitle,
		PaperID:    record.PaperID,
		TotalScore: record.TotalScore,
		Status:     record.Status,
		StartTime:  record.StartTime.Format("2006-01-02 15:04:05"),
		Details:    itemDetails,
	}
	if record.SubmitTime != nil {
		t := record.SubmitTime.Format("2006-01-02 15:04:05")
		resp.SubmitTime = &t
	}

	return resp, errcode.Success, nil
}

// buildExamResponse 从已有记录构建考试响应（继续考试）
func (s *ExamService) buildExamResponse(record *model.ExamRecord, paper *model.Paper) (*dto.StartExamResponse, int, error) {
	questions, err := s.pqRepo.GetQuestionsWithDetail(record.PaperID)
	if err != nil {
		return nil, errcode.UnknownError, err
	}
	return s.buildExamResponseFromQuestions(record, paper, questions)
}

func (s *ExamService) buildExamResponseFromQuestions(record *model.ExamRecord, paper *model.Paper, questions []model.PaperQuestion) (*dto.StartExamResponse, int, error) {
	// 处理已过考试时间
	elapsed := time.Since(record.StartTime)
	remaining := int(paper.Duration)*60 - int(elapsed.Seconds())
	if remaining < 0 {
		remaining = 0
	}

	// 打乱题目顺序
	shuffled := make([]model.PaperQuestion, len(questions))
	copy(shuffled, questions)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	items := make([]dto.ExamQuestionItem, len(shuffled))
	for i, pq := range shuffled {
		q := pq.Question
		options := q.Options

		// 选择题选项随机打乱
		if q.Type == model.QuestionTypeSingle || q.Type == model.QuestionTypeMultiple {
			options = shuffleOptions(q.Options)
		}

		items[i] = dto.ExamQuestionItem{
			SortOrder:  i + 1,
			QuestionID: pq.QuestionID,
			Type:       q.Type,
			Content:    q.Content,
			Options:    options,
			Score:      pq.Score,
		}
	}

	return &dto.StartExamResponse{
		RecordID:        record.ID,
		Questions:       items,
		RemainingSeconds: remaining,
		TotalScore:      paper.TotalScore,
	}, errcode.Success, nil
}

// gradeQuestion 判分：返回是否正确
func gradeQuestion(qType int8, provided, correct string) bool {
	provided = strings.TrimSpace(provided)
	correct = strings.TrimSpace(correct)

	switch qType {
	case model.QuestionTypeSingle, model.QuestionTypeJudge:
		return strings.EqualFold(provided, correct)
	case model.QuestionTypeMultiple:
		// 多选题：选项分割后排序比较
		p := splitAndSort(provided)
		c := splitAndSort(correct)
		return p == c
	}
	return false
}

func splitAndSort(s string) string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

// shuffleOptions 打乱选择题选项显示顺序（保留原有字母标识）
func shuffleOptions(options string) string {
	var opts []string
	if err := json.Unmarshal([]byte(options), &opts); err != nil {
		return options
	}
	if len(opts) <= 2 {
		return options
	}

	rand.Shuffle(len(opts), func(i, j int) {
		opts[i], opts[j] = opts[j], opts[i]
	})

	result, _ := json.Marshal(opts)
	return string(result)
}

func int8Ptr(v int8) *int8 {
	return &v
}
