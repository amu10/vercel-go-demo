package handler

import (
	"net/http"
	"strconv"
	"strings"

	"vercel-go-demo/internal/response"
)

var greetings = map[string]string{
	"zh": "你好",
	"en": "Hello",
	"ja": "こんにちは",
}

// Handler 演示查询参数的处理套路：取默认值、白名单校验、非法值兜底。
//
// GET /api/hello?name=张三&lang=zh&count=3
func Handler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	name := strings.TrimSpace(q.Get("name"))
	if name == "" {
		name = "World"
	}

	lang := q.Get("lang")
	word, ok := greetings[lang]
	if !ok {
		// 不支持的语言回退到中文，不要返回 400 —— 这是展示型接口
		lang = "zh"
		word = greetings[lang]
	}

	// 参数校验：非法值一律用默认值兜住，绝不 panic
	count := 1
	if raw := q.Get("count"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			if n > 10 {
				n = 10 // 封顶，防止被刷
			}
			count = n
		}
	}

	messages := make([]string, 0, count)
	for i := 0; i < count; i++ {
		messages = append(messages, word+", "+name+"!")
	}

	response.OK(w, map[string]any{
		"name":     name,
		"lang":     lang,
		"count":    count,
		"messages": messages,
		"query":    r.URL.RawQuery,
	})
}
