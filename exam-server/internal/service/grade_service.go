package service

import (
	"errors"
	"fmt"
	"math"
	"sort"

	"exam-server/internal/dto"
	"exam-server/internal/model"
	"exam-server/internal/repository"
	"exam-server/pkg/errcode"

	"github.com/xuri/excelize/v2"
)

type GradeService struct {
	gradeRepo *repository.GradeRepo
	paperRepo *repository.PaperRepo
	userRepo  *repository.UserRepo
	pqRepo    *repository.PaperQuestionRepo
	examRepo  *repository.ExamRepo
}

func NewGradeService() *GradeService {
	return &GradeService{
		gradeRepo: repository.NewGradeRepo(),
		paperRepo: repository.NewPaperRepo(),
		userRepo:  repository.NewUserRepo(),
		pqRepo:    repository.NewPaperQuestionRepo(),
		examRepo:  repository.NewExamRepo(),
	}
}

// GetPendingList 获取待批阅列表
func (s *GradeService) GetPendingList(teacherID uint64, page, pageSize int) (*dto.PendingGradeListResponse, int, error) {
	records, total, err := s.gradeRepo.GetPendingRecords(teacherID, page, pageSize)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	// 批量加载关联数据，避免 N+1
	paperIDs := make([]uint64, 0, len(records))
	userIDs := make([]uint64, 0, len(records))
	for _, r := range records {
		paperIDs = append(paperIDs, r.PaperID)
		userIDs = append(userIDs, r.UserID)
	}

	paperMap := make(map[uint64]*model.Paper)
	for _, pid := range paperIDs {
		p, e := s.paperRepo.FindByID(pid)
		if e == nil {
			paperMap[pid] = p
		}
	}

	userMap := make(map[uint64]*model.User)
	for _, uid := range userIDs {
		u, e := s.userRepo.FindByID(uid)
		if e == nil {
			userMap[uid] = u
		}
	}

	items := make([]dto.PendingGradeRecord, 0, len(records))
	for _, r := range records {
		paperTitle := ""
		if p, ok := paperMap[r.PaperID]; ok {
			paperTitle = p.Title
		}
		studentName := ""
		if u, ok := userMap[r.UserID]; ok {
			studentName = u.RealName
		}

		details, _ := s.gradeRepo.GetAllDetailsByRecord(r.ID)
		subjectiveCount := 0
		gradedCount := 0

		questions, _ := s.pqRepo.GetQuestionsWithDetail(r.PaperID)
		qTypeMap := make(map[uint64]int8)
		for _, q := range questions {
			qTypeMap[q.QuestionID] = q.Question.Type
		}

		for _, d := range details {
			qType, ok := qTypeMap[d.QuestionID]
			if !ok {
				continue
			}
			if qType == model.QuestionTypeFill || qType == model.QuestionTypeShort {
				subjectiveCount++
				if d.IsCorrect != nil {
					gradedCount++
				}
			}
		}

		submitTime := ""
		if r.SubmitTime != nil {
			submitTime = r.SubmitTime.Format("2006-01-02 15:04:05")
		}

		items = append(items, dto.PendingGradeRecord{
			RecordID:        r.ID,
			PaperID:         r.PaperID,
			PaperTitle:      paperTitle,
			StudentName:     studentName,
			StudentID:       r.UserID,
			SubmitTime:      submitTime,
			SubjectiveCount: subjectiveCount,
			GradedCount:     gradedCount,
			TotalScore:      r.TotalScore,
			Status:          r.Status,
		})
	}

	return &dto.PendingGradeListResponse{
		Total: total,
		Items: items,
	}, errcode.Success, nil
}

