package handler

import (
	"net/http"

	"vercel-go-demo/internal/config"
	"vercel-go-demo/internal/response"
)

// Handler 展示配置来源。这个接口存在的意义是证明：
// 在 Vercel 这种只读文件系统上，go:embed 能 100% 保证配置读得到。
//
// 传统写法 os.ReadFile("config.yaml") 在 Vercel 上会失败，因为：
//  1. 运行目录是 /var/task，不是项目根
//  2. 非编译产物默认不会打进函数包
//  3. 文件系统只读，运行时写不了任何配置
func Handler(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		response.Fail(w, http.StatusInternalServerError, "配置加载失败: "+err.Error())
		return
	}

	response.OK(w, map[string]any{
		"config":             cfg,
		"source":             "go:embed —— 编译期内嵌进二进制，零文件 IO",
		"overridable_by_env": []string{"APP_ENV", "MAX_ITEMS"},
		"hint":               "在 Vercel → Settings → Environment Variables 里配置，Redeploy 后生效",
	})
}
