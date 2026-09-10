package handler

import (
	"net/http"

	"vercel-go-demo/pkg/handlers"
)

// Handler 查询参数演示。实现在 internal/handlers，服务器模式复用同一份。
func Handler(w http.ResponseWriter, r *http.Request) { handlers.Hello(w, r) }
