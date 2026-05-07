package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"

	"exam-desktop/lanfind"

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
		Windows: application.WindowsOptions{
			// 减轻 WebView2 对「页面请求回环地址」的限制，便于 axios/fetch 访问本机 exam-server
			AdditionalBrowserArgs: []string{
				"--disable-features=BlockInsecurePrivateNetworkRequests",
			},
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
			// 与页面同源提供 /discover，避免 WebView2 拦截对 127.0.0.1 随机端口的访问（本地环回限制）
			Middleware: func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/discover" {
						next.ServeHTTP(w, r)
						return
					}
					if r.Method == http.MethodOptions {
						w.WriteHeader(http.StatusNoContent)
						return
					}
					if r.Method != http.MethodGet {
						http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
						return
					}
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					u, err := lanfind.Discover(r.Context())
					if err != nil {
						log.Printf("[discover] %v", err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					if u == "" {
						log.Printf("[discover] finished: no exam-server found (UDP/TCP/HTTP sweep)")
					} else {
						log.Printf("[discover] ok url=%q", u)
					}
					_ = json.NewEncoder(w).Encode(map[string]string{"url": u})
				})
			},
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
