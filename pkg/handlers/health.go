package handlers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"vercel-go-demo/pkg/response"
)

// Health 健康检查，顺带把运行环境打印出来。
//
// 在 Vercel 上 hostname 每次可能不同 —— 这就是 Serverless：
// 实例随时被创建和回收，所以不要在这里存任何状态。
func Health(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	response.OK(w, map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"runtime": map[string]any{
			"go_version": runtime.Version(),
			"goos":       runtime.GOOS,
			"goarch":     runtime.GOARCH,
			"goroutines": runtime.NumGoroutine(),
			"heap_mb":    mem.HeapAlloc / 1024 / 1024,
		},
		"instance": map[string]any{
			"hostname": hostname,
			"region":   os.Getenv("VERCEL_REGION"),
			"env":      os.Getenv("VERCEL_ENV"),
		},
		"note": "Vercel 文件系统只读，只有 /tmp 可写且随实例销毁",
	})
}
