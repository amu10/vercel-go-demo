# Vercel Go Runtime 示例项目

能在 Vercel 上跑起来的 Go API 服务。**无状态、不碰本地磁盘、配置用 `go:embed` 内嵌** —— 专为 Serverless 环境设计。

```
GET  /api                                    路由清单
GET  /api/health                             健康检查 + 运行环境
GET  /api/hello?name=张三&lang=zh&count=3     查询参数演示
POST /api/echo    {"message":"hi"}           JSON 请求体演示
GET  /api/config                             go:embed 配置演示
```

## 一、目录结构

```
vercel-go-demo/
├── api/                    ← 【函数模式】Vercel 只认这个目录，每个子目录编译成一个函数
│   ├── index.go            → /api
│   ├── health/index.go     → /api/health
│   ├── hello/index.go      → /api/hello
│   ├── echo/index.go       → /api/echo
│   └── config/index.go     → /api/config
├── pkg/                    ← 共享代码（**不能叫 internal/**，原因见第八节）
│   ├── handlers/           业务实现，两种模式共用这一份
│   ├── config/             go:embed 内嵌配置 + 环境变量覆盖
│   └── response/           统一 JSON 响应
├── cmd/
│   ├── api/                ← 【服务器模式】入口，ServeMux 汇总全部路由
│   └── devserver/          ← 本地调试服务器（不参与部署）
├── public/index.html       ← 静态页面，部署后访问根路径即可
├── site.go                 ← 根包，go:embed 内嵌 public/（服务器模式托管首页用）
├── go.mod
├── vercel.json             ← 函数模式配置（当前生效）
└── vercel.server.json      ← 服务器模式配置（切换时覆盖 vercel.json）
```

> **为什么业务代码在 `pkg/handlers`，`api/` 里只剩一行转发？**
> 这样两种部署模式共用同一份实现：`api/hello/index.go` 是
> `func Handler(w, r) { handlers.Hello(w, r) }`，`cmd/api/main.go` 里是
> `mux.HandleFunc("/api/hello", handlers.Hello)`。切模式只动 `vercel.json`，业务零改动。

> **为什么每个端点一个子目录，而不是 `api/hello.go` 平铺？**
> 平铺 Vercel 也能跑（它逐个文件单独编译），但本地 `go build ./...` 会报
> `Handler redeclared in this block` —— Go 把整个 `api/` 当成一个包，多个 `Handler` 冲突。
> 拆成子目录后，本地和 Vercel 两边都干净。

## 二、本地运行

**跑函数模式那份（等价于线上的 `api/` 目录）：**

```bash
go run ./cmd/devserver          # http://localhost:8080
```

**跑服务器模式那份（等价于线上的 `cmd/api`）：**

```bash
PORT=8081 go run ./cmd/api      # http://localhost:8081
```

两者的业务实现是同一份 `pkg/handlers`，路由和响应完全一致。

带环境变量跑（验证配置覆盖）：

```bash
MAX_ITEMS=2 APP_ENV=production go run ./cmd/devserver
curl -X POST http://localhost:8080/api/echo \
  -H 'Content-Type: application/json' \
  -d '{"message":"hi","items":["a","b","c","d","e"]}'
# 返回 items 只剩 2 条，truncated=true —— 环境变量生效了
```

浏览器打开 http://localhost:8080 有个交互测试页，点按钮就能调各接口。

## 三、部署到 Vercel

**方式一：CLI**

```bash
npm i -g vercel
vercel deploy          # 预览
vercel deploy --prod   # 生产
```

**方式二：Git 导入**

推到 GitHub 后，在 Vercel 里 Import Project：

- **Framework Preset 选 `Other`**（`vercel.json` 里已用 `"framework": null` 锁死）
- Root Directory 留空

Vercel 会自动检测 `api/**/*.go` 并逐个编译，不需要在 `vercel.json` 里显式指定 runtime。

> ⚠️ **Framework Preset 千万别选 `Go`。**
> 选了 Go 会走 Vercel 的 Go Framework Preset（服务器模式，要求 `main.go` / `cmd/api/main.go` /
> `cmd/server/main.go` 入口），该模式会把 `@vercel/go` 放进 `ignoreRuntimes`，
> `api/` 下的文件不再被识别为函数，构建直接报错：
> `The pattern "api/**/*.go" defined in functions doesn't match any Serverless Functions`。
> 服务器模式见下一节。

