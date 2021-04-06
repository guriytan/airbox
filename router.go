package main

import (
	"airbox/controller"
	"airbox/middleware"

	"github.com/gin-gonic/gin"
)

type Router struct {
	*gin.Engine
}

func (router *Router) PathMapping() *Router {
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.InjectContext)

	// info组，负责一些显示的数据以及文件分享api
	info := router.Group("/info", middleware.Login)
	infoController := controller.GetInfoController()
	info.GET("/list", infoController.ListFile)   // 获取文件和文件夹列表
	info.GET("/user", infoController.UserInfo)   // 获取用户信息
	info.GET("/type", infoController.ListType)   // 获取类型文件列表
	info.GET("/share", infoController.ShareFile) // 下载分享的文件

	// auth组，负责一些提供权限的api
	auth := router.Group("/auth")
	authController := controller.GetAuthController()
	auth.POST("/token", authController.LoginToken)                              // 获取登录token
	auth.POST("/unsubscribe", middleware.Login, authController.UnsubscribeCode) // 获取注销账号captcha
	auth.POST("/share", middleware.Login, authController.ShareLink)             // 获取文件分享link
	auth.POST("/register", authController.RegisterCode)                         // 获取注册账号captcha
	auth.POST("/password", authController.PasswordCode)                         // 获取重置密码link
	auth.POST("/email", middleware.Login, authController.EmailCode)             // 获取重置邮箱captcha

	// file组，负责文件相关操作的api
	file := router.Group("/file", middleware.Login)
	fileController := controller.GetFileController()
	file.POST("/new", fileController.NewFile)           // 上传文件
	file.POST("/upload", fileController.UploadFile)     // 上传文件
	file.GET("/:file_id", fileController.DownloadFile)  // 下载文件
	file.DELETE("/:file_id", fileController.DeleteFile) // 删除文件
	file.PUT("/update", fileController.UpdateFile)      // 修改文件（包括重命名、移动、复制）

	// user组，负责账号的一些操作api
	user := router.Group("/user")
	userController := controller.GetUserController()
	user.POST("/new", userController.Register)                                   // 注册账号
	user.PUT("/password", userController.ResetPwd)                               // 忘记密码
	user.DELETE("/:id", middleware.Login, userController.Unsubscribe)            // 删除账号
	user.PUT("/email/:id", middleware.Login, userController.ResetEmail)          // 修改密码
	user.PUT("/password/:id", middleware.Login, userController.ResetPwdByOrigin) // 修改邮箱

	return router
}

func NewRouter() *Router {
	return &Router{gin.Default()}
}
