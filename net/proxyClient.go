package net

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

// ProxyClient websocket客户端
type ProxyClient struct {
	proxy string
	conn  *ClientConn
}

func NewProxyClient(proxy string) *ProxyClient {
	return &ProxyClient{
		proxy: proxy,
	}
}

func (c *ProxyClient) SetProperty(key string, data interface{}) {
	if c.conn != nil {
		c.conn.SetProperty(key, data)
	}
}

func (c *ProxyClient) Connect() error {
	// 去连接 websocket 服务端
	// 通过 Dialer 连接websocket服务器
	var dialer = websocket.Dialer{
		Subprotocols:     []string{"p1", "p2"},
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 30 * time.Second,
	}

	// 获取 ws 连接
	ws, _, err := dialer.Dial(c.proxy, nil)
	if err == nil {
		c.conn = newClientConn(ws)
		fmt.Println("创建ws连接！")
		if !c.conn.Start() {
			return errors.New("握手失败")
		}
	}
	return err
}

func (c *ProxyClient) SetOnPush(hook func(conn *ClientConn, body *RspBody)) {
	if c.conn != nil {
		c.conn.SetOnPush(hook)
	}
}

func (c *ProxyClient) Send(name string, msg interface{}) (*RspBody, error) {
	if c.conn != nil {
		//fmt.Println("发送数据", c.proxy)
		return c.conn.Send(name, msg)
	}
	return nil, errors.New("连接未发现...")
}
