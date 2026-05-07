package service

import (
	"errors"
	"fmt"
	"math/rand"

	"exam-server/internal/dto"
	"exam-server/internal/model"
	"exam-server/internal/repository"
	"exam-server/pkg/errcode"
)

type PaperService struct {
	paperRepo *repository.PaperRepo
	pqRepo    *repository.PaperQuestionRepo
	questionRepo *repository.QuestionRepo
}

func NewPaperService() *PaperService {
	return &PaperService{
		paperRepo:    repository.NewPaperRepo(),
		pqRepo:       repository.NewPaperQuestionRepo(),
		questionRepo: repository.NewQuestionRepo(),
	}
}

func (s *PaperService) Create(req dto.CreatePaperRequest, creatorID uint64) (*dto.PaperResponse, int, error) {
	if req.StartTime.After(req.EndTime) || req.StartTime.Equal(req.EndTime) {
		return nil, errcode.InvalidParams, errors.New("开始时间必须早于结束时间")
	}

	paper := &model.Paper{
		Title:       req.Title,
		Description: req.Description,
		Duration:    req.Duration,
		TotalScore:  req.TotalScore,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Status:      model.PaperStatusDraft,
		CreatorID:   creatorID,
	}

	if err := s.paperRepo.Create(paper); err != nil {
		return nil, errcode.UnknownError, err
	}

	return toPaperResponse(paper, 0), errcode.Success, nil
}

func (s *PaperService) Update(id uint64, req dto.UpdatePaperRequest, teacherID uint64) (*dto.PaperResponse, int, error) {
	paper, err := s.paperRepo.FindByID(id)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限操作此试卷")
	}
	if paper.Status != model.PaperStatusDraft {
		return nil, errcode.PaperNotDraft, errors.New("仅草稿状态的试卷可修改")
	}
	if req.StartTime.After(req.EndTime) || req.StartTime.Equal(req.EndTime) {
		return nil, errcode.InvalidParams, errors.New("开始时间必须早于结束时间")
	}

	paper.Title = req.Title
	paper.Description = req.Description
	paper.Duration = req.Duration
	paper.TotalScore = req.TotalScore
	paper.StartTime = req.StartTime
	paper.EndTime = req.EndTime

	if err := s.paperRepo.Update(paper); err != nil {
		return nil, errcode.UnknownError, err
	}

	cnt, _ := s.pqRepo.CalcTotalScore(id)
	return toPaperResponse(paper, cnt), errcode.Success, nil
}

func (s *PaperService) Delete(id uint64, teacherID uint64) (int, error) {
	paper, err := s.paperRepo.FindByID(id)
	if err != nil {
		return errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return errcode.NoPermission, errors.New("无权限操作此试卷")
	}
	if paper.Status != model.PaperStatusDraft {
		return errcode.PaperNotDraft, errors.New("仅草稿状态的试卷可删除")
	}

	if err := s.paperRepo.Delete(id); err != nil {
		return errcode.UnknownError, err
	}
	// 同时清理关联
	_ = s.pqRepo.ClearByPaperID(id)

	return errcode.Success, nil
}

func (s *PaperService) GetByID(id uint64, teacherID uint64) (*dto.PaperResponse, int, error) {
	paper, err := s.paperRepo.FindByID(id)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限操作此试卷")
	}

	cnt, _ := s.pqRepo.GetMaxSortOrder(id)
	return toPaperResponse(paper, cnt), errcode.Success, nil
}

func (s *PaperService) List(query dto.PaperQuery, teacherID uint64) (*dto.PaperListResponse, int, error) {
	papers, total, err := s.paperRepo.List(query.Keyword, query.Status, query.Page, query.PageSize, teacherID)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	// 批量查询每份试卷的题目数量
	paperIDs := make([]uint64, len(papers))
	for i, p := range papers {
		paperIDs[i] = p.ID
	}
	cntMap, _ := s.paperRepo.GetQuestionCountByPaperIDs(paperIDs)

	items := make([]dto.PaperResponse, len(papers))
	for i, p := range papers {
		cnt := cntMap[p.ID]
		items[i] = *toPaperResponse(&p, cnt)
	}

	return &dto.PaperListResponse{
		Total: total,
		Items: items,
	}, errcode.Success, nil
}

