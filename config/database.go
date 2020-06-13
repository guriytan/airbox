package config

import (
	"airbox/model"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

// 初始化数据库
func InitializeDB() {
	var err error
	db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&loc=Local&parseTime=true",
		Env.DataSource.Username,
		Env.DataSource.Password,
		Env.DataSource.Host,
		Env.DataSource.Port,
		Env.DataSource.Database))
	if err != nil {
		panic(fmt.Sprintf("DB 初始化失败: %v", err))
	}
	if err = db.DB().Ping(); err != nil {
		panic(fmt.Sprintf("DB 连接失败: %v", err))
	}
	db.SingularTable(true)
	db.DB().SetMaxIdleConns(5)
	db.DB().SetMaxOpenConns(20)
	db.DB().SetConnMaxLifetime(30 * time.Second)
	createTables()
}

// 初始化数据表
func createTables() {
	if !db.HasTable(&model.User{}) {
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.User{})
	}
	if !db.HasTable(&model.Storage{}) {
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.Storage{})
	}
	if !db.HasTable(&model.Folder{}) {
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.Folder{})
	}
	if !db.HasTable(&model.File{}) {
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.File{})
	}
	if !db.HasTable(&model.FileInfo{}) {
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.FileInfo{})
	}
	if !db.HasTable(&model.FileCount{}) {
		db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").CreateTable(&model.FileCount{})
	}
}
