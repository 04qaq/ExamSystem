package middleware

import (
	"net/http"

	"exam-server/internal/dto"
	"exam-server/internal/model"
	"exam-server/pkg/errcode"

	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...int) gin.HandlerFunc {
	roleMap := make(map[int]bool, len(roles))
	for _, r := range roles {
		roleMap[r] = true
	}

	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, dto.Error(errcode.NoPermission, "未授权"))
			c.Abort()
			return
		}

		r, ok := role.(int)
		if !ok || !roleMap[r] {
			c.JSON(http.StatusForbidden, dto.Error(errcode.NoPermission, "无权限访问"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleAdmin)
}

func RequireTeacher() gin.HandlerFunc {
	return RequireRole(model.RoleTeacher)
}

func RequireStudent() gin.HandlerFunc {
	return RequireRole(model.RoleStudent)
}
