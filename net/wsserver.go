package net

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/forgoer/openssl"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"log"
	"sgserver/utils"
	"sync"
	"time"
)

// websocket服务
type wsServer struct {
	wsConn       *websocket.Conn        //websocket连接
	router       *Router                //路由
	outChan      chan *WsMsgRsp         //写队列
	Seq          int64                  //状态信息
	propertyLock sync.RWMutex           //读写锁
	property     map[string]interface{} //存储一些信息
	needSecret   bool                   //是否加密
}

var cid int64

func NewWsServer(wsConn *websocket.Conn, needSecret bool) *wsServer {
	s := &wsServer{
		wsConn:     wsConn,
		outChan:    make(chan *WsMsgRsp, 1000),
		property:   make(map[string]interface{}),
		Seq:        0,
		needSecret: needSecret,
	}
	cid++
	s.SetProperty("cid", cid)
	return s
}

func (w *wsServer) Router(router *Router) {
	w.router = router
}

func (w *wsServer) SetProperty(key string, values interface{}) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	w.property[key] = values
}

func (w *wsServer) GetProperty(key string) (interface{}, error) {
	w.propertyLock.RLock()
	defer w.propertyLock.RUnlock()
	if value, ok := w.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

func (w *wsServer) RemoveProperty(key string) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	delete(w.property, key)
}

func (w *wsServer) Addr() string {
	return w.wsConn.RemoteAddr().String()
}

func (w *wsServer) Push(name string, data interface{}) { // 推送到通道中
	rep := &WsMsgRsp{Body: &RspBody{
		Name: name,
		Msg:  data,
		Seq:  0,
	}}
	w.outChan <- rep
}

// Start 通道建立收发消息需要一直监听才行
func (w *wsServer) Start() {
	go w.writeMsgLoop()
	go w.readMsgLoop()
}

// 读通道然后处理发送给前端
func (w *wsServer) writeMsgLoop() {
	for {
		select {
		// 有缓冲区和无缓冲区
		case msg := <-w.outChan:
			w.Write(msg.Body)
		}
	}
}

// 读到客户端发送过来的数据,发送到通道中就可以
func (w *wsServer) readMsgLoop() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("服务端捕捉到异常", err)
			w.Close()
		}
	}()
	for {
		_, data, err := w.wsConn.ReadMessage()
		if err != nil {
			log.Println("收消息出现错误！", err)
			break
		}
		//收到消息 解析消息 前端发送过来的消息 就是json格式
		//1. data 解压 unzip
		data, err = utils.UnZip(data)
		if err != nil {
			log.Println("解压数据出错，非法格式：", err)
			continue
		}
		//2. 前端的消息 加密消息 进行解密
		if w.needSecret {
			secretKey, err := w.GetProperty("secretKey")
			if err == nil {
				//有加密
				key := secretKey.(string)
				//客户端传过来的数据是加密的 需要解密
				d, err := utils.AesCBCDecrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
				if err != nil {
					log.Println("数据格式有误，解密失败:", err)
					//出错后 发起握手
					w.Handshake()
				} else {
					data = d
				}
			}
		}

		//3. data 转为body
		body := &ReqBody{}
		err = json.Unmarshal(data, body)
		if err != nil {
			log.Println("服务端json格式解析有误，非法格式:", err)
		} else {
			context := &WsContext{property: make(map[string]interface{})}

			// 获取到前端传递的数据了，拿上这些数据 去具体的业务进行处理
			req := &WsMsgReq{Conn: w, Body: body, Context: context}
			rsp := &WsMsgRsp{Body: &RspBody{Name: body.Name, Seq: req.Body.Seq}}
			if req.Body.Name == HeartbeatMsg {
				h := &Heartbeat{}
				// body.Msg 中的数据解码并映射到结构体 h 中
				mapstructure.Decode(body.Msg, h)
				// 获取当前时间并将其转换为毫秒级别的时间戳
				h.STime = time.Now().UnixNano() / 1e6
				rsp.Body.Msg = h
			} else {
				if w.router != nil {
					log.Println("req", req)
					w.router.Run(req, rsp)
				}
			}
			w.outChan <- rsp
		}
	}
	//退出循环直接关闭连接
	w.Close()
}

func (w *wsServer) Close() {
	_ = w.wsConn.Close()
}

func (w *wsServer) Write(body interface{}) {
	fmt.Println("写给客户端数据", body)
	// 序列化
	data, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
	}
	//2.前端消息是加密消息 需要进行加密
	secretKey, err := w.GetProperty("secretKey")
	if err == nil {
		//有加密
		key := secretKey.(string)
		//数据做加密
		data, _ = utils.AesCBCEncrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
	}
	//压缩
	if data, err := utils.Zip(data); err == nil {
		err := w.wsConn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			log.Println("写数据出错", err)
			return
		}
	}
}

func (w *wsServer) Handshake() {
	// 获取加密信息
	secretKey := ""
	key, err := w.GetProperty("secretKey")
	fmt.Println("进行加密连接", secretKey)
	if err == nil {
		secretKey = key.(string)
	} else {
		// 若出错，则生成随机16位随机字符串
		secretKey = utils.RandSeq(16)
	}
	handshake := &Handshake{
		secretKey,
	}
	body := &RspBody{Name: HandshakeMsg, Msg: handshake}

	if data, err := json.Marshal(body); err == nil {
		// 若密钥不为空，则需要设置密钥
		if secretKey != "" {
			w.SetProperty("secretKey", secretKey)
		} else {
			w.RemoveProperty("secretKey")
		}
		if data, err := utils.Zip(data); err == nil { //不报错
			err := w.wsConn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				return
			}
		}
	}
}
