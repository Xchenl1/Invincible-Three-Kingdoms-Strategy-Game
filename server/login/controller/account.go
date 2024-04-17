package controller

import (
	"github.com/mitchellh/mapstructure"
	"log"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/net"
	"sgserver/server/login/model"
	"sgserver/server/login/proto"
	"sgserver/server/models"
	"sgserver/utils"
	"time"
)

var DefaultAccount = &Account{}

type Account struct {
}

func (a *Account) Router(r *net.Router) {
	g := r.Group("account")
	g.AddRouter("login", a.login)
}

func (a *Account) login(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	/*
		1.用户名 密码 硬件id
		2.根据用户名查询user表得到数据 进行密码比对 如果密码正确 登录成功
		3.保存用户登录记录 保存用户最后一次登录信息
		4.需要session 使用jwt登录 生成一个加密算法
		5.判断用户是否合法
	*/
	loginReq := &proto.LoginReq{}
	loginRes := &proto.LoginRsp{}
	err := mapstructure.Decode(req.Body.Msg, loginReq) //解析requset到结构体中
	if err != nil {
		log.Println(err)
		return
	}
	user := &models.User{}
	ok, err := db.Engine.Table(user).Where("username=?", loginReq.Username).Get(user)
	if err != nil {
		log.Println("用户表查询错误！", err)
		return
	}
	if !ok {
		rsp.Body.Code = constant.UserNotExist
		return
	}
	pwd := utils.Password(loginReq.Password, user.Passcode)
	if pwd != user.Passwd {
		rsp.Body.Code = constant.PwdIncorrect
		return
	}
	//jwt加密 A.B.C A定义加密算法 B定义放入的数据  C部分根据A+B生成加密字符串
	token, _ := utils.Award(user.UId)
	rsp.Body.Code = constant.OK
	loginRes.Uid = user.UId
	loginRes.Username = user.Username
	loginRes.Session = token
	loginRes.Password = ""
	rsp.Body.Msg = loginRes

	//保存用户登录信息
	ul := &model.LoginHistory{
		UId: user.UId, CTime: time.Now(), Ip: loginReq.Ip,
		Hardware: loginReq.Hardware, State: model.Login,
	}
	db.Engine.Table(ul).Insert(ul)

	//最后一次登录记录
	ll := &model.LoginLast{}
	ok, _ = db.Engine.Table(ll).Where("uid=?", user.UId).Get(ll)
	if ok {
		//有数据 更新
		ll.IsLogout = 0
		ll.Ip = loginReq.Ip
		ll.LoginTime = time.Now()
		ll.Session = token
		ll.Hardware = loginReq.Hardware
		db.Engine.Table(ll).Update(ll)
	} else {
		ll.IsLogout = 0
		ll.Ip = loginReq.Ip
		ll.LoginTime = time.Now()
		ll.Session = token
		ll.Hardware = loginReq.Hardware
		ll.UId = user.UId
		_, err = db.Engine.Table(ll).Insert(ll)
		if err != nil {
			log.Println(err)
		}
	}
	//缓存一下 此用户和当前的ws连接 tishi
	net.Mgr.UserLogin(req.Conn, user.UId, token)
}
