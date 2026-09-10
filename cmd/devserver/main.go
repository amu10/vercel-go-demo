// 本地调试用的开发服务器。
//
// Vercel 在生产环境会逐个把 api/**/*.go 编译成独立的 Serverless 函数，
// 本地没法直接跑。这个程序用标准 net/http 把同样的 Handler 挂起来，
// 让你在本地就能 curl 调试。
//
// 注意：放在 cmd/ 下，不在 api/ 下，所以不会被 Vercel 编译成函数。
//
// 用法：go run ./cmd/devserver  然后访问 http://localhost:8080
package main

import (
	"log"
	"net/http"

	rootapi "vercel-go-demo/api"
	cfgh "vercel-go-demo/api/config"
	echoh "vercel-go-demo/api/echo"
	healthh "vercel-go-demo/api/health"
	helloh "vercel-go-demo/api/hello"
)

func main() {
	mux := http.NewServeMux()

	// 和 Vercel 的路由一一对应
	mux.HandleFunc("/api", rootapi.Handler)
	mux.HandleFunc("/api/health", healthh.Handler)
	mux.HandleFunc("/api/hello", helloh.Handler)
	mux.HandleFunc("/api/echo", echoh.Handler)
	mux.HandleFunc("/api/config", cfgh.Handler)

	// 静态页面
	mux.Handle("/", http.FileServer(http.Dir("public")))

	log.Println("dev server → http://localhost:8080")
	log.Println("试着改环境变量：MAX_ITEMS=2 APP_ENV=production go run ./cmd/devserver")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
