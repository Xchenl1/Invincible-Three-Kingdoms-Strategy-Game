package net

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/forgoer/openssl"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"log"
	"sgserver/constant"
	"sgserver/utils"
	"sync"
	"time"
)

// 处理代理服务响应数据
type syncCtx struct {
	// 处理协程的上下文
	ctx     context.Context
	cancel  context.CancelFunc
	outChan chan *RspBody
}

func NewSyncCtx() *syncCtx {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	return &syncCtx{
		ctx,
		cancel,
		make(chan *RspBody),
	}
}

func (s *syncCtx) wait() *RspBody {
	// 15 秒等待服务发送信息
	defer s.cancel()
	select {
	case msg := <-s.outChan:
		fmt.Println("各服务发来的数据", msg)
		return msg
	case <-s.ctx.Done():
		fmt.Println("----------------")
		return nil
	}
}

// ClientConn 客户端连接
type ClientConn struct {
	wsConn *websocket.Conn
	// 是否关闭
	isClosed      bool
	property      map[string]interface{} // 设置属性
	propertyLock  sync.RWMutex
	Seq           int64 // 序列号
	handshake     bool
	handshakeChan chan bool
	onPush        func(conn *ClientConn, body *RspBody) // 通知代理服务器
	onClose       func(conn *ClientConn)                // 关闭处理
	syncCtxMap    map[int64]*syncCtx                    // 写入 websocket 中
	syncCtxLock   sync.RWMutex
}

func (c *ClientConn) Start() bool {
	// 一直不停的接收消息
	// 返回等待握手的消息返回
	c.handshake = false
	go c.wsReadLoop()
	return c.waitHandshake()
}

func (c *ClientConn) waitHandshake() bool {
	// 万一程序出现问题 一直响应不到；5 秒响应与其他服务的握手
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 等待握手的消息
	select {
	case <-c.handshakeChan:
		log.Println("握手成功！")
		return true
	case <-ctx.Done():
		log.Println("握手超时")
		return false
	}
}

func (c *ClientConn) wsReadLoop() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("捕捉到异常！", err)
			c.Close()
		}
	}()
	for {
		_, data, err := c.wsConn.ReadMessage()
		if err != nil {
			log.Println("收消息出现错误！", err)
			break
		}
		//收到消息会解析消息
		//1.解压unzip
		data, err = utils.UnZip(data)
		if err != nil {
			log.Println("解压失败！，格式不合法！", err)
			continue
		}
		secretKey, err := c.GetProperty("secretKey")
		if err == nil {
			// 有加密
			key := secretKey.(string)
			// 客户端传过来的数据是加密的 需要解密
			d, err := utils.AesCBCDecrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
			if err != nil {
				log.Println("数据格式有误，解密失败:", err)
			} else {
				data = d
			}
		}

		//3.data转为body
		body := &RspBody{}
		err = json.Unmarshal(data, body)
		if err != nil {
			log.Println("数据格式有误！非法格式", err)
		} else {
			// 握手 还是别的请求
			if body.Seq == 0 {
				if body.Name == HandshakeMsg {
					//获取密钥
					hs := &Handshake{}
					mapstructure.Decode(body.Msg, hs)
					if hs.Key != "" {
						c.SetProperty("secretKey", hs.Key)
					} else {
						c.RemoveProperty("secretKey")
					}
					c.handshake = true
					c.handshakeChan <- true
				} else {
					if c.onPush != nil {
						c.onPush(c, body)
					}
				}
			} else {
				c.syncCtxLock.RLock()
				ctx, ok := c.syncCtxMap[body.Seq]
				c.syncCtxLock.RUnlock()
				if ok {
					ctx.outChan <- body
				} else {
					log.Println("no seq syncCtx find")
				}
			}
		}
	}
	//退出循环直接关闭连接
	c.Close()
}

func (c *ClientConn) Close() {
	_ = c.wsConn.Close()
}

func newClientConn(wsConn *websocket.Conn) *ClientConn {
	return &ClientConn{
		wsConn:        wsConn,
		handshakeChan: make(chan bool),
		Seq:           0,
		isClosed:      false,
		property:      make(map[string]interface{}),
		syncCtxMap:    map[int64]*syncCtx{},
	}
}
func (c *ClientConn) Addr() string {
	return c.wsConn.RemoteAddr().String()
}

func (c *ClientConn) SetProperty(key string, values interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	c.property[key] = values
}

func (c *ClientConn) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()
	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

func (c *ClientConn) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	delete(c.property, key)
}

func (c *ClientConn) Push(name string, data interface{}) {
	rsp := &WsMsgRsp{Body: &RspBody{
		Name: name,
		Msg:  data,
		Seq:  0,
	}}
	//w.outChan <- rep
	c.write(rsp.Body)
}

func (c *ClientConn) write(body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
		return err
	}
	//2.前端消息是加密消息 需要进行加密
	//secretKey, err := c.GetProperty("secretKey")
	//if err == nil {
	//	//有加密
	//	key := secretKey.(string)
	//	//数据做加密
	//	data, err = utils.AesCBCEncrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
	//	if err != nil {
	//		log.Println("加密失败", err)
	//		return err
	//	}
	//}
	//压缩
	if data, err := utils.Zip(data); err == nil {
		err := c.wsConn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			log.Println("写数据失败")
			return err
		}
	} else {
		log.Println("压缩失败")
		return err
	}
	return nil
}

func (c *ClientConn) SetOnPush(hook func(conn *ClientConn, body *RspBody)) {
	c.onPush = hook
}

func (c *ClientConn) Send(name string, msg interface{}) (*RspBody, error) {
	// 把请求 发送给 游戏服务器 登陆服务器
	c.syncCtxLock.Lock()
	c.Seq += 1
	seq := c.Seq

	sc := NewSyncCtx()
	c.syncCtxMap[seq] = sc
	c.syncCtxLock.Unlock()

	// 构建 request 请求
	req := &ReqBody{Seq: seq, Name: name, Msg: msg}
	rsp := &RspBody{Seq: seq, Name: name, Code: constant.OK}

	err := c.write(req)
	if err != nil {
		sc.cancel()
	} else {
		r := sc.wait()
		if r == nil {
			rsp.Code = constant.ProxyConnectError
		} else {
			rsp = r
		}
	}
	c.syncCtxLock.Lock()
	delete(c.syncCtxMap, seq)
	c.syncCtxLock.Unlock()

	return rsp, nil
}
