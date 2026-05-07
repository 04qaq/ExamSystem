package service

import (
	"errors"

	"exam-server/config"
	"exam-server/internal/dto"
	"exam-server/internal/model"
	"exam-server/internal/repository"
	"exam-server/internal/utils"
	"exam-server/pkg/errcode"

	"gorm.io/gorm"
)

type AuthService struct {
	userRepo *repository.UserRepo
}

func NewAuthService() *AuthService {
	return &AuthService{userRepo: repository.NewUserRepo()}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.RegisterResponse, int, error) {
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, errcode.UnknownError, err
	}
	if exists {
		return nil, errcode.UserAlreadyExists, errors.New("用户名已存在")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Role:         model.RoleStudent,
		RealName:     req.RealName,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, errcode.UnknownError, err
	}

	return &dto.RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		RealName: user.RealName,
	}, errcode.Success, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.LoginResponse, int, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.UserNotFound, errors.New("用户不存在")
		}
		return nil, errcode.UnknownError, err
	}

	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		return nil, errcode.PasswordWrong, errors.New("密码错误")
	}

	if user.Status != 1 {
		return nil, errcode.UserDisabled, errors.New("用户已被禁用")
	}

	accessToken, err := utils.GenerateToken(user.ID, int(user.Role), user.Username, config.AppConfig.JWT.AccessExpire)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	refreshToken, err := utils.GenerateToken(user.ID, int(user.Role), user.Username, config.AppConfig.JWT.RefreshExpire)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(config.AppConfig.JWT.AccessExpire.Seconds()),
		User: &dto.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
			RealName: user.RealName,
		},
	}, errcode.Success, nil
}

func (s *AuthService) GetUserInfo(userID uint64) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	return &dto.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		RealName: user.RealName,
	}, nil
}