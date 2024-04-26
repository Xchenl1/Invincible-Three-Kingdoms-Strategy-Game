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

var GeneralHandler = &generalHandler{}

type generalHandler struct {
}

func (gh *generalHandler) InitRouter(r *net.Router) {
	g := r.Group("general")
	g.Use(middleware.Log())
	g.AddRouter("myGenerals", gh.myGenerals, middleware.CheckRole())
	g.AddRouter("drawGeneral", gh.drawGeneral, middleware.CheckRole())
}
func (gh *generalHandler) myGenerals(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.MyGeneralRsp{}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name

	role, _ := req.Conn.GetProperty("role")
	r := role.(*data.Role)
	gs, err := logic.DefaultGeneralService.GetGenerals(r.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.Generals = gs
	rsp.Body.Msg = rspObj
}
func (r *generalHandler) drawGeneral(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//1. 计算抽卡花费的金钱
	//2. 判断金钱是否足够
	//3. 抽卡的次数 + 已有的武将 卡池是否足够
	//4. 随机生成武将即可（之前有实现）
	//5. 金币的扣除
	reqObj := &model.DrawGeneralReq{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj := &model.DrawGeneralRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj

	role, _ := req.Conn.GetProperty("role")
	rid := role.(*data.Role).RId

	// 计算抽卡所需要的金钱是否足够
	cost := gameConfig.Base.General.DrawGeneralCost * reqObj.DrawTimes
	if !logic.RoleResService.IsEnoughGold(rid, cost) {
		rsp.Body.Code = constant.GoldNotEnough
		return
	}
	limit := gameConfig.Base.General.Limit

	// 抽武将卡
	gs, err := logic.DefaultGeneralService.GetGenerals(rid)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	if len(gs)+reqObj.DrawTimes > limit {
		rsp.Body.Code = constant.OutGeneralLimit
		return
	}
	mgs := logic.DefaultGeneralService.Draw(rid, reqObj.DrawTimes)
	logic.RoleResService.CostGold(rid, cost)
	rspObj.Generals = mgs
}
