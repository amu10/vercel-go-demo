// Package config 演示如何在 Vercel 这种「只读 + 无状态」的 Serverless 环境里
// 可靠地读到配置。
//
// 背景：Vercel 的函数运行目录是 /var/task，文件系统只读，且只会打包编译产物。
// 传统写法 os.ReadFile("config.yaml") 必然报 no such file or directory。
//
// 解法：用 go:embed 在编译期就把配置打进二进制，运行时零文件 IO。
// 需要按环境改的字段，再用环境变量覆盖（Vercel 控制台里配）。
package config

import (
	"embed"
	"encoding/json"
	"os"
	"strconv"
	"sync"
)

//go:embed config.json
var embedded embed.FS

// Config 是应用的配置结构。新增字段记得同步改 config.json。
type Config struct {
	AppName  string   `json:"app_name"`
	Version  string   `json:"version"`
	Greeting string   `json:"greeting"`
	Features []string `json:"features"`

	// 以下字段由环境变量注入，config.json 里不存在
	Env      string `json:"env"`
	MaxItems int    `json:"max_items"`
}

var (
	once    sync.Once
	cached  *Config
	loadErr error
)

// Load 返回单例配置。第一次调用时解析，之后复用。
// 用 sync.Once 是因为 Serverless 实例会被复用，没必要每次请求都重新解析。
func Load() (*Config, error) {
	once.Do(func() {
		var c Config

		raw, err := embedded.ReadFile("config.json")
		if err != nil {
			loadErr = err
			return
		}
		if err := json.Unmarshal(raw, &c); err != nil {
			loadErr = err
			return
		}

		// 环境变量覆盖：Vercel 上在 Settings → Environment Variables 里配
		c.Env = getenv("APP_ENV", "development")
		c.MaxItems = getenvInt("MAX_ITEMS", 10)

		cached = &c
	})
	return cached, loadErr
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
