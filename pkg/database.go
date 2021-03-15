package pkg

import (
	"fmt"
	"time"

	"airbox/config"
	"airbox/model/do"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	log "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

// InitializeDB 初始化数据库
func InitializeDB() error {
	var err error
	db, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&loc=Local&parseTime=true",
		config.GetConfig().MySQL.Username,
		config.GetConfig().MySQL.Password,
		config.GetConfig().MySQL.Host,
		config.GetConfig().MySQL.Database)), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy:         schema.NamingStrategy{SingularTable: true},
		Logger:                 log.Default,
		PrepareStmt:            true,
		CreateBatchSize:        100,
	})
	if err != nil {
		return fmt.Errorf("DB 初始化失败: %v", err)
	}
	sql, err := db.DB()
	if err != nil {
		return fmt.Errorf("DB 初始化失败: %v", err)
	}
	if err = sql.Ping(); err != nil {
		return fmt.Errorf("DB 连接失败: %v", err)
	}
	sql.SetMaxIdleConns(5)
	sql.SetMaxOpenConns(20)
	sql.SetConnMaxLifetime(30 * time.Second)
	if err := createTables(); err != nil {
		return fmt.Errorf("DB 初始化数据失败: %v", err)
	}

	return nil
}

// createTables 初始化数据表
func createTables() error {
	migrator := db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=UTF8MB4").Migrator()
	if !migrator.HasTable(&do.User{}) {
		if err := migrator.CreateTable(&do.User{}); err != nil {
			return err
		}
	}
	if !migrator.HasTable(&do.Storage{}) {
		if err := migrator.CreateTable(&do.Storage{}); err != nil {
			return err
		}
	}
	if !migrator.HasTable(&do.FileInfo{}) {
		if err := migrator.CreateTable(&do.FileInfo{}); err != nil {
			return err
		}
	}
	if !migrator.HasTable(&do.File{}) {
		if err := migrator.CreateTable(&do.File{}); err != nil {
			return err
		}
	}
	return nil
}
