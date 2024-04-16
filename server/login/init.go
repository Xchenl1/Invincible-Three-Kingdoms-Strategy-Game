package login

import (
	"sgserver/db"
	"sgserver/net"
	"sgserver/server/login/controller"
)

var Router = net.NewRouter()

func Init() {
	//测试并初始化数据库
	db.TestDB()
	initRouter()
}

func initRouter() {
	controller.DefaultAccount.Router(Router)
}
