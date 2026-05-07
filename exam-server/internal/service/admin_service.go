package service

import (
	"errors"
	"net"

	"exam-server/internal/dto"
	"exam-server/internal/model"
	"exam-server/internal/repository"
	"exam-server/internal/utils"
	"exam-server/pkg/errcode"
)

type AdminService struct {
	adminRepo *repository.AdminRepo
}

func NewAdminService() *AdminService {
	return &AdminService{
		adminRepo: repository.NewAdminRepo(),
	}
}

// 用户管理
func (s *AdminService) ListUsers(page, pageSize int) (*dto.UserListResponse, int, error) {
	users, total, err := s.adminRepo.ListUsers(page, pageSize)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	items := make([]dto.UserListItem, len(users))
	for i, u := range users {
		items[i] = dto.UserListItem{
			ID:        u.ID,
			Username:  u.Username,
			Role:      u.Role,
			RealName:  u.RealName,
			Status:    u.Status,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return &dto.UserListResponse{Total: total, Items: items}, errcode.Success, nil
}

func (s *AdminService) CreateUser(req dto.CreateUserRequest) (int, error) {
	// 检查用户名是否已存在
	existing, _ := s.adminRepo.GetUserByUsername(req.Username)
	if existing != nil {
		return errcode.UserAlreadyExists, errors.New("用户名已存在")
	}

	if len(req.Password) < 6 {
		return errcode.PasswordTooShort, errors.New("密码长度至少6位")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return errcode.UserCreateFailed, err
	}

	user := model.User{
		Username:     req.Username,
		PasswordHash: hash,
		Role:         req.Role,
		RealName:     req.RealName,
		Status:       model.UserStatusActive,
	}
	if err := s.adminRepo.CreateUser(&user); err != nil {
		return errcode.UserCreateFailed, err
	}
	return errcode.Success, nil
}

func (s *AdminService) UpdateUser(id uint64, req dto.UpdateUserRequest) (int, error) {
	user, err := s.adminRepo.GetUserByID(id)
	if err != nil {
		return errcode.UserNotFound, errors.New("用户不存在")
	}

	updates := make(map[string]interface{})
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.Status != nil {
		if *req.Status != model.UserStatusActive && *req.Status != model.UserStatusDisabled {
			return errcode.InvalidParams, errors.New("状态值无效")
		}
		updates["status"] = *req.Status
	}
	if req.Password != "" {
		if len(req.Password) < 6 {
			return errcode.PasswordTooShort, errors.New("密码长度至少6位")
		}
		hash, err := utils.HashPassword(req.Password)
		if err != nil {
			return errcode.UserUpdateFailed, err
		}
		updates["password_hash"] = hash
	}

	if len(updates) == 0 {
		return errcode.Success, nil
	}

	if err := s.adminRepo.UpdateUser(user.ID, updates); err != nil {
		return errcode.UserUpdateFailed, err
	}
	return errcode.Success, nil
}

// 操作日志
func (s *AdminService) ListLogs(page, pageSize int, action string) (*dto.LogListResponse, int, error) {
	logs, total, err := s.adminRepo.ListLogs(page, pageSize, action)
	if err != nil {
		return nil, errcode.UnknownError, err
	}

	items := make([]dto.OperationLogItem, len(logs))
	for i, l := range logs {
		items[i] = dto.OperationLogItem{
			ID:        l.ID,
			UserID:    l.UserID,
			Username:  l.Username,
			Action:    l.Action,
			Target:    l.Target,
			Detail:    l.Detail,
			IP:        l.IP,
			CreatedAt: l.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return &dto.LogListResponse{Total: total, Items: items}, errcode.Success, nil
}

func (s *AdminService) CreateLog(userID uint64, username, action, target, detail, ip string) {
	// 清理 IP 中的端口号（支持 IPv6）
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	log := model.OperationLog{
		UserID:   userID,
		Username: username,
		Action:   action,
		Target:   target,
		Detail:   detail,
		IP:       ip,
	}
	_ = s.adminRepo.CreateLog(&log)
}