### 方式三：Go Framework Preset（服务器模式）

2026 年起 Vercel 主推的形态：**一个 Go 二进制跑全部路由**，不再是「一个目录一个函数」。
本项目已经把这条路铺好了，入口就是 `cmd/api/main.go`：

```go
mux := http.NewServeMux()
mux.HandleFunc("/api", handlers.Index)
mux.HandleFunc("/api/health", handlers.Health)
mux.HandleFunc("/api/hello", handlers.Hello)
mux.HandleFunc("/api/echo", handlers.Echo)
mux.HandleFunc("/api/config", handlers.Config)
mux.HandleFunc("/", site.Handler)          // 首页：public/ 不再被托管，得靠 go:embed
log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), mux))
```

**切换方式**（二选一）：

```bash
# CLI：用 --local-config 指定，不动主配置
vercel deploy --local-config vercel.server.json

# 或 Git 集成：直接覆盖（记得切回来时用 git checkout vercel.json）
cp vercel.server.json vercel.json
```

`vercel.server.json` 的内容就三件事：`"framework": "go"`、`cleanUrls`、接口不缓存。
**注意它没有 `functions` 段** —— 服务器模式下 `api/**/*.go` 匹配不到任何东西，写了必报错。

| | 函数模式 | 服务器模式 |
| --- | --- | --- |
| 产物 | 5 个独立函数，各自冷启动 | 1 个二进制，1 次冷启动 |
| 路由 | 目录路径即路由 | 代码里的 `ServeMux` |
| 首页 | Vercel 自动托管 `public/` | 必须 `go:embed`（见根包 `site`） |
| 能用 gin/chi | 不能 | 能 |
| 内存共享 | 不行，实例互相隔离 | 可以，进程内缓存有效 |

代价：丢掉「每目录一函数」的隔离性，一个 panic 会波及整个进程（虽然 Vercel 会重启）。

## 四、Vercel Go 的硬性约束（踩过就懂）

| 约束 | 说明 | 后果 |
| --- | --- | --- |
| **入口签名固定** | `api/**/*.go` 里必须导出 `func Handler(w http.ResponseWriter, r *http.Request)` | 写成 `main()` + `ListenAndServe` 部署后无响应 |
| **不能自己 Listen** | Vercel 生成 `main` 并托管监听 | 本地能跑、线上 404 |
| **文件系统只读** | 运行目录 `/var/task` 只读，只有 `/tmp` 可写且随实例销毁 | `os.ReadFile("config.yaml")` 必报错 |
| **无状态** | 实例随时创建/回收，全局变量不可靠 | 内存缓存命中率随机，别当真 |
| **有执行超时** | Hobby 10 秒；Pro/Enterprise 可用 `maxDuration` 提高 | 长任务必然被杀 |
| **方法要自己判** | 所有 HTTP 方法都进同一个 `Handler` | 不做 `r.Method` 判断等于放行全部动词 |
| **不能 import `internal/`** | Vercel 构建函数时会改写模块路径 | `use of internal package ... not allowed`，详见第八节 |

## 五、配置怎么读（重点）

这是最容易踩的坑。**在 Vercel 上 `os.ReadFile("config.yaml")` 一定失败**，三个原因叠加：

1. 运行目录是 `/var/task`，不是项目根
2. 非编译产物默认不进函数包
3. 文件系统只读

三种方案按推荐度排序：

### ✅ 方案一：go:embed（本项目采用）

```go
import "embed"

//go:embed config.json
var embedded embed.FS

func Load() (*Config, error) {
    raw, _ := embedded.ReadFile("config.json")   // 编译期内嵌，零文件 IO
    ...
}
```

编译期打进二进制，**100% 读得到**，也完全不受只读文件系统影响。适合不常变的配置。

### ✅ 方案二：环境变量（需要按环境改的字段）

```go
env := os.Getenv("APP_ENV")   // Vercel → Settings → Environment Variables
```

改了要 Redeploy。本项目的 `APP_ENV`、`MAX_ITEMS` 走这条路，优先级高于 `config.json`。

### ⚠️ 方案三：includeFiles（能用但不推荐）

