package config

import (
	"airbox/model"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"os"
	"time"
)

var DB *gorm.DB

// 初始化数据库
func initializeDB() {
	var err error
	DB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&loc=Local&parseTime=true",
		Env.DataSource.Username,
		Env.DataSource.Password,
		Env.DataSource.Host,
		Env.DataSource.Port,
		Env.DataSource.Database))
	if err != nil {
		fmt.Println("failed to initialize DB: ", err)
		os.Exit(0)
	}
	if err = DB.DB().Ping(); err != nil {
		fmt.Println("failed to ping DB: ", err)
		os.Exit(0)
	}
	DB.SingularTable(true)
	DB.DB().SetMaxIdleConns(5)
	DB.DB().SetMaxOpenConns(20)
	DB.DB().SetConnMaxLifetime(30 * time.Second)
	createTables()
}

// 初始化数据表
func createTables() {
	if !DB.HasTable(&model.User{}) {
		DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.User{})
	}
	if !DB.HasTable(&model.Storage{}) {
		DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.Storage{})
	}
	if !DB.HasTable(&model.Folder{}) {
		DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.Folder{})
	}
	if !DB.HasTable(&model.File{}) {
		DB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.File{})
	}
}
