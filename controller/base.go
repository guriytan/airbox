package controller

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"airbox/global"
	"airbox/logger"
	"airbox/model"

	"github.com/gin-gonic/gin"
)

// BaseController is responsible for the common operation of controller
type BaseController struct {
}

// getBaseController return instance of BaseController
func getBaseController() BaseController {
	return BaseController{}
}

// verify return the authority of request
func (*BaseController) auth(c *gin.Context) *model.User {
	user, ok := c.Get("Authorization")
	if ok {
		return user.(*model.User)
	}
	return nil
}

// downloadFile 公共使用的下载文件模块
func (*BaseController) downloadFile(c *gin.Context, file *model.File) {
	ctx := c.Copy()

	log := logger.GetLogger(ctx, "downloadFile")
	open, err := os.Open(file.FileInfo.Path + file.FileInfo.Name)
	if err != nil {
		log.Infof("%+v\n", err)
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	stat, err := open.Stat()
	if err != nil {
		log.Infof("%+v\n", err)
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(stat.Name()))
	http.ServeContent(c.Writer, c.Request, stat.Name(), stat.ModTime(), open)
	defer func() {
		if err = open.Close(); err != nil {
			log.Infof("%+v\n", err)
		}
	}()
}

// Update 更新文件夹或者文件信息，包括重命名、移动和复制
// 其中重命名需要name信息
// 移动需要fid信息，若fid为空则表明移动到根目录
// 复制则是在移动的基础上增加copy参数，当copy=true时表示复制
func (*BaseController) Update(c *gin.Context, rename, copy, move func(ctx context.Context, param, id string) error) {
	ctx := c.Copy()

	log := logger.GetLogger(ctx, "Update")
	name, fid, copy2 := c.PostForm("name"), c.PostForm("fid"), c.PostForm("copy")
	id := c.Param("id")
	if name != "" {
		// 重命名
		if err := rename(ctx, name, id); err != nil {
			log.Infof("%+v\n", err)
			c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
			return
		}
	} else {
		if fid == id {
			c.JSON(http.StatusBadRequest, global.ErrorOfCopyFile)
			return
		}
		if copy2 == "true" {
			// 复制
			if err := copy(ctx, fid, id); err != nil {
				log.Infof("%+v\n", err)
				c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
				return
			}
		} else if copy2 == "false" {
			// 移动
			if err := move(ctx, fid, id); err != nil {
				log.Infof("%+v\n", err)
				c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
				return
			}
		}
	}
	c.Status(http.StatusOK)
}

// Delete 删除文件或文件夹
func (*BaseController) Delete(c *gin.Context, delete func(ctx context.Context, id string) error) {
	ctx := c.Copy()

	log := logger.GetLogger(ctx, "Delete")
	if err := delete(ctx, c.Param("id")); err != nil {
		log.Infof("%+v\n", err)
		c.JSON(http.StatusInternalServerError, global.ErrorOfSystem)
		return
	}
	c.Status(http.StatusOK)
}
