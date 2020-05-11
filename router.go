package main

import (
	. "airbox/config"
	"airbox/controller"
	middleware2 "airbox/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"os"
)

type Router struct {
	*echo.Echo
}

func (router *Router) Init() *Router {
	// 注册日志
	router.Logger.SetOutput(os.Stderr)
	router.Logger.SetPrefix("[airbox]")
	DB.SetLogger(router.Logger)
	return router
}

func (router *Router) PathMapping() *Router {
	router.Use(middleware.Recover())
	//router.Use(middleware.CORS())

	// info组，负责一些显示的数据以及文件分享api
	info := router.Group("/info", middleware2.Login)
	infoController := controller.GetInfoController()
	info.GET("/list", infoController.ListFile)   // 获取文件和文件夹列表。query: fid
	info.GET("/user", infoController.UserInfo)   // 获取用户信息。
	info.GET("/type", infoController.ListType)   // 获取类型文件列表。query: type
	info.GET("/share", infoController.ShareFile) // 下载分享的文件。query: link

	// auth组，负责一些提供权限的api
	auth := router.Group("/auth")
	authController := controller.GetAuthController()
	auth.POST("/token", authController.LoginToken)                               // 获取登录token。form: user, password
	auth.POST("/unsubscribe", authController.UnsubscribeCode, middleware2.Login) // 获取注销账号captcha。
	auth.POST("/share", authController.ShareLink, middleware2.Login)             // 获取文件分享link。form: id（文件）
	auth.POST("/register", authController.RegisterCode, middleware2.Login)       // 获取注册账号captcha。form: username, email
	auth.POST("/password", authController.PasswordCode, middleware2.Login)       // 获取重置密码link。form: user
	auth.POST("/email", authController.EmailCode, middleware2.Login)             // 获取重置邮箱captcha。form: email, password

	// file组，负责文件相关操作的api
	file := router.Group("/file", middleware2.Login)
	fileController := controller.GetFileController()
	file.POST("/upload", fileController.UploadFile) // 上传文件。query: fid, size. (optional form: relative_path)
	file.GET("/:id", fileController.DownloadFile)   // 下载文件
	file.DELETE("/:id", fileController.DeleteFile)  // 删除文件
	file.PUT("/:id", fileController.UpdateFile)     // 修改文件（包括重命名、移动、复制）。form: name / fid, copy

	// folder组，负责文件夹相关操作的api
	folder := router.Group("/folder", middleware2.Login)
	folderController := controller.GetFolderController()
	folder.GET("/get", folderController.GetFolder)       // 获取fid下的文件夹列表。query: fid
	folder.POST("/new", folderController.AddFolder)      // 新建文件夹。form: name, fid
	folder.DELETE("/:id", folderController.DeleteFolder) // 删除文件夹
	folder.PUT("/:id", folderController.UpdateFolder)    // 修改文件夹（包括重命名、移动、复制）。form: name / fid, copy

	// user组，负责账号的一些操作api
	user := router.Group("/user")
	userController := controller.GetUserController()
	user.POST("/new", userController.Register)                                    // 注册账号。form: username, password, email, code
	user.PUT("/password", userController.ResetPwd, middleware2.CheckLink)         // 忘记密码。form: token, password
	user.PUT("/:id/email", userController.ResetEmail, middleware2.Login)          // 修改密码。form: email, code
	user.PUT("/:id/password", userController.ResetPwdByOrigin, middleware2.Login) // 修改邮箱。form: origin, password
	user.DELETE("/:id", userController.Unsubscribe, middleware2.Login)            // 删除账号。query: code

	return router
}

func NewRouter() *Router {
	return &Router{echo.New()}
}
