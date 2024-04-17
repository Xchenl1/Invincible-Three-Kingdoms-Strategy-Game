package controller

import (
	"github.com/mitchellh/mapstructure"
	"sgserver/constant"
	"sgserver/net"
	"sgserver/server/common"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/logic"
	"sgserver/server/game/middleware"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
)

var DefaultNationMapHandler = &NationMapHandler{}

type NationMapHandler struct {
}

func (n *NationMapHandler) InitRouter(r *net.Router) {
	g := r.Group("nationMap")
	g.Use(middleware.Log())
	g.AddRouter("config", n.config, middleware.CheckRole())
	g.AddRouter("scanBlock", n.scanBlock, middleware.CheckRole())
}

func (n *NationMapHandler) config(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.ConfigReq{}
	rspObj := &model.ConfigRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	m := gameConfig.MapBuildConf.Cfg
	rspObj.Confs = make([]model.Conf, len(m))
	for index, v := range m {
		rspObj.Confs[index].Type = v.Type
		rspObj.Confs[index].Name = v.Name
		rspObj.Confs[index].Level = v.Level
		rspObj.Confs[index].Defender = v.Defender
		rspObj.Confs[index].Durable = v.Durable
		rspObj.Confs[index].Grain = v.Grain
		rspObj.Confs[index].Iron = v.Iron
		rspObj.Confs[index].Stone = v.Stone
		rspObj.Confs[index].Wood = v.Wood
	}
}

func (n *NationMapHandler) scanBlock(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.ScanBlockReq{}
	rspObj := &model.ScanRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	//扫描建筑
	rm, err := logic.DefaultRoleBuildService.ScanBlock(reqObj)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.MRBuilds = rm
	//扫描城池
	rc, err := logic.Default.ScanBlock(reqObj)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.MCBuilds = rc

	//扫描军队
	armys, err := logic.DefaultArmyService.ScanBlock(role.RId, reqObj)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.Armys = armys
}
