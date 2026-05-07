package repository

import (
	"exam-server/internal/database"
	"exam-server/internal/model"
)

type AdminRepo struct{}

func NewAdminRepo() *AdminRepo {
	return &AdminRepo{}
}

// 用户管理
func (r *AdminRepo) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := database.DB.Model(&model.User{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("id ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error
	return users, total, err
}

func (r *AdminRepo) GetUserByID(id uint64) (*model.User, error) {
	var user model.User
	err := database.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AdminRepo) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AdminRepo) CreateUser(user *model.User) error {
	return database.DB.Create(user).Error
}

func (r *AdminRepo) UpdateUser(id uint64, updates map[string]interface{}) error {
	return database.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// 操作日志
func (r *AdminRepo) CreateLog(log *model.OperationLog) error {
	return database.DB.Create(log).Error
}

func (r *AdminRepo) ListLogs(page, pageSize int, action string) ([]model.OperationLog, int64, error) {
	var logs []model.OperationLog
	var total int64

	query := database.DB.Model(&model.OperationLog{})
	if action != "" {
		query = query.Where("action LIKE ?", "%"+action+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error
	return logs, total, err
}
