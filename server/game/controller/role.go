package controller

import (
	"github.com/mitchellh/mapstructure"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/net"
	"sgserver/server/common"
	"sgserver/server/game/logic"
	"sgserver/server/game/middleware"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"sgserver/utils"
	"time"
)

var DefaultRoleHandler = &RoleHandler{}

type RoleHandler struct {
}

func (rh *RoleHandler) InitRouter(r *net.Router) {
	g := r.Group("role")
	g.Use(middleware.Log())
	g.AddRouter("create", rh.create)
	g.AddRouter("enterServer", rh.enterServer)
	g.AddRouter("myProperty", rh.myproperty, middleware.CheckRole())
	g.AddRouter("posTagList", rh.posTagList, middleware.CheckRole())
}

// 登录游戏
func (rh *RoleHandler) enterServer(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	// 验证 session 是否合法 取出登录用户的id 根据id 查询用户的游戏角色 如果有就继续 没有就处理错误
	//先验证session是否合法
	reqObj := &model.EnterServerReq{}
	rspObj := &model.EnterServerRsp{}

	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name

	//解析 session 数据到 reqobj
	err := mapstructure.Decode(req.Body.Msg, reqObj)
	if err != nil {
		rsp.Body.Code = constant.InvalidParam
		return
	}

	//解析session是否合法，用户登录是否过期
	token := reqObj.Session
	_, claim, err := utils.ParseToken(token)
	if err != nil {
		rsp.Body.Code = constant.SessionInvalid
		return
	}

	//用户id
	uid := claim.Uid
	err = logic.DefaultRoleService.EnterServer(uid, rspObj, req)
	if err != nil {
		rspObj.Time = time.Now().UnixNano() / 1e6
		rsp.Body.Msg = rspObj
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

}

// 我的财产
func (rh *RoleHandler) myproperty(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.MyRolePropertyReq{}
	rspObj := &model.MyRolePropertyRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	r, err := req.Conn.GetProperty("role")
	if err != nil {
		rsp.Body.Code = constant.SessionInvalid
		return
	}
	role := r.(*data.Role)
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	//城池
	rspObj.Citys, err = logic.Default.GetCitys(role.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	//建筑
	rspObj.MRBuilds, err = logic.DefaultRoleBuildService.GetBuilds(role.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	//资源
	rspObj.RoleRes, err = logic.DefaultRoleService.GetRoleRes(role.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	//武将
	rspObj.Generals, err = logic.DefaultGeneralService.GetGenerals(role.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	//军队
	rspObj.Armys, err = logic.DefaultArmyService.GetArmys(role.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rsp.Body.Msg = rspObj
}

func (rh *RoleHandler) posTagList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.PosTagListRsp{}
	role, err := req.Conn.GetProperty("role")
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	if err != nil {
		rsp.Body.Code = constant.InvalidParam
		return
	}
	r := role.(*data.Role)
	rspObj.PosTags = logic.DefaultRoleAttrService.GetPosTags(r.RId)
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj
}

func (rh *RoleHandler) create(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.CreateRoleReq{}
	rspObj := &model.CreateRoleRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)

	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	role := &data.Role{}
	//查询数据库中是否存在用户
	ok, err := db.Engine.Where("uid=?", reqObj.UId).Get(role)
	if err != nil {
		rsp.Body.Code = constant.DBError
		return
	}
	if ok {
		rsp.Body.Code = constant.RoleAlreadyCreate
		return
	}
	role.UId = reqObj.UId
	role.Sex = reqObj.Sex
	role.NickName = reqObj.NickName
	role.Balance = 0
	role.HeadId = reqObj.HeadId
	role.CreatedAt = time.Now()
	role.LoginTime = time.Now()
	_, err = db.Engine.InsertOne(role)
	if err != nil {
		rsp.Body.Code = constant.DBError
		return
	}
	rspObj.Role = role.ToModel().(model.Role)
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj
}