func (s *PaperService) AddQuestions(id uint64, req dto.AddQuestionsRequest, teacherID uint64) (*dto.PaperResponse, int, error) {
	paper, err := s.paperRepo.FindByID(id)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限操作此试卷")
	}
	if paper.Status != model.PaperStatusDraft {
		return nil, errcode.PaperNotDraft, errors.New("仅草稿状态的试卷可添加题目")
	}

	maxSort, _ := s.pqRepo.GetMaxSortOrder(id)
	items := make([]*model.PaperQuestion, len(req.Questions))
	for i, item := range req.Questions {
		// 校验题目是否存在
		q, err := s.questionRepo.FindByID(item.QuestionID)
		if err != nil {
			return nil, errcode.QuestionNotFound, fmt.Errorf("题目 %d 不存在", item.QuestionID)
		}
		if q.CreatorID != teacherID {
			return nil, errcode.NoPermission, fmt.Errorf("题目 %d 不属于当前教师", item.QuestionID)
		}

		items[i] = &model.PaperQuestion{
			PaperID:    id,
			QuestionID: item.QuestionID,
			SortOrder:  maxSort + i + 1,
			Score:      item.Score,
		}
	}

	if err := s.pqRepo.BatchCreate(items); err != nil {
		return nil, errcode.UnknownError, err
	}

	// 更新总分 = 原有总分 + 新增题目总分
	actualTotal, _ := s.pqRepo.CalcTotalScore(id)
	if paper.TotalScore != actualTotal {
		paper.TotalScore = actualTotal
		_ = s.paperRepo.Update(paper)
	}

	cnt, _ := s.pqRepo.GetMaxSortOrder(id)
	return toPaperResponse(paper, cnt), errcode.Success, nil
}

func (s *PaperService) RandomSelect(id uint64, req dto.RandomSelectRequest, teacherID uint64) (*dto.PaperResponse, int, error) {
	paper, err := s.paperRepo.FindByID(id)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限操作此试卷")
	}
	if paper.Status != model.PaperStatusDraft {
		return nil, errcode.PaperNotDraft, errors.New("仅草稿状态的试卷可随机组卷")
	}

	// 收集所有规则需要的题目
	var allItems []*model.PaperQuestion
	var totalScore int
	sortOrder := 0

	for _, rule := range req.Rules {
		questions, err := s.questionRepo.FindByTypeAndCreator(rule.Type, teacherID)
		if err != nil {
			return nil, errcode.UnknownError, err
		}

		if len(questions) < rule.Count {
			return nil, errcode.InsufficientQuestions,
				fmt.Errorf("题型 %d 符合条件题目 %d 道，需要 %d 道",
					rule.Type, len(questions), rule.Count)
		}

		// Fisher-Yates 洗牌取前 count 个
		rand.Shuffle(len(questions), func(i, j int) {
			questions[i], questions[j] = questions[j], questions[i]
		})

		selected := questions[:rule.Count]
		perScore := rule.TotalScore / rule.Count
		remainder := rule.TotalScore - perScore*rule.Count
		for i, q := range selected {
			score := perScore
			if i < remainder {
				score++
			}
			sortOrder++
			allItems = append(allItems, &model.PaperQuestion{
				PaperID:    id,
				QuestionID: q.ID,
				SortOrder:  sortOrder,
				Score:      score,
			})
		}
		totalScore += rule.TotalScore
	}

	// 清空旧题目，写入新题目
	_ = s.pqRepo.ClearByPaperID(id)
	if err := s.pqRepo.BatchCreate(allItems); err != nil {
		return nil, errcode.UnknownError, err
	}

	// 更新总分
	paper.TotalScore = totalScore
	_ = s.paperRepo.Update(paper)

	cnt, _ := s.pqRepo.GetMaxSortOrder(id)
	return toPaperResponse(paper, cnt), errcode.Success, nil
}

