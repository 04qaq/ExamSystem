package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"exam-server/internal/dto"
	"exam-server/internal/model"
	"exam-server/internal/repository"
	"exam-server/pkg/errcode"
)

type QuestionService struct {
	questionRepo *repository.QuestionRepo
}

func NewQuestionService() *QuestionService {
	return &QuestionService{
		questionRepo: repository.NewQuestionRepo(),
	}
}

func (s *QuestionService) Create(req dto.CreateQuestionRequest, creatorID uint64) (*dto.QuestionResponse, int, error) {
	if err := validateQuestion(req.Type, req.Options); err != nil {
		return nil, errcode.InvalidParams, err
	}

	q := &model.Question{
		Type:       req.Type,
		Content:    req.Content,
		Options:    optionsToJSON(req.Options),
		Answer:     req.Answer,
		Score:      req.Score,
		Difficulty: req.Difficulty,
		Tags:       req.Tags,
		CreatorID:  creatorID,
	}

	if err := s.questionRepo.Create(q); err != nil {
		return nil, errcode.UnknownError, err
	}

	return toQuestionResponse(q), errcode.Success, nil
}

func (s *QuestionService) Update(id uint64, req dto.UpdateQuestionRequest, teacherID uint64) (*dto.QuestionResponse, int, error) {
	q, err := s.questionRepo.FindByID(id)
	if err != nil {
		return nil, errcode.QuestionNotFound, errors.New("题目不存在")
	}

	if q.CreatorID != teacherID {
		return nil, errcode.NoPermission, errors.New("只能编辑自己的题目")
	}

	if err := validateQuestion(req.Type, req.Options); err != nil {
		return nil, errcode.InvalidParams, err
	}

	q.Type = req.Type
	q.Content = req.Content
	q.Options = optionsToJSON(req.Options)
	q.Answer = req.Answer
	q.Score = req.Score
	q.Difficulty = req.Difficulty
	q.Tags = req.Tags

	if err := s.questionRepo.Update(q); err != nil {
		return nil, errcode.UnknownError, err
	}

	return toQuestionResponse(q), errcode.Success, nil
}

func (s *QuestionService) Delete(id uint64, teacherID uint64) (int, error) {
	q, err := s.questionRepo.FindByID(id)
	if err != nil {
		return errcode.QuestionNotFound, errors.New("题目不存在")
	}

	if q.CreatorID != teacherID {
		return errcode.NoPermission, errors.New("只能删除自己的题目")
	}

	if err := s.questionRepo.Delete(id); err != nil {
		return errcode.UnknownError, err
	}

	return errcode.Success, nil
}

func (s *QuestionService) GetByID(id uint64) (*dto.QuestionResponse, int, error) {
	q, err := s.questionRepo.FindByID(id)
	if err != nil {
		return nil, errcode.QuestionNotFound, errors.New("题目不存在")
	}

	return toQuestionResponse(q), errcode.Success, nil
}

func (s *QuestionService) List(query dto.QuestionQuery) (*dto.QuestionListResponse, int, error) {
	questions, total, err := s.questionRepo.List(
		query.Type, query.Difficulty, query.Tag, query.Keyword,
		query.Page, query.PageSize,
	)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	items := make([]dto.QuestionResponse, len(questions))
	for i, q := range questions {
		items[i] = *toQuestionResponse(&q)
	}

	return &dto.QuestionListResponse{
		Total: total,
		Items: items,
	}, errcode.Success, nil
}

func (s *QuestionService) Import(req dto.ImportRequest, creatorID uint64) (*dto.ImportResponse, int, error) {
	questions := make([]*model.Question, 0, len(req.Questions))
	var fail int

	for _, item := range req.Questions {
		if err := validateQuestion(item.Type, item.Options); err != nil {
			fail++
			continue
		}
		questions = append(questions, &model.Question{
			Type:       item.Type,
			Content:    item.Content,
			Options:    optionsToJSON(item.Options),
			Answer:     item.Answer,
			Score:      item.Score,
			Difficulty: item.Difficulty,
			Tags:       item.Tags,
			CreatorID:  creatorID,
		})
	}

	if len(questions) == 0 {
		return &dto.ImportResponse{
			Total:   len(req.Questions),
			Success: 0,
			Fail:    fail,
		}, errcode.Success, nil
	}

	if err := s.questionRepo.BatchCreate(questions); err != nil {
		return nil, errcode.UnknownError, err
	}

	return &dto.ImportResponse{
		Total:   len(req.Questions),
		Success: len(questions),
		Fail:    fail,
	}, errcode.Success, nil
}

func validateQuestion(qType int8, options string) error {
	if qType == model.QuestionTypeSingle || qType == model.QuestionTypeMultiple {
		if options == "" {
			return errors.New("选择题必须提供选项")
		}
	}
	return nil
}

func toQuestionResponse(q *model.Question) *dto.QuestionResponse {
	return &dto.QuestionResponse{
		ID:         q.ID,
		Type:       q.Type,
		Content:    q.Content,
		Options:    optionsFromJSON(q.Options),
		Answer:     q.Answer,
		Score:      q.Score,
		Difficulty: q.Difficulty,
		Tags:       q.Tags,
		CreatorID:  q.CreatorID,
		CreatedAt:  q.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  q.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// optionsToJSON 将前端传的多行文本选项转为 JSON 数组字符串存储
func optionsToJSON(text string) string {
	if text == "" {
		return "[]"
	}
	parts := strings.Split(strings.TrimSpace(text), "\n")
	opts := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			opts = append(opts, p)
		}
	}
	b, _ := json.Marshal(opts)
	return string(b)
}

// optionsFromJSON 将存储的 JSON 数组字符串转为多行文本供前端显示
func optionsFromJSON(jsonStr string) string {
	if jsonStr == "" || jsonStr == "[]" {
		return ""
	}
	var opts []string
	if err := json.Unmarshal([]byte(jsonStr), &opts); err != nil {
		return jsonStr
	}
	return strings.Join(opts, "\n")
}

// GetTypeLabel 返回题型中文名
func GetTypeLabel(qType int8) string {
	switch qType {
	case model.QuestionTypeSingle:
		return "单选题"
	case model.QuestionTypeMultiple:
		return "多选题"
	case model.QuestionTypeJudge:
		return "判断题"
	case model.QuestionTypeFill:
		return "填空题"
	case model.QuestionTypeShort:
		return "简答题"
	default:
		return fmt.Sprintf("未知(%d)", qType)
	}
}

// GetDifficultyLabel 返回难度中文名
func GetDifficultyLabel(diff int8) string {
	switch diff {
	case model.DifficultyEasy:
		return "简单"
	case model.DifficultyMedium:
		return "中等"
	case model.DifficultyHard:
		return "困难"
	default:
		return "未知"
	}
}
