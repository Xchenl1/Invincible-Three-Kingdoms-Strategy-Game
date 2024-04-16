package net

import (
	"encoding/json"
	"errors"
	"github.com/forgoer/openssl"
	"github.com/gorilla/websocket"
	"log"
	"sgserver/utils"
	"sync"
)

// websocket服务
type wsServer struct {
	wsConn       *websocket.Conn        //websocket连接
	router       *Router                //路由
	outChan      chan *WsMsgRsp         //写队列
	Seq          int64                  //状态信息
	propertyLock sync.RWMutex           //读写锁
	property     map[string]interface{} //存储一些信息
}

func NewWsServer(wsConn *websocket.Conn) *wsServer {
	return &wsServer{
		wsConn:   wsConn,
		outChan:  make(chan *WsMsgRsp, 1000),
		property: make(map[string]interface{}),
		Seq:      0,
	}
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

func (w *wsServer) Push(name string, data interface{}) {
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

func (w *wsServer) writeMsgLoop() {
	for {
		select {
		case msg := <-w.outChan:
			w.Write(msg)
		}
	}
}

func (w *wsServer) readMsgLoop() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(err)
			w.Close()
		}
	}()
	for {
		_, data, err := w.wsConn.ReadMessage()
		if err != nil {
			log.Println("收消息出现错误！", err)
			break
		}
		//收到消息会解析消息
		//1.解压unzip
		data, err = utils.UnZip(data)
		if err != nil {
			log.Println("解压失败！，格式不合法！")
			continue
		}
		//2.前端消息是加密消息 需要进行解密
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
		//3.data转为body
		body := &ReqBody{}
		err = json.Unmarshal(data, body)
		if err != nil {
			log.Println("数据格式有误！非法格式", err)
		} else {
			//拿到前端的代码
			req := &WsMsgReq{Conn: w, Body: body}
			//返回前端的代码
			rsp := &WsMsgRsp{Body: &RspBody{Name: body.Name, Seq: req.Body.Seq}}
			w.router.Run(req, rsp)
			w.outChan <- rsp
		}
	}
	//退出循环直接关闭连接
	w.Close()
}

func (w *wsServer) Close() {
	_ = w.wsConn.Close()
}

func (w *wsServer) Write(msg *WsMsgRsp) {
	data, err := json.Marshal(msg.Body)
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
		w.wsConn.WriteMessage(websocket.BinaryMessage, data)
	}
}

func (w *wsServer) Handshake() {
	secretKey := ""
	key, err := w.GetProperty("secretKey")
	if err == nil {
		secretKey = key.(string)
	} else {
		secretKey = utils.RandSeq(16)
	}
	handshake := &Handshake{
		secretKey,
	}
	body := &RspBody{Name: HandshakeMsg, Msg: handshake}

	if data, err := json.Marshal(body); err == nil {
		if secretKey != "" {
			w.SetProperty("secretKey", secretKey)
		} else {
			w.RemoveProperty("secretKey")
		}
		if data, err := utils.Zip(data); err == nil { //不报错
			w.wsConn.WriteMessage(websocket.BinaryMessage, data)
		}
	}
}
