package repository

import (
	"exam-server/internal/database"
	"exam-server/internal/model"
)

type GradeRepo struct{}

func NewGradeRepo() *GradeRepo {
	return &GradeRepo{}
}

func (r *GradeRepo) GetPendingRecords(teacherID uint64, page, pageSize int) ([]model.ExamRecord, int64, error) {
	var records []model.ExamRecord
	var total int64

	subQuery := database.DB.Model(&model.Paper{}).
		Select("id").
		Where("creator_id = ?", teacherID)

	query := database.DB.Model(&model.ExamRecord{}).
		Where("paper_id IN (?)", subQuery).
		Where("status IN (?, ?)", model.ExamStatusSubmitted, model.ExamStatusGraded)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error
	return records, total, err
}

func (r *GradeRepo) GetSubjectiveDetails(recordID uint64) ([]model.AnswerDetail, error) {
	var details []model.AnswerDetail
	err := database.DB.
		Where("exam_record_id = ?", recordID).
		Find(&details).Error
	return details, err
}

func (r *GradeRepo) GetAllDetailsByRecord(recordID uint64) ([]model.AnswerDetail, error) {
	var details []model.AnswerDetail
	err := database.DB.
		Where("exam_record_id = ?", recordID).
		Order("id ASC").
		Find(&details).Error
	return details, err
}

func (r *GradeRepo) UpdateGrade(detailID uint64, scoreGained int, comment string) error {
	return database.DB.Model(&model.AnswerDetail{}).
		Where("id = ?", detailID).
		Updates(map[string]interface{}{
			"score_gained": scoreGained,
			"comment":      comment,
		}).Error
}

func (r *GradeRepo) UpdateRecordStatus(recordID uint64, status int8, totalScore int) error {
	return database.DB.Model(&model.ExamRecord{}).
		Where("id = ?", recordID).
		Updates(map[string]interface{}{
			"status":      status,
			"total_score": totalScore,
		}).Error
}

func (r *GradeRepo) GetRecordsByPaperID(paperID uint64) ([]model.ExamRecord, error) {
	var records []model.ExamRecord
	err := database.DB.
		Where("paper_id = ? AND status IN (?, ?)", paperID, model.ExamStatusSubmitted, model.ExamStatusGraded).
		Find(&records).Error
	return records, err
}
