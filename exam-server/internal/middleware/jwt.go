package middleware

import (
	"net/http"
	"strings"

	"exam-server/internal/dto"
	"exam-server/internal/utils"
	"exam-server/pkg/errcode"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.Error(errcode.TokenInvalid, "缺少Token"))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.Error(errcode.TokenInvalid, "Token格式错误"))
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.Error(errcode.TokenExpired, "Token无效或已过期"))
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("username", claims.Username)
		c.Next()
	}
}
