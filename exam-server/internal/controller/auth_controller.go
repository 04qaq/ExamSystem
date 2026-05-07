package controller

import (
	"net/http"

	"exam-server/internal/dto"
	"exam-server/internal/service"
	"exam-server/pkg/errcode"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController() *AuthController {
	return &AuthController{authService: service.NewAuthService()}
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	resp, code, err := ctrl.authService.Register(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(errcode.InvalidParams, err.Error()))
		return
	}

	resp, code, err := ctrl.authService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.Error(code, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(resp))
}
