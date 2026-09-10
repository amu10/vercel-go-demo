// Vercel 函数模式约定：
//  1. 文件放在 api/ 目录下
//  2. 导出 func Handler(w http.ResponseWriter, r *http.Request)
//  3. 路径即路由：api/index.go → /api，api/hello/index.go → /api/hello
//
// 业务实现已经抽到 internal/handlers，这里只做一层转发 ——
// 服务器模式（cmd/api/main.go）复用同一份实现，切换部署模式时业务代码零改动。
//
// 注意这里用「一个端点一个子目录」而不是 api/hello.go 平铺。
// 平铺的话 Vercel 能跑（它逐个文件单独编译），但本地 go build ./... 会报
// "Handler redeclared"，因为 Go 把整个 api/ 目录当成一个包。
// 拆成子目录后两边都干净。
package handler

import (
	"net/http"

	"vercel-go-demo/pkg/handlers"
)

// Handler 返回路由清单。
func Handler(w http.ResponseWriter, r *http.Request) { handlers.Index(w, r) }
