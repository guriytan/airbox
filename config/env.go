package config

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v2"
)

// go:embed config.yml
var config []byte

var cfg *Config

func GetConfig() *Config {
	return cfg
}

// Config 环境配置对象
type Config struct {
	// 邮箱服务配置
	Mail struct {
		Addr     string `yaml:"addr"`     // 邮箱地址
		Port     int    `yaml:"port"`     // 邮箱端口
		Username string `yaml:"username"` // 邮箱用户名
		Password string `yaml:"password"` // 邮箱密码
	}
	// MySQL数据库服务配置
	DataSource struct {
		Host     string `yaml:"host"`     // 数据库主机
		Port     string `yaml:"port"`     // 数据库端口
		Database string `yaml:"database"` // 数据库类型
		Username string `yaml:"username"` // 数据库用户名
		Password string `yaml:"password"` // 数据库密码
		MinIdle  int    `yaml:"min-idle"` // 数据库连接池最小维持数量
		MaxIdle  int    `yaml:"max-idle"` // 数据库连接池数量
		Timeout  int    `yaml:"timeout"`  // 数据库连接超时时间
	}
	// Redis缓存服务配置
	Redis struct {
		Host     string `yaml:"host"`     // Redis主机
		Port     string `yaml:"port"`     // Redis端口
		Password string `yaml:"password"` // Redis密码
		Pool     int    `yaml:"pool"`     // Redis连接池数量
		MinIdle  int    `yaml:"min-idle"` // Redis连接池最小维持数量
		Timeout  int    `yaml:"timeout"`  // Redis连接超时时间
	}
	Upload struct {
		Goroutine int    `yaml:"goroutine"` // Default routine, 2
		Timeout   int    `yaml:"timeout"`   // Timeout for transmission, second
		Dir       string `yaml:"dir"`       // Default Store Path, don't end with "/"
		PartSize  int    `yaml:"part-size"` // Default download part size, KB
	}
	// 服务器配置
	Web struct {
		Port string `yaml:"port"` // 服务器绑定端口
		Site string `yaml:"site"` // 前端网站，主要用于json跨域处理
	}
	Register bool `yaml:"register"` // 注册权限
}

// 加载配置文件
func LoadConfig() error {
	cfg = &Config{}
	if err := yaml.Unmarshal(config, cfg); err != nil {
		return fmt.Errorf("config 初始化失败: %v", err)
	}
	return nil
}