// GetGradeDetail 获取批阅详情
func (s *GradeService) GetGradeDetail(recordID uint64, teacherID uint64) (*dto.PendingGradeDetailResponse, int, error) {
	record, err := s.examRepo.FindRecordByID(recordID)
	if err != nil {
		return nil, errcode.ExamRecordNotFound, errors.New("考试记录不存在")
	}

	paper, err := s.paperRepo.FindByID(record.PaperID)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限批阅该试卷")
	}

	if record.Status == model.ExamStatusInProgress {
		return nil, errcode.ExamNotSubmitted, errors.New("考试尚未提交")
	}

	student, _ := s.userRepo.FindByID(record.UserID)
	studentName := ""
	if student != nil {
		studentName = student.RealName
	}

	questions, _ := s.pqRepo.GetQuestionsWithDetail(record.PaperID)
	qMap := make(map[uint64]model.PaperQuestion)
	for _, q := range questions {
		qMap[q.QuestionID] = q
	}

	details, err := s.gradeRepo.GetAllDetailsByRecord(recordID)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	itemDetails := make([]dto.GradeDetailItem, len(details))
	for i, d := range details {
		item := dto.GradeDetailItem{
			DetailID:       d.ID,
			QuestionID:     d.QuestionID,
			ProvidedAnswer: d.ProvidedAnswer,
			IsCorrect:      d.IsCorrect,
			ScoreGained:    d.ScoreGained,
			Comment:        d.Comment,
		}
		if pq, ok := qMap[d.QuestionID]; ok {
			q := pq.Question
			item.Type = q.Type
			item.Content = q.Content
			item.Options = q.Options
			item.Score = pq.Score
			item.CorrectAnswer = q.Answer
		}
		itemDetails[i] = item
	}

	submitTime := ""
	if record.SubmitTime != nil {
		submitTime = record.SubmitTime.Format("2006-01-02 15:04:05")
	}

	return &dto.PendingGradeDetailResponse{
		RecordID:    record.ID,
		PaperTitle:  paper.Title,
		StudentName: studentName,
		TotalScore:  record.TotalScore,
		Status:      record.Status,
		SubmitTime:  submitTime,
		Details:     itemDetails,
	}, errcode.Success, nil
}

// GradeSubmit 提交批阅
func (s *GradeService) GradeSubmit(recordID uint64, teacherID uint64, req dto.GradeBatchRequest) (int, error) {
	record, err := s.examRepo.FindRecordByID(recordID)
	if err != nil {
		return errcode.ExamRecordNotFound, errors.New("考试记录不存在")
	}

	paper, err := s.paperRepo.FindByID(record.PaperID)
	if err != nil {
		return errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return errcode.NoPermission, errors.New("无权限批阅该试卷")
	}
	if record.Status == model.ExamStatusInProgress {
		return errcode.ExamNotSubmitted, errors.New("考试尚未提交")
	}

	questions, _ := s.pqRepo.GetQuestionsWithDetail(record.PaperID)
	qTypeMap := make(map[uint64]int8)
	qScoreMap := make(map[uint64]int)
	for _, q := range questions {
		qTypeMap[q.QuestionID] = q.Question.Type
		qScoreMap[q.QuestionID] = q.Score
	}

	allDetails, _ := s.gradeRepo.GetAllDetailsByRecord(recordID)
	detailMap := make(map[uint64]model.AnswerDetail)
	for _, d := range allDetails {
		detailMap[d.ID] = d
	}

	for _, g := range req.Grades {
		detail, ok := detailMap[g.DetailID]
		if !ok {
			continue
		}
		qType, typeOk := qTypeMap[detail.QuestionID]
		if !typeOk {
			continue
		}
		if qType != model.QuestionTypeFill && qType != model.QuestionTypeShort {
			continue
		}
		maxScore := qScoreMap[detail.QuestionID]
		if g.ScoreGained > maxScore {
			g.ScoreGained = maxScore
		}
		if g.ScoreGained < 0 {
			g.ScoreGained = 0
		}
		if err := s.gradeRepo.UpdateGrade(g.DetailID, g.ScoreGained, g.Comment); err != nil {
			return errcode.UnknownError, err
		}
	}

	updatedDetails, _ := s.gradeRepo.GetAllDetailsByRecord(recordID)
	allGraded := true
	totalScore := 0
	for _, d := range updatedDetails {
		qType, ok := qTypeMap[d.QuestionID]
		if !ok {
			totalScore += d.ScoreGained
			continue
		}
		if qType == model.QuestionTypeFill || qType == model.QuestionTypeShort {
			if d.IsCorrect == nil {
				allGraded = false
			}
		}
		totalScore += d.ScoreGained
	}

	newStatus := record.Status
	if allGraded {
		newStatus = model.ExamStatusGraded
	}

	if err := s.gradeRepo.UpdateRecordStatus(recordID, newStatus, totalScore); err != nil {
		return errcode.UnknownError, err
	}

	return errcode.Success, nil
}