```json
{
  "functions": {
    "api/**/*.go": { "includeFiles": "config/**" }
  }
}
```

文件会被塞进 `/var/task`，但**运行时相对路径容易对不上**，得反复试。能用 go:embed 就别用这个。

## 六、接口说明

**GET /api/hello**

| 参数 | 默认 | 说明 |
| --- | --- | --- |
| `name` | `World` | 名字 |
| `lang` | `zh` | `zh`/`en`/`ja`，其他值回退 `zh` |
| `count` | 1 | 1–10，超出封顶，非法值兜底为 1 |

**POST /api/echo**

```json
{ "message": "必填", "items": ["可选数组，受 MAX_ITEMS 限制"] }
```

异常分支返回：`400` 空体 / JSON 非法 / 缺 `message`；`405` 非 POST。

## 七、vercel.json

```json
{
  "framework": null,
  "cleanUrls": true,
  "headers": [{ "source": "/api/(.*)", "headers": [{ "key": "Cache-Control", "value": "no-store" }] }],
  "functions": { "api/**/*.go": { "memory": 1024, "maxDuration": 10 } }
}
```

- `framework: null`：**关键**。禁止 Vercel 走 Go Framework Preset，保住 `api/` 目录函数模式
- `cleanUrls`：去掉 `.html` 后缀
- `no-store`：接口别被 CDN 缓存
- `maxDuration`：超时秒数（Hobby 最高 10）

## 七之一、报 `unmatched-function-pattern` 怎么排查

按命中率排序：

| # | 原因 | 判断方法 | 修法 |
| --- | --- | --- | --- |
| 1 | Framework Preset 被设为 `Go`（或服务端按 `go.mod` 判成了 Go） | 构建日志里 Framework 一行显示 `Go` | `vercel.json` 加 `"framework": null`，或面板改 `Other` |
| 2 | Root Directory 指错，Vercel 看不到 `api/` | 构建日志没有 `api/xxx/index.go` 相关输出 | Root Directory 留空或指向 `go.mod` 所在层 |
| 3 | `.vercelignore` / 未提交，`api/**/*.go` 实际没上传 | `git ls-files \| grep api` 为空 | 提交文件或改 ignore |
| 4 | 只是想止血 | — | 直接删掉 `functions` 段：`memory` 1024 和 `maxDuration` 10 本来就是默认值，删了零损失 |

> 补充：`functions` 的 key 是用 `minimatch` 去匹配**源文件相对路径**的，
> `api/**/*.go` 能同时命中 `api/index.go` 和 `api/hello/index.go`（`**` 可匹配零层），
> 所以 pattern 本身没问题——问题永远在于「这些文件有没有被识别成函数」。

## 八、为什么共享代码叫 `pkg/` 而不是 `internal/`

这是 Vercel Go **函数模式**最反直觉的一条限制。

Vercel 构建 `api/` 下的函数时，会生成一个临时的 `main__vc__go__.go` 并**改写模块路径**，
所以报错长这样：

```
Error: Command failed: go build -ldflags -s -w -o /tmp/xxx/bootstrap /vercel/path0/main__vc__go__.go
package command-line-arguments
imports handler/api/config
index.go:6:2: use of internal package vercel-go-demo/internal/config not allowed
```

重点看 `imports handler/api/config` —— 模块名从 `vercel-go-demo` 被改成了包名 `handler`。
这下 `vercel-go-demo/internal/config` 相对它就变成了「别的模块」，
Go 的 internal 可见性规则（只允许同一模块树内引用）当场拒绝。

三条路：

| 方案 | 做法 | 适用 |
| --- | --- | --- |
| 改用普通包（本项目采用） | `internal/` → `pkg/`，引用 `vercel-go-demo/pkg/xxx` | 最省事 |
| 公开桥接包 | `api/x.go` → `pkg/app`（普通包）→ `internal/xxx` | 想保留 internal 语义时 |
| 走服务器模式 | `framework: "go"`，整个模块一起 `go build`，不改写路径 | 已经是服务器模式就无所谓 |

**本地 `go build ./...` 发现不了这个问题** —— 本地模块路径正常，internal 完全合法，
只有部署到 Vercel 函数模式才炸。别指望本地编译给你兜底。
