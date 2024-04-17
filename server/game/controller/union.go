package controller

import (
	"github.com/mitchellh/mapstructure"
	"sgserver/constant"
	"sgserver/net"
	"sgserver/server/common"
	"sgserver/server/game/logic"
	"sgserver/server/game/middleware"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
)

var CoalitionController = &coalitionController{}

type coalitionController struct {
}

func (c *coalitionController) Router(router *net.Router) {
	g := router.Group("union")
	g.Use(middleware.Log())
	g.AddRouter("list", c.list, middleware.CheckRole())
	g.AddRouter("info", c.info, middleware.CheckRole())
	g.AddRouter("applyList", c.applyList, middleware.CheckRole())
}

func (c *coalitionController) list(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.ListRsp{}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	uns, err := logic.CoalitionService.List()
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.List = uns
	rsp.Body.Msg = rspObj
}

func (c *coalitionController) info(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.InfoReq{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj := &model.InfoRsp{}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	un, err := logic.CoalitionService.Get(reqObj.Id)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.Info = un
	rspObj.Id = reqObj.Id
	rsp.Body.Msg = rspObj
}

func (u *coalitionController) applyList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//根据联盟id 去查询申请列表，rid申请人，你角色表 查询详情即可
	// state 0 正在申请 1 拒绝 2 同意
	//什么人能看到申请列表 只有盟主和副盟主能看到申请列表
	reqObj := &model.ApplyReq{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj := &model.ApplyRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	//查询联盟
	un := logic.CoalitionService.GetCoalition(reqObj.Id)
	if un == nil {
		rsp.Body.Code = constant.DBError
		return
	}
	if un.Chairman != role.RId && un.ViceChairman != role.RId {
		rspObj.Id = reqObj.Id
		rspObj.Applys = make([]model.ApplyItem, 0)
		return
	}

	ais, err := logic.CoalitionService.GetListApply(reqObj.Id, 0)
	if err != nil {
		rsp.Body.Code = constant.DBError
		return
	}
	rspObj.Id = reqObj.Id
	rspObj.Applys = ais
}
