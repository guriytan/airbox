package config

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v2"
)

//go:embed config.yml
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
	} `yaml:"mail"`
	// MySQL数据库服务配置
	MySQL struct {
		Host     string `yaml:"host"`     // 数据库主机
		Database string `yaml:"database"` // 数据库名称
		Username string `yaml:"username"` // 数据库用户名
		Password string `yaml:"password"` // 数据库密码
	} `yaml:"mysql"`
	// Redis缓存服务配置
	Redis struct {
		Host     string `yaml:"host"`     // Redis主机
		Password string `yaml:"password"` // Redis密码
	} `yaml:"redis"`
	MinIO struct {
		Endpoint  string `yaml:"endpoint"`
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
	} `yaml:"minio"`
	// 服务器配置
	Web struct {
		Site string `yaml:"site"` // 前端网站，主要用于json跨域处理
	} `yaml:"web"`
	Register bool `yaml:"register"` // 注册权限
}

// LoadConfig 加载配置文件
func LoadConfig() error {
	cfg = &Config{}
	if err := yaml.Unmarshal(config, cfg); err != nil {
		return fmt.Errorf("config 初始化失败: %v", err)
	}
	return nil
}
