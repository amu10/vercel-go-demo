// Vercel Go Framework Preset（服务器模式）的入口。
//
// 和函数模式（api/**/*.go）的区别：
//   - 整个项目编译成「一个」二进制，所有路由由这里的 ServeMux 分发
//   - 必须自己 ListenAndServe，端口从 PORT 环境变量读（Vercel 注入）
//   - public/ 不再被 Vercel 托管，首页要用 go:embed 内嵌（见根包 site）
//
// 业务 handler 与函数模式共用 internal/handlers，切换模式时业务代码零改动。
//
// 对应 vercel.json：{"framework": "go"}，且必须删掉 functions 段。
package main

import (
	"log"
	"net/http"
	"os"

	site "vercel-go-demo"
	"vercel-go-demo/pkg/handlers"
)

func main() {
	mux := http.NewServeMux()

	// 与函数模式的路由一一对应（函数模式靠目录路径，这里靠代码）
	mux.HandleFunc("/api", handlers.Index)
	mux.HandleFunc("/api/health", handlers.Health)
	mux.HandleFunc("/api/hello", handlers.Hello)
	mux.HandleFunc("/api/echo", handlers.Echo)
	mux.HandleFunc("/api/config", handlers.Config)

	// 服务器模式下连首页在内的全部流量都进这个二进制
	mux.HandleFunc("/", site.Handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("go server (framework preset) listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
