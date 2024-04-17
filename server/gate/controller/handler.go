package controller

import (
	"log"
	"sgserver/config"
	"sgserver/constant"
	"sgserver/net"
	"strings"
	"sync"
)

var GateHandler = &Handler{
	proxyMap: make(map[string]map[int64]*net.ProxyClient),
}

type Handler struct {
	proxyMutex sync.Mutex
	proxyMap   map[string]map[int64]*net.ProxyClient // 代理地址-> 客户端连接(游戏客户端id-> 连接)
	loginProxy string
	gameProxy  string
}

func (h *Handler) Router(r *net.Router) {
	h.loginProxy = config.File.MustValue("gate_server", "login_proxy", "ws://127.0.0.1:8003")
	h.gameProxy = config.File.MustValue("gate_server", "game_proxy", "ws://127.0.0.1:8001")

	g := r.Group("*")
	g.AddRouter("*", h.all)
}

func (h *Handler) all(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	proxyStr := req.Body.Proxy

	// 判断是 name 是否是以 account. 开头
	if isAccount(req.Body.Name) {
		proxyStr = h.loginProxy
	} else {
		proxyStr = h.gameProxy
	}
	// 如果不是以 account. 开头，直接返回 -4
	if proxyStr == "" {
		rsp.Body.Code = constant.ProxyConnectError
		return
	}

	// ws://127.0.0.1:8003 的 value：map[int64]*net.ProxyClient，如果为空，重新赋值
	h.proxyMutex.Lock()
	_, ok := h.proxyMap[proxyStr]
	if !ok {
		h.proxyMap[proxyStr] = make(map[int64]*net.ProxyClient)
	}
	h.proxyMutex.Unlock()

	// 取 cid
	c, err := req.Conn.GetProperty("cid")
	if err != nil {
		log.Println("cid未取到")
		rsp.Body.Code = constant.InvalidParam
		return
	}

	// m[cid]
	cid := c.(int64)
	proxy, ok := h.proxyMap[proxyStr][cid]
	if !ok {
		//没有 建立连接 并发起调用
		proxy = net.NewProxyClient(proxyStr)
		h.proxyMutex.Lock()
		h.proxyMap[proxyStr][cid] = proxy
		h.proxyMutex.Unlock()

		err := proxy.Connect()

		//fmt.Println(err)

		if err != nil {
			h.proxyMutex.Lock()
			delete(h.proxyMap[proxyStr], cid)
			h.proxyMutex.Unlock()
			rsp.Body.Code = constant.ProxyConnectError
			return
		}
		proxy.SetProperty("cid", cid)
		proxy.SetProperty("proxy", proxyStr)
		proxy.SetProperty("gateConn", req.Conn)
		proxy.SetOnPush(h.onPush)
	}

	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	r, err := proxy.Send(req.Body.Name, req.Body.Msg)
	//fmt.Println("写给代理客户端", r)
	if err == nil {
		rsp.Body.Code = r.Code
		rsp.Body.Msg = r.Msg
	} else {
		rsp.Body.Code = constant.ProxyNotInConnect
		rsp.Body.Msg = nil
		return
	}
}

func (h *Handler) onPush(conn *net.ClientConn, body *net.RspBody) {
	gateConn, err := conn.GetProperty("gateConn")

	if err != nil {
		return
	}
	wc := gateConn.(net.WSConn)
	wc.Push(body.Name, body.Msg)
}

func isAccount(name string) bool {
	return strings.HasPrefix(name, "account.")
}
