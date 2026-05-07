package repository

import (
	"exam-server/internal/database"
	"exam-server/internal/model"
)

type PaperRepo struct{}

func NewPaperRepo() *PaperRepo {
	return &PaperRepo{}
}

func (r *PaperRepo) Create(paper *model.Paper) error {
	return database.DB.Create(paper).Error
}

func (r *PaperRepo) Update(paper *model.Paper) error {
	return database.DB.Save(paper).Error
}

func (r *PaperRepo) Delete(id uint64) error {
	return database.DB.Delete(&model.Paper{}, id).Error
}

func (r *PaperRepo) FindByID(id uint64) (*model.Paper, error) {
	var p model.Paper
	err := database.DB.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaperRepo) List(keyword string, status int8, page, pageSize int, creatorID uint64) ([]model.Paper, int64, error) {
	var papers []model.Paper
	var total int64

	query := database.DB.Model(&model.Paper{}).Where("creator_id = ?", creatorID)

	if status > 0 {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("title LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&papers).Error
	if err != nil {
		return nil, 0, err
	}

	return papers, total, nil
}

// GetQuestionCountByPaperIDs 批量查询每份试卷的题目数量
func (r *PaperRepo) GetQuestionCountByPaperIDs(paperIDs []uint64) (map[uint64]int, error) {
	type Result struct {
		PaperID uint64
		Count   int
	}
	var results []Result
	err := database.DB.Model(&model.PaperQuestion{}).
		Select("paper_id, COUNT(*) as count").
		Where("paper_id IN ?", paperIDs).
		Group("paper_id").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[uint64]int, len(results))
	for _, r := range results {
		counts[r.PaperID] = r.Count
	}
	return counts, nil
}
