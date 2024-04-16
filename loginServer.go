package main

import (
	"fmt"
	"sgserver/config"
	"sgserver/net"
	"sgserver/server/login"
)

func main() {
	host := config.File.MustValue("login_server", "host", "127.0.0.1")
	port := config.File.MustValue("login_server", "port", "8004")

	s := net.NewServer(host + ":" + port)

	fmt.Println("项目启动在", host, ":", port, "登录服务器启动成功")
	login.Init()
	s.Router(login.Router)
	s.Start()
}
