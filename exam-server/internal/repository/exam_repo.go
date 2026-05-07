package repository

import (
	"exam-server/internal/database"
	"exam-server/internal/model"
)

type ExamRepo struct{}

func NewExamRepo() *ExamRepo {
	return &ExamRepo{}
}

func (r *ExamRepo) CreateRecord(record *model.ExamRecord) error {
	return database.DB.Create(record).Error
}

func (r *ExamRepo) FindRecordByID(id uint64) (*model.ExamRecord, error) {
	var rec model.ExamRecord
	err := database.DB.First(&rec, id).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *ExamRepo) FindInProgressByUserAndPaper(userID, paperID uint64) (*model.ExamRecord, error) {
	var rec model.ExamRecord
	err := database.DB.
		Where("user_id = ? AND paper_id = ? AND status = ?", userID, paperID, model.ExamStatusInProgress).
		First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *ExamRepo) UpdateRecord(record *model.ExamRecord) error {
	return database.DB.Save(record).Error
}

func (r *ExamRepo) ListRecordsByUser(userID uint64, page, pageSize int) ([]model.ExamRecord, int64, error) {
	var records []model.ExamRecord
	var total int64

	query := database.DB.Model(&model.ExamRecord{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error
	return records, total, err
}

func (r *ExamRepo) GetAnswerDetails(recordID uint64) ([]model.AnswerDetail, error) {
	var details []model.AnswerDetail
	err := database.DB.
		Where("exam_record_id = ?", recordID).
		Order("id ASC").
		Find(&details).Error
	return details, err
}

func (r *ExamRepo) BatchSaveAnswers(details []*model.AnswerDetail) error {
	return database.DB.CreateInBatches(details, 100).Error
}

// UpsertAnswer 插入或更新答案（用于实时保存）
func (r *ExamRepo) UpsertAnswer(detail *model.AnswerDetail) error {
	return database.DB.
		Where("exam_record_id = ? AND question_id = ?", detail.ExamRecordID, detail.QuestionID).
		Assign(model.AnswerDetail{ProvidedAnswer: detail.ProvidedAnswer}).
		FirstOrCreate(detail).Error
}

// GetPapersWithStudentRecord 查询学生已参加的试卷ID列表
func (r *ExamRepo) GetSubmittedPaperIDsByUser(userID uint64) (map[uint64]uint64, error) {
	type Result struct {
		PaperID  uint64
		RecordID uint64
	}
	var results []Result
	err := database.DB.Model(&model.ExamRecord{}).
		Select("paper_id, id as record_id").
		Where("user_id = ? AND status = ?", userID, model.ExamStatusSubmitted).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	m := make(map[uint64]uint64, len(results))
	for _, r := range results {
		m[r.PaperID] = r.RecordID
	}
	return m, nil
}