func (s *PaperService) Publish(id uint64, teacherID uint64) (*dto.PaperResponse, int, error) {
	paper, err := s.paperRepo.FindByID(id)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限操作此试卷")
	}
	if paper.Status != model.PaperStatusDraft {
		return nil, errcode.PaperNotDraft, errors.New("仅草稿状态的试卷可发布")
	}

	// 校验是否有题目
	qCnt, _ := s.pqRepo.GetMaxSortOrder(id)
	if qCnt == 0 {
		return nil, errcode.PaperNoQuestions, errors.New("试卷中无题目，无法发布")
	}

	// 校验总分一致性
	actualTotal, _ := s.pqRepo.CalcTotalScore(id)
	if actualTotal != paper.TotalScore {
		return nil, errcode.PaperScoreMismatch,
			fmt.Errorf("题目总分 %d 与试卷总分 %d 不一致", actualTotal, paper.TotalScore)
	}

	paper.Status = model.PaperStatusPublished
	if err := s.paperRepo.Update(paper); err != nil {
		return nil, errcode.UnknownError, err
	}

	return toPaperResponse(paper, qCnt), errcode.Success, nil
}

func (s *PaperService) Unpublish(id uint64, teacherID uint64) (*dto.PaperResponse, int, error) {
	paper, err := s.paperRepo.FindByID(id)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限操作此试卷")
	}
	if paper.Status != model.PaperStatusPublished {
		return nil, errcode.PaperCannotModify, errors.New("仅已发布的试卷可撤销")
	}

	paper.Status = model.PaperStatusDraft
	if err := s.paperRepo.Update(paper); err != nil {
		return nil, errcode.UnknownError, err
	}

	qCnt, _ := s.pqRepo.GetMaxSortOrder(id)
	return toPaperResponse(paper, qCnt), errcode.Success, nil
}

func (s *PaperService) Copy(id uint64, teacherID uint64) (*dto.PaperResponse, int, error) {
	original, err := s.paperRepo.FindByID(id)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if original.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限操作此试卷")
	}

	newPaper := &model.Paper{
		Title:       original.Title + "（副本）",
		Description: original.Description,
		Duration:    original.Duration,
		TotalScore:  original.TotalScore,
		StartTime:   original.StartTime,
		EndTime:     original.EndTime,
		Status:      model.PaperStatusDraft,
		CreatorID:   teacherID,
	}
	if err := s.paperRepo.Create(newPaper); err != nil {
		return nil, errcode.UnknownError, err
	}

	// 复制题目关联
	questions, err := s.pqRepo.GetQuestions(id)
	if err != nil {
		return nil, errcode.UnknownError, err
	}
	if len(questions) > 0 {
		items := make([]*model.PaperQuestion, len(questions))
		for i, q := range questions {
			items[i] = &model.PaperQuestion{
				PaperID:    newPaper.ID,
				QuestionID: q.QuestionID,
				SortOrder:  q.SortOrder,
				Score:      q.Score,
			}
		}
		if err := s.pqRepo.BatchCreate(items); err != nil {
			return nil, errcode.UnknownError, err
		}
	}

	qCnt, _ := s.pqRepo.GetMaxSortOrder(newPaper.ID)
	return toPaperResponse(newPaper, qCnt), errcode.Success, nil
}

func (s *PaperService) Preview(id uint64, teacherID uint64) (*dto.PaperPreviewResponse, int, error) {
	paper, err := s.paperRepo.FindByID(id)
	if err != nil {
		return nil, errcode.PaperNotFound, errors.New("试卷不存在")
	}
	if paper.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("无权限操作此试卷")
	}

	questions, err := s.pqRepo.GetQuestionsWithDetail(id)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	qCnt := len(questions)
	paperResp := toPaperResponse(paper, qCnt)
	qItems := make([]dto.PaperQuestionItem, len(questions))
	for i, q := range questions {
		qItems[i] = dto.PaperQuestionItem{
			ID:              q.ID,
			PaperID:         q.PaperID,
			QuestionID:      q.QuestionID,
			SortOrder:       q.SortOrder,
			Score:           q.Score,
			QuestionType:    q.Question.Type,
			QuestionContent: q.Question.Content,
			QuestionOptions: q.Question.Options,
		}
	}

	return &dto.PaperPreviewResponse{
		Paper:     *paperResp,
		Questions: qItems,
	}, errcode.Success, nil
}

func toPaperResponse(p *model.Paper, questionCnt int) *dto.PaperResponse {
	return &dto.PaperResponse{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Duration:    p.Duration,
		TotalScore:  p.TotalScore,
		StartTime:   p.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:     p.EndTime.Format("2006-01-02 15:04:05"),
		Status:      p.Status,
		CreatorID:   p.CreatorID,
		QuestionCnt: questionCnt,
		CreatedAt:   p.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