// GetPaperStatistics 获取试卷统计
func (s *GradeService) GetPaperStatistics(paperID uint64, teacherID uint64) (*dto.StatisticsResponse, int, error) {
	paper, err := s.paperRepo.FindByID(paperID)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限")
	}

	records, err := s.gradeRepo.GetRecordsByPaperID(paperID)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	submittedCount := len(records)
	if submittedCount == 0 {
		return &dto.StatisticsResponse{
			Paper: dto.PaperStatistics{
				PaperID:    paperID,
				PaperTitle: paper.Title,
			},
			QuestionStats: []dto.QuestionStatItem{},
			ScoreDist:     make(map[string]int),
		}, errcode.Success, nil
	}

	passCount := 0
	failCount := 0
	totalScores := 0
	maxScore := records[0].TotalScore
	minScore := records[0].TotalScore

	passThreshold := int(math.Floor(float64(paper.TotalScore) * 0.6))

	for _, r := range records {
		s := r.TotalScore
		totalScores += s
		if s > maxScore {
			maxScore = s
		}
		if s < minScore {
			minScore = s
		}
		if s >= passThreshold {
			passCount++
		} else {
			failCount++
		}
	}
	avgScore := float64(totalScores) / float64(submittedCount)

	// 批量加载所有答题详情，避免 N+1
	recordIDs := make([]uint64, len(records))
	for i, r := range records {
		recordIDs[i] = r.ID
	}
	allDetails := make(map[uint64][]model.AnswerDetail)
	for _, rid := range recordIDs {
		d, _ := s.gradeRepo.GetAllDetailsByRecord(rid)
		allDetails[rid] = d
	}

	questions, _ := s.pqRepo.GetQuestionsWithDetail(paperID)
	qStats := make([]dto.QuestionStatItem, 0, len(questions))
	for _, pq := range questions {
		q := pq.Question
		if q.Type > model.QuestionTypeJudge {
			qStats = append(qStats, dto.QuestionStatItem{
				QuestionID: q.ID,
				Content:    q.Content,
				Type:       q.Type,
				FullScore:  pq.Score,
			})
			continue
		}

		totalCount := 0
		correctCount := 0
		scoreSum := 0

		for _, r := range records {
			for _, d := range allDetails[r.ID] {
				if d.QuestionID == q.ID {
					totalCount++
					scoreSum += d.ScoreGained
					if d.IsCorrect != nil && *d.IsCorrect == 1 {
						correctCount++
					}
				}
			}
		}

		correctRate := 0.0
		avgQScore := 0.0
		if totalCount > 0 {
			correctRate = float64(correctCount) / float64(totalCount) * 100
			avgQScore = float64(scoreSum) / float64(totalCount)
		}

		qStats = append(qStats, dto.QuestionStatItem{
			QuestionID:   q.ID,
			Content:      q.Content,
			Type:         q.Type,
			TotalCount:   totalCount,
			CorrectCount: correctCount,
			CorrectRate:  math.Round(correctRate*100) / 100,
			AvgScore:     math.Round(avgQScore*100) / 100,
			FullScore:    pq.Score,
		})
	}

	scoreDist := make(map[string]int)
	for _, r := range records {
		s := r.TotalScore
		percent := float64(s) / float64(paper.TotalScore) * 100
		var bucket string
		switch {
		case percent < 30:
			bucket = "0-30%"
		case percent < 60:
			bucket = "30-60%"
		case percent < 70:
			bucket = "60-70%"
		case percent < 80:
			bucket = "70-80%"
		case percent < 90:
			bucket = "80-90%"
		default:
			bucket = "90-100%"
		}
		scoreDist[bucket]++
	}

	for _, b := range []string{"0-30%", "30-60%", "60-70%", "70-80%", "80-90%", "90-100%"} {
		if _, ok := scoreDist[b]; !ok {
			scoreDist[b] = 0
		}
	}

	return &dto.StatisticsResponse{
		Paper: dto.PaperStatistics{
			PaperID:        paperID,
			PaperTitle:     paper.Title,
			TotalStudents:  submittedCount,
			SubmittedCount: submittedCount,
			AvgScore:       math.Round(avgScore*100) / 100,
			MaxScore:       maxScore,
			MinScore:       minScore,
			PassCount:      passCount,
			FailCount:      failCount,
			PassRate:       math.Round(float64(passCount)/float64(submittedCount)*10000) / 100,
		},
		QuestionStats: qStats,
		ScoreDist:     scoreDist,
	}, errcode.Success, nil
}

