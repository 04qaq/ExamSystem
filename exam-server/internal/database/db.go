package database

import (
	"fmt"
	"log"

	"exam-server/config"
	"exam-server/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg config.DatabaseConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	var err error
	logLevel := logger.Info
	if config.AppConfig.Server.Mode == "release" {
		logLevel = logger.Warn
	}

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("数据库 Ping 失败: %v", err)
	}

	log.Println("数据库连接成功")
}

func AutoMigrate() {
	if err := DB.AutoMigrate(
		&model.User{},
		&model.Question{},
		&model.Paper{},
		&model.PaperQuestion{},
		&model.ExamRecord{},
		&model.AnswerDetail{},
		&model.OperationLog{},
	); err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	log.Println("数据库表迁移完成")
}

func Close() {
	sqlDB, err := DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}
