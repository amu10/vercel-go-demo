// Package handlers 存放全部业务 handler，两种部署模式共用这一份实现：
//
//   - 函数模式（framework: null）：api/<name>/index.go 薄封装一层，导出 Handler
//   - 服务器模式（framework: "go"）：cmd/api/main.go 用 ServeMux 统一注册
//
// 好处是切换模式时业务代码零改动，只动 vercel.json。
package handlers

import (
	"net/http"

	"vercel-go-demo/pkg/response"
)

type Route struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Desc    string `json:"desc"`
	Example string `json:"example"`
}

// Index 返回路由清单。
func Index(w http.ResponseWriter, r *http.Request) {
	routes := []Route{
		{"GET", "/api", "路由清单（本接口）", "/api"},
		{"GET", "/api/health", "健康检查 + 运行环境信息", "/api/health"},
		{"GET", "/api/hello", "演示查询参数", "/api/hello?name=张三&lang=zh"},
		{"POST", "/api/echo", "演示 JSON 请求体 + 环境变量生效", "/api/echo"},
		{"GET", "/api/config", "演示 go:embed 内嵌配置", "/api/config"},
	}

	mode := "函数模式（api 目录，framework: null）"
	if r.Header.Get("X-Vercel-Deployment-Id") == "" {
		mode = "本地 devserver / 服务器模式"
	}

	response.OK(w, map[string]any{
		"message": "Vercel Go runtime 示例项目",
		"mode":    mode,
		"routes":  routes,
	})
}
