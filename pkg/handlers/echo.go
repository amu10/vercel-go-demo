package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"vercel-go-demo/pkg/config"
	"vercel-go-demo/pkg/response"
)

type echoRequest struct {
	Message string   `json:"message"`
	Items   []string `json:"items"`
}

// Echo 演示 POST + JSON 请求体。
//
// 方法判断写在代码里：函数模式下 Vercel 只认 Handler 一个入口，
// 所有 HTTP 方法都会进来，不能靠 ServeMux 分发。
// 服务器模式下也可以顺手在 main.go 里写 "POST /api/echo"，但这里保持两种模式行为一致。
//
// POST /api/echo  {"message":"hi","items":["a","b"]}
func Echo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Fail(w, http.StatusMethodNotAllowed, "只支持 POST 方法")
		return
	}

	// 限制请求体大小，防止被超大 payload 打爆（Serverless 内存很贵）
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB
	if err != nil {
		response.Fail(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	if len(strings.TrimSpace(string(body))) == 0 {
		response.Fail(w, http.StatusBadRequest, "请求体不能为空")
		return
	}

	var req echoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		response.Fail(w, http.StatusBadRequest, "JSON 格式不合法: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		response.Fail(w, http.StatusBadRequest, "message 字段必填")
		return
	}

	// 演示环境变量 MAX_ITEMS 真的生效（Vercel 控制台配了就能改，不用重新编译）
	cfg, err := config.Load()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, "配置加载失败: "+err.Error())
		return
	}

	truncated := false
	if len(req.Items) > cfg.MaxItems {
		req.Items = req.Items[:cfg.MaxItems]
		truncated = true
	}

	response.OK(w, map[string]any{
		"echo":      req.Message,
		"length":    len(req.Message),
		"items":     req.Items,
		"truncated": truncated,
		"max_items": cfg.MaxItems,
		"received": map[string]any{
			"content_type": r.Header.Get("Content-Type"),
			"user_agent":   r.Header.Get("User-Agent"),
		},
	})
}
