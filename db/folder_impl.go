package db

import (
	"airbox/db/base"
	"airbox/model"
	"github.com/jinzhu/gorm"
)

// 文件夹数据库操作实体
type FolderDaoImpl struct {
}

// 在数据仓库sid下查看父文件夹fid下是否有文件夹名为name的文件夹
func (f *FolderDaoImpl) SelectFolderByName(db *gorm.DB, name string, sid, fid string) (*model.Folder, error) {
	tx := db.Where("storage_id = ?", sid)
	if fid != "" {
		tx = tx.Where("father_id = ?", fid)
	} else {
		tx = tx.Where("father_id is null")
	}
	folder := &model.Folder{}
	err := tx.Where("name = ?", name).First(folder).Error
	return folder, err
}

// 获取在数据仓库Sid下，父文件夹为Fid的文件夹
func (f *FolderDaoImpl) SelectFolderByStorageID(db *gorm.DB, sid string) (folders []model.Folder, err error) {
	err = db.Where("storage_id = ? and father_id is null", sid).Find(&folders).Error
	return
}

// 新增文件夹
func (f *FolderDaoImpl) InsertFolder(db *gorm.DB, folder *model.Folder) error {
	return db.Create(folder).Error
}

// 根据文件夹ID删除文件夹
func (f *FolderDaoImpl) DeleteFolderByID(db *gorm.DB, id string) error {
	return db.Where("id = ?", id).Delete(&model.Folder{}).Error
}

// 根据仓库ID删除文件夹
func (f *FolderDaoImpl) DeleteFolderBySID(db *gorm.DB, sid string) error {
	return db.Where("storage_id = ?", sid).Delete(&model.Folder{}).Error
}

// 更新文件夹信息
func (f *FolderDaoImpl) UpdateFolder(db *gorm.DB, folder *model.Folder) error {
	return db.Model(folder).Update("name", folder.Name).Error
}

// 根据文件夹ID获得文件夹
func (f *FolderDaoImpl) SelectFolderByID(db *gorm.DB, id string) (*model.Folder, error) {
	folder := &model.Folder{}
	err := db.Where("id = ?", id).First(folder).Error
	return folder, err
}

// 根据父文件夹ID获得文件夹
func (f *FolderDaoImpl) SelectFolderByFatherID(db *gorm.DB, id string) (folders []model.Folder, err error) {
	err = db.Where("father_id = ?", id).Find(&folders).Error
	return
}

var folder base.FolderDao

func GetFolderDao() base.FolderDao {
	if folder == nil {
		folder = &FolderDaoImpl{}
	}
	return folder
}
