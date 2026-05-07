package scripts

import (
	"log"

	"exam-server/internal/database"
	"exam-server/internal/model"
	"exam-server/internal/utils"
)

func SeedAdmin() {
	seeds := []struct {
		Username string
		Password string
		Role     int8
		RealName string
	}{
		{"admin", "admin123", model.RoleAdmin, "系统管理员"},
		{"teacher", "123456", model.RoleTeacher, "张老师"},
		{"student", "student123", model.RoleStudent, "李同学"},
	}

	for _, s := range seeds {
		var count int64
		database.DB.Model(&model.User{}).Where("username = ?", s.Username).Count(&count)
		if count > 0 {
			log.Printf("账号 %s 已存在，跳过", s.Username)
			continue
		}

		hash, err := utils.HashPassword(s.Password)
		if err != nil {
			log.Fatalf("创建密码失败: %v", err)
		}

		user := model.User{
			Username:     s.Username,
			PasswordHash: hash,
			Role:         s.Role,
			RealName:     s.RealName,
		}
		if err := database.DB.Create(&user).Error; err != nil {
			log.Fatalf("创建账号 %s 失败: %v", s.Username, err)
		}
		log.Printf("默认账号已创建: %s / %s", s.Username, s.Password)
	}
}
