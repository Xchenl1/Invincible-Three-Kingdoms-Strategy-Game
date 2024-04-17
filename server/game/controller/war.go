package controller

import (
	"sgserver/constant"
	"sgserver/net"
	"sgserver/server/common"
	"sgserver/server/game/logic"
	"sgserver/server/game/middleware"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
)

var WarHandler = &warHandler{}

type warHandler struct {
}

func (w *warHandler) InitRouter(r *net.Router) {
	g := r.Group("war")
	g.Use(middleware.Log())
	g.AddRouter("report", w.report, middleware.CheckRole())
}

func (w *warHandler) report(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.WarReportRsp{}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name

	role, err := req.Conn.GetProperty("role")
	if err != nil {
		rsp.Body.Code = constant.SessionInvalid
		return
	}
	r := role.(*data.Role)

	wReports, err := logic.DefaultWarService.GetWarReports(r.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.List = wReports
	rsp.Body.Msg = rspObj
}
