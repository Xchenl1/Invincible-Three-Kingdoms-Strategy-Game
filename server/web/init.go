package web

import (
	"github.com/gin-gonic/gin"
	"sgserver/db"
	"sgserver/server/web/controller"
	"sgserver/server/web/midderware"
)

func Init(router *gin.Engine) {
	//测试并初始化数据库
	db.TestDB()
	initRouter(router)
}

func initRouter(router *gin.Engine) {
	router.Use(midderware.Cors()) //跨域
	router.Any("/account/register", controller.DefaultAccountController.Register)
}
