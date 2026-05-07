package repository

import (
	"exam-server/internal/database"
	"exam-server/internal/model"
)

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

func (r *UserRepo) CreateUser(user *model.User) error {
	return database.DB.Create(user).Error
}

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(id uint64) (*model.User, error) {
	var user model.User
	err := database.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
