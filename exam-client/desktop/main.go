package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"

	"exam-desktop/lanfind"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// 由 scripts/sync-frontend.ps1 同步 ../frontend/dist 后再构建。
// Windows 发布建议附带：-ldflags="-s -w -H windowsgui"（无控制台窗口）。
//
//go:embed all:embed/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "在线考试系统",
		Description: "在线考试系统桌面客户端",
		Windows: application.WindowsOptions{
			// 减轻 WebView2 对页面访问本机地址的限制，便于客户端连接本机考试服务
			AdditionalBrowserArgs: []string{
				"--disable-features=BlockInsecurePrivateNetworkRequests",
			},
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
			// 与页面同源提供 /discover，避免 WebView2 拦截本地发现请求
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
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
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
