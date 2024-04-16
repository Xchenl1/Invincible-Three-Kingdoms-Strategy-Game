package net

import (
	"log"
	"strings"
)

type HandlerFunc func(req *WsMsgReq, rsp *WsMsgRsp)

type Router struct {
	group []*group
}

func NewRouter() *Router {
	return &Router{}
}

// 路由组
type group struct {
	prefix     string
	handlerMap map[string]HandlerFunc
}

func (r *group) AddRouter(name string, HandlerFunc HandlerFunc) {
	r.handlerMap[name] = HandlerFunc
}

func (r *Router) Group(prefix string) *group {
	g := &group{
		prefix:     prefix,
		handlerMap: make(map[string]HandlerFunc),
	}
	r.group = append(r.group, g)
	return g
}

func (g *group) exec(name string, req *WsMsgReq, rsp *WsMsgRsp) {
	h := g.handlerMap[name]
	if h != nil {
		h(req, rsp)
	} else {
		h = g.handlerMap["*"]
		if h != nil {
			h(req, rsp)
		} else {
			log.Println("路由未定义！")
		}
	}
}

func (r *Router) Run(req *WsMsgReq, rsp *WsMsgRsp) {
	//account.login
	strs := strings.Split(req.Body.Name, ".")
	prefix := ""
	name := ""
	if len(strs) == 2 {
		prefix = strs[0]
		name = strs[1]
	}
	for _, g := range r.group {
		if g.prefix == prefix {
			g.exec(name, req, rsp)
		} else if g.prefix == "*" {
			g.exec(name, req, rsp)
		}
	}
}
