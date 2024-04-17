package net

import "sync"

// ReqBody 接收前端结构体
type ReqBody struct {
	Seq   int64       `json:"seq"`
	Name  string      `json:"name"`
	Msg   interface{} `json:"msg"`
	Proxy string      `json:"proxy"`
}

// RspBody 返回前端结构体
type RspBody struct {
	Seq  int64       `json:"seq"`
	Name string      `json:"name"`
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

type WsContext struct {
	mutex    sync.RWMutex
	property map[string]interface{}
}

func (ws *WsContext) Set(key string, values interface{}) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.property[key] = values
}

func (ws *WsContext) Get(key string) interface{} {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	values, ok := ws.property[key]
	if ok {
		return values
	}
	return nil
}

type WsMsgReq struct {
	Body    *ReqBody
	Conn    WSConn
	Context *WsContext
}

type WsMsgRsp struct {
	Body *RspBody
}

// WSConn request请求 请求中会有参数 取参数
type WSConn interface {
	SetProperty(key string, valus interface{})
	GetProperty(key string) (interface{}, error)
	RemoveProperty(key string)
	Addr() string
	Push(name string, data interface{})
}

const HandshakeMsg = "handshake"

type Handshake struct {
	Key string `json:"key"`
}

const HeartbeatMsg = "heartbeat"

type Heartbeat struct {
	CTime int64 `json:"ctime"`
	STime int64 `json:"stime"`
}
