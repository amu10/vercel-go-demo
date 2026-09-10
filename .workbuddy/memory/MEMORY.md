# vercel-go-demo 项目长期约定

## 部署铁律
- 必须保持 **api 目录函数模式**：`vercel.json` 里 `"framework": null`。
  选 Go Framework Preset 会触发 `ignoreRuntimes: ['@vercel/go']`，
  导致 `api/**/*.go` 不被识别为函数，构建报 `unmatched-function-pattern`。
- 若哪天要迁服务器模式：入口只能是 `main.go` / `cmd/api/main.go` / `cmd/server/main.go`，
  监听 `PORT` 环境变量，`framework` 设为 `"go"`，并删除 `functions` 段。

## 结构约定
- `api/<name>/index.go` 一个端点一个子目录（避免 `Handler redeclared`），导出 `func Handler(http.ResponseWriter, *http.Request)`。
- 共享代码放 `internal/`（Vercel 不会编译成函数）；配置用 `go:embed`，禁止运行时读文件。
- `cmd/devserver` 仅本地调试，不参与部署。
