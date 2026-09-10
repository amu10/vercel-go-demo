// Vercel Go runtime 约定：
//  1. 文件放在 api/ 目录下
//  2. 导出 func Handler(w http.ResponseWriter, r *http.Request)
//  3. 路径即路由：api/index.go → /api，api/hello/index.go → /api/hello
//
// 注意这里用「一个端点一个子目录」而不是 api/hello.go 平铺。
// 平铺的话 Vercel 能跑（它逐个文件单独编译），但本地 go build ./... 会报
// "Handler redeclared"，因为 Go 把整个 api/ 目录当成一个包。
// 拆成子目录后两边都干净。
package handler

import (
	"net/http"

	"vercel-go-demo/internal/response"
)

type route struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Desc    string `json:"desc"`
	Example string `json:"example"`
}

// Handler 返回本项目的路由清单，方便部署后一眼看清有哪些接口。
func Handler(w http.ResponseWriter, r *http.Request) {
	routes := []route{
		{"GET", "/api", "路由清单（本接口）", "/api"},
		{"GET", "/api/health", "健康检查 + 运行环境信息", "/api/health"},
		{"GET", "/api/hello", "演示查询参数", "/api/hello?name=张三&lang=zh"},
		{"POST", "/api/echo", "演示 JSON 请求体 + 环境变量生效", "/api/echo"},
		{"GET", "/api/config", "演示 go:embed 内嵌配置", "/api/config"},
	}

	response.OK(w, map[string]any{
		"message": "Vercel Go runtime 示例项目",
		"routes":  routes,
	})
}