// ExportExcel 导出成绩 Excel
func (s *GradeService) ExportExcel(paperID uint64, teacherID uint64) ([]byte, string, int, error) {
	paper, err := s.paperRepo.FindByID(paperID)
	if err != nil {
		return nil, "", errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, "", errcode.NoPermission, errors.New("无权限")
	}

	records, err := s.gradeRepo.GetRecordsByPaperID(paperID)
	if err != nil {
		return nil, "", errcode.UnknownError, err
	}

	questions, _ := s.pqRepo.GetQuestionsWithDetail(paperID)
	sort.Slice(questions, func(i, j int) bool {
		return questions[i].SortOrder < questions[j].SortOrder
	})

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Sheet1"
	f.SetCellValue(sheet, "A1", "序号")
	f.SetCellValue(sheet, "B1", "学生姓名")
	f.SetCellValue(sheet, "C1", "得分")

	col := 'D'
	colIdx := make(map[uint64]string)
	for _, q := range questions {
		letter := string(col)
		colIdx[q.QuestionID] = letter
		title := fmt.Sprintf("第%d题(%d分)", q.SortOrder, q.Score)
		f.SetCellValue(sheet, fmt.Sprintf("%s1", letter), title)
		col++
	}
	f.SetCellValue(sheet, fmt.Sprintf("%c1", col), "提交时间")

	// 批量加载学生姓名
	userIDs := make([]uint64, len(records))
	for i, r := range records {
		userIDs[i] = r.UserID
	}
	userNameMap := make(map[uint64]string)
	for _, uid := range userIDs {
		u, e := s.userRepo.FindByID(uid)
		if e == nil {
			userNameMap[uid] = u.RealName
		}
	}

	for i, r := range records {
		row := i + 2
		f.SetCellInt(sheet, fmt.Sprintf("A%d", row), int64(i+1))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), userNameMap[r.UserID])
		f.SetCellInt(sheet, fmt.Sprintf("C%d", row), int64(r.TotalScore))

		details, _ := s.gradeRepo.GetAllDetailsByRecord(r.ID)
		for _, d := range details {
			if letter, ok := colIdx[d.QuestionID]; ok {
				f.SetCellInt(sheet, fmt.Sprintf("%s%d", letter, row), int64(d.ScoreGained))
			}
		}

		submitTime := ""
		if r.SubmitTime != nil {
			submitTime = r.SubmitTime.Format("2006-01-02 15:04:05")
		}
		f.SetCellValue(sheet, fmt.Sprintf("%c%d", col, row), submitTime)
	}

	f.SetColWidth(sheet, "A", "A", 6)
	f.SetColWidth(sheet, "B", "B", 15)
	f.SetColWidth(sheet, "C", "C", 8)
	f.SetColWidth(sheet, string(col), string(col), 20)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", errcode.UnknownError, err
	}

	filename := fmt.Sprintf("%s_成绩.xlsx", paper.Title)
	return buf.Bytes(), filename, errcode.Success, nil
}
