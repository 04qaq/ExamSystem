package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"exam-server/config"
	"exam-server/internal/cache"
	"exam-server/internal/database"
	"exam-server/internal/discovery"
	"exam-server/internal/router"
	"exam-server/scripts"
)

func main() {
	// 自动定位配置路径：优先取可执行文件所在目录
	configPath := "config/config.yaml"
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		cfg := filepath.Join(dir, "config/config.yaml")
		if _, err := os.Stat(cfg); err == nil {
			configPath = cfg
		}
	}
	if err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	database.Init(config.AppConfig.Database)
	database.AutoMigrate()

	// 初始化答案缓存
	cache.Init()

	// 创建种子数据
	scripts.SeedAdmin()

	// 启动 UDP 服务发现（局域网自动发现）
	discovery.Start(config.AppConfig.Server.Port)

	// 设置 Gin 模式
	ginMode := "debug"
	if config.AppConfig.Server.Mode == "release" {
		ginMode = "release"
	}
	os.Setenv("GIN_MODE", ginMode)

	// 注册路由
	r := router.Setup()

	// 启动 HTTP 服务
	srv := &http.Server{
		Addr:    ":" + config.AppConfig.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("服务启动于端口 %s", config.AppConfig.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭服务...")
	discovery.Stop()
	cache.GlobalCache.Stop()
	database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务关闭异常: %v", err)
	}
	log.Println("服务已关闭")
}
