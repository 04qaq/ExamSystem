package repository

import (
	"exam-server/internal/database"
	"exam-server/internal/model"
)

type PaperQuestionRepo struct{}

func NewPaperQuestionRepo() *PaperQuestionRepo {
	return &PaperQuestionRepo{}
}

func (r *PaperQuestionRepo) BatchCreate(items []*model.PaperQuestion) error {
	return database.DB.CreateInBatches(items, 100).Error
}

func (r *PaperQuestionRepo) GetQuestions(paperID uint64) ([]model.PaperQuestion, error) {
	var items []model.PaperQuestion
	err := database.DB.
		Where("paper_id = ?", paperID).
		Order("sort_order ASC").
		Find(&items).Error
	return items, err
}

// GetQuestionsWithDetail 获取试卷题目并关联题目详情
func (r *PaperQuestionRepo) GetQuestionsWithDetail(paperID uint64) ([]model.PaperQuestion, error) {
	var items []model.PaperQuestion
	err := database.DB.
		Preload("Question").
		Where("paper_id = ?", paperID).
		Order("sort_order ASC").
		Find(&items).Error
	return items, err
}

func (r *PaperQuestionRepo) ClearByPaperID(paperID uint64) error {
	return database.DB.
		Where("paper_id = ?", paperID).
		Delete(&model.PaperQuestion{}).Error
}

func (r *PaperQuestionRepo) RemoveQuestions(paperID uint64, questionIDs []uint64) error {
	return database.DB.
		Where("paper_id = ? AND question_id IN ?", paperID, questionIDs).
		Delete(&model.PaperQuestion{}).Error
}

func (r *PaperQuestionRepo) GetQuestionIDs(paperID uint64) ([]uint64, error) {
	var ids []uint64
	err := database.DB.Model(&model.PaperQuestion{}).
		Where("paper_id = ?", paperID).
		Pluck("question_id", &ids).Error
	return ids, err
}

// GetMaxSortOrder 获取当前试卷最大排序值
func (r *PaperQuestionRepo) GetMaxSortOrder(paperID uint64) (int, error) {
	type Result struct {
		MaxSort int
	}
	var res Result
	err := database.DB.Model(&model.PaperQuestion{}).
		Select("COALESCE(MAX(sort_order), 0) as max_sort").
		Where("paper_id = ?", paperID).
		Scan(&res).Error
	return res.MaxSort, err
}

// CalcTotalScore 计算试卷题目分值总和
func (r *PaperQuestionRepo) CalcTotalScore(paperID uint64) (int, error) {
	type Result struct {
		Total int
	}
	var res Result
	err := database.DB.Model(&model.PaperQuestion{}).
		Select("COALESCE(SUM(score), 0) as total").
		Where("paper_id = ?", paperID).
		Scan(&res).Error
	return res.Total, err
}
