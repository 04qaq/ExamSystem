package repository

import (
	"exam-server/internal/database"
	"exam-server/internal/model"
)

type QuestionRepo struct{}

func NewQuestionRepo() *QuestionRepo {
	return &QuestionRepo{}
}

func (r *QuestionRepo) Create(question *model.Question) error {
	return database.DB.Create(question).Error
}

func (r *QuestionRepo) Update(question *model.Question) error {
	return database.DB.Save(question).Error
}

func (r *QuestionRepo) Delete(id uint64) error {
	return database.DB.Model(&model.Question{}).
		Where("id = ?", id).
		Update("is_deleted", 1).Error
}

func (r *QuestionRepo) FindByID(id uint64) (*model.Question, error) {
	var q model.Question
	err := database.DB.Where("is_deleted = 0").First(&q, id).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QuestionRepo) List(typeVal, difficulty int8, tag, keyword string, page, pageSize int) ([]model.Question, int64, error) {
	var questions []model.Question
	var total int64

	query := database.DB.Model(&model.Question{}).Where("is_deleted = 0")

	if typeVal > 0 {
		query = query.Where("type = ?", typeVal)
	}
	if difficulty > 0 {
		query = query.Where("difficulty = ?", difficulty)
	}
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}
	if keyword != "" {
		query = query.Where("content LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&questions).Error
	if err != nil {
		return nil, 0, err
	}

	return questions, total, nil
}

// FindByTypeAndCreator 按题型和创建者查询所有题目（随机组卷用）
func (r *QuestionRepo) FindByTypeAndCreator(typeVal int8, creatorID uint64) ([]model.Question, error) {
	var questions []model.Question
	err := database.DB.Model(&model.Question{}).
		Where("type = ? AND creator_id = ? AND is_deleted = 0", typeVal, creatorID).
		Find(&questions).Error
	return questions, err
}

func (r *QuestionRepo) BatchCreate(questions []*model.Question) error {
	return database.DB.CreateInBatches(questions, 100).Error
}
