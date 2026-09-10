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
├── api/                    ← Vercel 只认这个目录，每个子目录编译成一个函数
│   ├── index.go            → /api
│   ├── health/index.go     → /api/health
│   ├── hello/index.go      → /api/hello
│   ├── echo/index.go       → /api/echo
│   └── config/index.go     → /api/config
├── internal/               ← 共享代码（Vercel 不会编译成函数）
│   ├── config/             go:embed 内嵌配置 + 环境变量覆盖
│   └── response/           统一 JSON 响应
├── cmd/devserver/          ← 本地调试服务器（不参与部署）
├── public/index.html       ← 静态页面，部署后访问根路径即可
├── go.mod
└── vercel.json
```

> **为什么每个端点一个子目录，而不是 `api/hello.go` 平铺？**
> 平铺 Vercel 也能跑（它逐个文件单独编译），但本地 `go build ./...` 会报
> `Handler redeclared in this block` —— Go 把整个 `api/` 当成一个包，多个 `Handler` 冲突。
> 拆成子目录后，本地和 Vercel 两边都干净。

## 二、本地运行

```bash
go run ./cmd/devserver          # http://localhost:8080
```

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

推到 GitHub 后，在 Vercel 里 Import Project，**Framework Preset 选 Go 或 Other**，Root Directory 留空。

Vercel 会自动检测 `api/**/*.go` 并逐个编译，不需要在 `vercel.json` 里显式指定 runtime。

## 四、Vercel Go 的硬性约束（踩过就懂）

| 约束 | 说明 | 后果 |
| --- | --- | --- |
| **入口签名固定** | `api/**/*.go` 里必须导出 `func Handler(w http.ResponseWriter, r *http.Request)` | 写成 `main()` + `ListenAndServe` 部署后无响应 |
| **不能自己 Listen** | Vercel 生成 `main` 并托管监听 | 本地能跑、线上 404 |
| **文件系统只读** | 运行目录 `/var/task` 只读，只有 `/tmp` 可写且随实例销毁 | `os.ReadFile("config.yaml")` 必报错 |
| **无状态** | 实例随时创建/回收，全局变量不可靠 | 内存缓存命中率随机，别当真 |
| **有执行超时** | Hobby 10 秒；Pro/Enterprise 可用 `maxDuration` 提高 | 长任务必然被杀 |
| **方法要自己判** | 所有 HTTP 方法都进同一个 `Handler` | 不做 `r.Method` 判断等于放行全部动词 |

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
  "cleanUrls": true,
  "headers": [{ "source": "/api/(.*)", "headers": [{ "key": "Cache-Control", "value": "no-store" }] }],
  "functions": { "api/**/*.go": { "memory": 1024, "maxDuration": 10 } }
}
```

- `cleanUrls`：去掉 `.html` 后缀
- `no-store`：接口别被 CDN 缓存
- `maxDuration`：超时秒数（Hobby 最高 10）
