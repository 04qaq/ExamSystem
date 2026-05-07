package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// 由 scripts/sync-frontend.ps1 将 ../frontend/dist 同步到此目录后再执行 wails3 build / go build。
//
//go:embed all:embed/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "在线考试系统",
		Description: "在线考试系统桌面客户端（嵌入 Vue 前端，API 连接独立 exam-server）",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:             "在线考试系统",
		Width:             1280,
		Height:            800,
		MinWidth:          900,
		MinHeight:         600,
		URL:               "/",
		BackgroundColour:  application.NewRGB(255, 255, 255),
		InitialPosition:   application.WindowCentered,
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
