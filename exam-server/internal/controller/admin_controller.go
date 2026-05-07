package controller

import (
	"net/http"
	"strconv"

	"exam-server/internal/dto"
	"exam-server/internal/service"
	"exam-server/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	adminService *service.AdminService
}

func NewAdminController() *AdminController {
	return &AdminController{
		adminService: service.NewAdminService(),
	}
}

// ListUsers 用户列表
func (ctrl *AdminController) ListUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	resp, code, err := ctrl.adminService.ListUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}

// CreateUser 创建用户
func (ctrl *AdminController) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	code, err := ctrl.adminService.CreateUser(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	// 记录日志
	userID := c.GetUint64("userID")
	ctrl.adminService.CreateLog(userID, c.GetString("username"), "创建用户", req.Username, "", c.ClientIP())

	c.JSON(http.StatusOK, dto.Success(nil))
}

// UpdateUser 更新用户
func (ctrl *AdminController) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, "无效的用户ID"))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	code, err := ctrl.adminService.UpdateUser(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	// 记录日志
	adminID := c.GetUint64("userID")
	ctrl.adminService.CreateLog(adminID, c.GetString("username"), "更新用户", strconv.FormatUint(id, 10), "", c.ClientIP())

	c.JSON(http.StatusOK, dto.Success(nil))
}

// ListLogs 操作日志列表
func (ctrl *AdminController) ListLogs(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	action := c.Query("action")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	resp, code, err := ctrl.adminService.ListLogs(page, pageSize, action)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.Success(resp))
}
