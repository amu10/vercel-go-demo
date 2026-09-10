// Package site 内嵌静态资源，供服务器模式托管首页。
//
// 放在模块根目录是刻意的：//go:embed 只能引用当前包目录及其子目录，
// 不能写 ../public，而 public/ 就在项目根 —— 只有根包能 embed 到它。
//
// 为什么需要内嵌？
//   - 函数模式：Vercel 自动静态托管 public/，首页直接可访问
//   - 服务器模式：所有流量都进你那个二进制，public/ 不再被托管，
//     而且文件系统只读、CWD 也不一定是项目根，运行时读文件必翻车
//
// 所以服务器模式只能靠 go:embed 把首页打进二进制。
package site

import (
	"embed"
	"net/http"
)

//go:embed public/index.html
var Public embed.FS

// IndexHTML 返回内嵌的首页内容。
func IndexHTML() ([]byte, error) {
	return Public.ReadFile("public/index.html")
}

// Handler 首页，可直接挂到 ServeMux 上。
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}

	html, err := IndexHTML()
	if err != nil {
		http.Error(w, "首页未内嵌: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(html)
}
