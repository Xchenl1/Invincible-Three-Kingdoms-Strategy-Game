package main

import (
	"fmt"
	"sgserver/config"
	"sgserver/net"
	"sgserver/server/gate"
)

/*
	1.登录功能account.login 需要通过网关 转发 登陆服务器
	2.网关（websocket的客户端）如何和 登陆服务器 （websocket服务器）交互
	3.网关又和游戏客户端 进行交互，网关是websocket的服务器
	4.websocket的服务器 已经实现了
	5.网关：代理服务器  （代理地址  代理连接通道） 客户端连接（websocket连接）
	6.路由：接受所有的请求  网关的websocket服务端的功能
	7.
*/
//websocket的服务端 将客户端的请求转发给各个服务器 比如登陆服务器、注册服务器
func main() {
	// 读取文件的ip和端口
	host := config.File.MustValue("gate_server", "host", "127.0.0.1")
	port := config.File.MustValue("gate_server", "port", "8004")

	s := net.NewServer(host + ":" + port)
	// 需要加密
	s.NeedSecret(true)

	fmt.Println("项目启动在", host, ":", port, "登录服务器启动成功")
	gate.Init()
	s.Router(gate.Router)
	s.Start()
}
