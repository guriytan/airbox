package config

import (
	"fmt"
	"time"

	"airbox/model"

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
	db, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&loc=Local&parseTime=true",
		GetConfig().DataSource.Username,
		GetConfig().DataSource.Password,
		GetConfig().DataSource.Host,
		GetConfig().DataSource.Port,
		GetConfig().DataSource.Database)), &gorm.Config{
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
	if !migrator.HasTable(&model.User{}) {
		if err := migrator.CreateTable(&model.User{}); err != nil {
			return err
		}
	}
	if !migrator.HasTable(&model.Storage{}) {
		if err := migrator.CreateTable(&model.Storage{}); err != nil {
			return err
		}
	}
	if !migrator.HasTable(&model.File{}) {
		if err := migrator.CreateTable(&model.File{}); err != nil {
			return err
		}
	}
	if !migrator.HasTable(&model.FileInfo{}) {
		if err := migrator.CreateTable(&model.FileInfo{}); err != nil {
			return err
		}
	}
	return nil
}
