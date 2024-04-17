package net

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Server struct {
	addr       string
	router     *Router
	needSecret bool // 加密
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (s *Server) NeedSecret(needSecret bool) {
	s.needSecret = needSecret
}

func (s *Server) Router(router *Router) {
	s.router = router
}

// Start 启动服务
func (s *Server) Start() {
	http.HandleFunc("/", s.wsHandler)
	err := http.ListenAndServe(s.addr, nil)
	if err != nil {
		panic(err)
	}
}

// 升级为websocket
var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {

	//websocket  1.将http协议升级位websocket协议
	wsConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("websocket服务连接出错", err)
	}
	fmt.Println("websocket服务连接成功！")
	//客户端发消息 类似于{Name:"account.login"} 收到之后 进行解析 表示需要处理go

	wsServer := NewWsServer(wsConn, s.needSecret)
	wsServer.Router(s.router)
	wsServer.Start()

	wsServer.Handshake()
}
