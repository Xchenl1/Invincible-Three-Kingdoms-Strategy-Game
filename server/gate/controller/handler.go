package controller

import (
	"fmt"
	"sgserver/net"
)

var GateHandler = &Handler{}

type Handler struct {
}

func (h *Handler) Router(r *net.Router) {
	g := r.Group("*")
	g.AddRouter("*", h.all)
}

func (h *Handler) all(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	fmt.Println("网关处理器！")
}
