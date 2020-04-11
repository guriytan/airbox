package main

import (
	. "airbox/config"
	"airbox/controller"
	m "airbox/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io"
	"log"
	"net/http"
	"os"
)

type Router struct {
	*echo.Echo
}

func (router *Router) Init() *Router {
	// 注册日志
	errFile, err := os.OpenFile("./errors.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("打开日志文件失败：", err)
	}
	router.Logger.SetOutput(io.MultiWriter(os.Stderr, errFile))
	router.Logger.SetPrefix("[airbox]")
	DB.SetLogger(router.Logger)
	return router
}

func (router *Router) PathMapping() *Router {
	router.Use(middleware.Recover())
	router.Use(m.CORS)
	// info panel module
	panel := controller.GetPanelController()
	g := router.Group("/panel", m.Login)
	// 文件列表
	g.GET("/files", panel.ListFile)
	// 用户信息
	g.GET("/info", panel.UserInfo)
	// 类型文件列表
	g.GET("/types", panel.ListType)
	// 申请注销账户
	g.GET("/apply-unsubscribe", panel.ApplyToUnsubscribe)
	// 注销账户
	g.POST("/unsubscribe", panel.Unsubscribe)

	// file and folder module
	storage := controller.GetStorageController()
	// 分享文件
	router.GET("/share", storage.ShareFile)

	file := router.Group("/file", m.Login)
	// 上传文件
	file.POST("/upload", storage.UploadFile)
	file.OPTIONS("/upload", func(context echo.Context) error {
		return context.NoContent(http.StatusOK)
	})
	// 下载文件
	file.GET("/download", storage.DownloadFile)
	// 获取文件的分享链接
	file.GET("/get-link", storage.GetShareLink)
	// 删除文件
	file.GET("/delete", storage.DeleteFile)
	// 重命名文件
	file.POST("/rename", storage.RenameFile)

	folder := router.Group("/folder", m.Login)
	// 删除文件夹
	folder.GET("/delete", storage.DeleteFolder)
	// 新建文件夹
	folder.POST("/add", storage.AddFolder)
	// 重命名文件夹
	folder.POST("/rename", storage.RenameFolder)

	// user module
	user := controller.GetUserService()
	u := router.Group("/user")
	// 登录
	u.POST("/login", user.Login)
	// 注册
	u.POST("/register", user.Register)
	u.POST("/send-register-code", user.SendRegisterCode)
	// 忘记密码
	u.POST("/forget", user.ForgetPwd)
	u.POST("/reset-pwd", user.ResetPwd, m.CheckLink)
	// 更换密码
	u.POST("/reset-pwd-origin", user.ResetPwdByOrigin, m.Login)
	// 更换邮箱
	u.POST("/reset-email", user.ResetEmail, m.Login)
	u.POST("/send-reset-code", user.SendResetEmailCode, m.Login)

	return router
}

func NewRouter() *Router {
	return &Router{echo.New()}
}
