package main

import (
	"context"

	"airbox/config"
	"airbox/logger"
	"airbox/pkg"
)

func main() {
	var ctx = context.Background()

	logger.InitializeLogger()
	defer logger.CloseLogger()

	log := logger.GetLogger(ctx, "main")
	if err := config.LoadConfig(); err != nil {
		log.Errorf("初始化config失败: %v", err)
		return
	}
	if err := pkg.InitializeCache(); err != nil {
		log.Errorf("初始化cache失败: %v", err)
		return
	}
	if err := pkg.InitializeDB(); err != nil {
		log.Errorf("初始化db失败: %v", err)
		return
	}
	if err := pkg.InitializeOSS(); err != nil {
		log.Errorf("初始化oss失败: %v", err)
		return
	}
	if err := pkg.InitializeMail(); err != nil {
		log.Errorf("初始化mail失败: %v", err)
		return
	}

	router := NewRouter().PathMapping()
	log.Fatal(router.Run(":8080"))
}
