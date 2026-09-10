// Package response 统一 JSON 响应格式，避免每个 handler 重复写 Content-Type。
package response

import (
	"encoding/json"
	"net/http"
)

// JSON 输出 JSON，并设置正确的 Content-Type。
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// OK 是 200 的快捷方式。
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, map[string]any{"ok": true, "data": data})
}

// Fail 是错误响应的统一出口。
func Fail(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]any{"ok": false, "error": message})
}
