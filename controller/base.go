package controller

import (
	"airbox/model"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"
)

// BaseController is responsible for the common operation of controller
type BaseController struct {
}

// getBaseController return instance of BaseController
func getBaseController() BaseController {
	return BaseController{}
}

// auth return the authority of request
func (*BaseController) auth(c echo.Context) *model.User {
	return c.Get("Authorization").(*model.User)
}

// downloadFile 公共使用的下载文件模块
func (*BaseController) downloadFile(c echo.Context, file *model.File) error {

	open, err := os.Open(file.Location + file.Name)
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	stat, err := open.Stat()
	if err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	c.Response().Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	c.Response().Header().Set("Content-Disposition", "attachment; filename="+url.QueryEscape(stat.Name()))
	http.ServeContent(c.Response(), c.Request(), stat.Name(), stat.ModTime(), open)
	return nil
}

// Update 更新文件夹或者文件信息，包括重命名、移动和复制
// 其中重命名需要name信息
// 移动需要fid信息，若fid为空则表明移动到根目录
// 复制则是在移动的基础上增加copy参数，当copy=true时表示复制
func (*BaseController) Update(c echo.Context, rename, copy, move func(param, id string) error) error {
	name, fid, copy2 := c.FormValue("name"), c.FormValue("fid"), c.FormValue("copy")
	id := c.Param("id")
	if name != "" {
		// 重命名
		if err := rename(name, id); err != nil {
			c.Logger().Errorf("%s\n", err.Error())
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
	} else {
		if fid == id {
			return c.JSON(http.StatusBadRequest, "不能复制或移动到自身")
		}
		if copy2 == "true" {
			// 复制
			if err := copy(fid, id); err != nil {
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
		} else if copy2 == "false" {
			// 移动
			if err := move(fid, id); err != nil {
				c.Logger().Errorf("%s\n", err.Error())
				return c.JSON(http.StatusInternalServerError, err.Error())
			}
		}
	}
	return c.NoContent(http.StatusOK)
}

// Delete 删除文件或文件夹
func (*BaseController) Delete(c echo.Context, delete func(id string) error) error {
	if err := delete(c.Param("id")); err != nil {
		c.Logger().Errorf("%s\n", err.Error())
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
