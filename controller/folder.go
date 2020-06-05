package controller

import (
	"airbox/global"
	"airbox/model"
	"airbox/service"
	"github.com/labstack/echo/v4"
	"net/http"
)

type FolderController struct {
	BaseController
	folder *service.FolderService
}

var folder *FolderController

func GetFolderController() *FolderController {
	if folder == nil {
		folder = &FolderController{
			BaseController: getBaseController(),
			folder:         service.GetFolderService(),
		}
	}
	return folder
}

// AddFolder 新建文件夹
func (f *FolderController) AddFolder(c echo.Context) error {
	name, fid := c.FormValue("name"), c.FormValue("fid")
	if name == "" {
		return c.JSON(http.StatusBadRequest, global.ErrorOfWithoutName)
	}
	folder, err := f.folder.AddFolder(name, f.auth(c).Storage.ID, fid)
	if err != nil {
		global.LOGGER.Printf("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"folder": folder,
	})
}

// DeleteFolder 删除文件夹
func (f *FolderController) DeleteFolder(c echo.Context) error {
	return f.Delete(c, f.folder.DeleteFolder)
}

// UpdateFolder 更新文件夹信息，包括重命名、移动和复制
func (f *FolderController) UpdateFolder(c echo.Context) error {
	return f.Update(c, f.folder.RenameFolder, f.folder.CopyFolder, f.folder.MoveFolder)
}

func (f *FolderController) GetFolder(c echo.Context) error {
	var folders []model.Folder
	var err error
	if fid := c.QueryParam("fid"); fid != "" {
		folders, err = f.folder.GetFolderByFatherID(fid)
	} else {
		folders, err = f.folder.GetFolderByStorageID(f.auth(c).Storage.ID)
	}
	if err != nil {
		global.LOGGER.Printf("%+v\n", err)
		return c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"folder": folders,
	})
}
