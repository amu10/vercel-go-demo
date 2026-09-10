# vercel-go-demo 项目长期约定

## 部署铁律
- 必须保持 **api 目录函数模式**：`vercel.json` 里 `"framework": null`。
  选 Go Framework Preset 会触发 `ignoreRuntimes: ['@vercel/go']`，
  导致 `api/**/*.go` 不被识别为函数，构建报 `unmatched-function-pattern`。
- 若哪天要迁服务器模式：入口只能是 `main.go` / `cmd/api/main.go` / `cmd/server/main.go`，
  监听 `PORT` 环境变量，`framework` 设为 `"go"`，并删除 `functions` 段。

## 双模式结构（业务代码只写一份）
- 业务实现放 `pkg/handlers`（Index/Health/Hello/Echo/Config），两种模式共用。
- 函数模式入口 `api/<name>/index.go`：一行转发 `func Handler(w,r){ handlers.Xxx(w,r) }`。
- 服务器模式入口 `cmd/api/main.go`：ServeMux 注册 + `ListenAndServe(":"+os.Getenv("PORT"))`。
- 根包 `site.go`（package site，导入路径 `vercel-go-demo`）`//go:embed public/index.html`：
  embed 不能引用父目录，所以只有根包能 embed 到 public/；服务器模式靠它托管首页。
- `vercel.json` = 函数模式（当前）；`vercel.server.json` = 服务器模式。
  切换：`vercel deploy --local-config vercel.server.json` 或 `cp` 覆盖。

## 结构约定
- `api/<name>/index.go` 一个端点一个子目录（避免 `Handler redeclared`），导出 `func Handler(http.ResponseWriter, *http.Request)`。
- 共享代码放 `pkg/`，**绝对不能叫 `internal/`**：Vercel 函数模式构建时会改写模块路径
  （报错形如 `imports handler/api/config`），internal 包相对它变成外部模块被拒绝。
  配置用 `go:embed`，禁止运行时读文件。
- `cmd/devserver` 仅本地调试（硬编码 8080，不走 PORT），不参与部署。
