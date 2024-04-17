package controller

import (
	"github.com/mitchellh/mapstructure"
	"sgserver/constant"
	"sgserver/net"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/logic"
	"sgserver/server/game/middleware"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"time"
)

var InteriorController = &interiorController{}

type interiorController struct {
}

// Router 征收
func (i *interiorController) Router(router *net.Router) {
	g := router.Group("interior")
	g.Use(middleware.Log())
	g.AddRouter("openCollect", i.openCollect, middleware.CheckRole())
	g.AddRouter("collect", i.collect, middleware.CheckRole())
	g.AddRouter("transform", i.transform, middleware.CheckRole())
}

func (i *interiorController) openCollect(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.OpenCollectionRsp{}

	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	roleAttr := logic.DefaultRoleAttrService.Get(role.RId)
	if roleAttr != nil {
		interval := gameConfig.Base.Role.CollectInterval
		timeLimit := gameConfig.Base.Role.CollectTimesLimit
		rspObj.Limit = timeLimit
		rspObj.CurTimes = roleAttr.CollectTimes

		if roleAttr.LastCollectTime.IsZero() {
			rspObj.NextTime = 0
		} else {
			if roleAttr.CollectTimes >= timeLimit {
				y, m, d := roleAttr.LastCollectTime.Add(24 * time.Hour).Date()
				nextTime := time.Date(y, m, d, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
				rspObj.NextTime = nextTime.UnixNano() / 1e6
			} else {
				nextTime := roleAttr.LastCollectTime.Add(time.Duration(interval) * time.Second)
				rspObj.NextTime = nextTime.UnixNano() / 1e6
			}
		}
	}
}

func (i *interiorController) collect(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	// 查询角色资源，得到金币
	rspObj := &model.CollectionRsp{}

	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	ra := logic.DefaultRoleAttrService.Get(role.RId)
	if ra == nil {
		rsp.Body.Code = constant.DBError
		return
	}
	rs := logic.RoleResService.Get(role.RId)
	if rs == nil {
		rsp.Body.Code = constant.DBError
		return
	}

	curTime := time.Now()
	lastTime := ra.LastCollectTime
	if curTime.YearDay() != lastTime.YearDay() || curTime.Year() != lastTime.Year() {
		ra.CollectTimes = 0
		ra.LastCollectTime = time.Time{}
	}

	timeLimit := gameConfig.Base.Role.CollectTimesLimit
	//是否超过征收次数上限
	if ra.CollectTimes >= timeLimit {
		rsp.Body.Code = constant.OutCollectTimesLimit
		return
	}

	//cd内不能操作
	need := lastTime.Add(time.Duration(gameConfig.Base.Role.CollectTimesLimit) * time.Second)
	if curTime.Before(need) {
		rsp.Body.Code = constant.InCdCanNotOperate
		return
	}

	gold := logic.RoleResService.GetYield(rs.RId).Gold
	rspObj.Gold = gold
	rs.Gold += gold

	rs.SyncExecute()

	ra.LastCollectTime = curTime
	ra.CollectTimes += 1
	ra.SyncExecute()

	interval := gameConfig.Base.Role.CollectInterval
	if ra.CollectTimes >= timeLimit {
		y, m, d := ra.LastCollectTime.Add(24 * time.Hour).Date()
		nextTime := time.Date(y, m, d, 0, 0, 0, 0, time.FixedZone("IST", 3600))
		rspObj.NextTime = nextTime.UnixNano() / 1e6
	} else {
		nextTime := ra.LastCollectTime.Add(time.Duration(interval) * time.Second)
		rspObj.NextTime = nextTime.UnixNano() / 1e6
	}

	rspObj.CurTimes = ra.CollectTimes
	rspObj.Limit = timeLimit
}

func (i *interiorController) transform(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.TransformReq{}
	rspObj := &model.TransformRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	// 查询资源
	roleRes := logic.RoleResService.Get(role.RId)
	if roleRes == nil {
		rsp.Body.Code = constant.DBError
		return
	}

	rc := logic.Default.GetMainCity(role.RId)
	if rc == nil {
		rsp.Body.Code = constant.DBError
		return
	}
	// 在主城做交易
	level := logic.CityFacilityService.GetFaciltyLevel(rc.CityId, gameConfig.JiShi)
	if level <= 0 {
		rsp.Body.Code = constant.NotHasJiShi
		return
	}
	roleRes.Wood -= reqObj.From[0]
	roleRes.Wood += reqObj.To[0]
	roleRes.Iron -= reqObj.From[1]
	roleRes.Iron += reqObj.To[1]
	roleRes.Stone -= reqObj.From[2]
	roleRes.Stone += reqObj.To[2]
	roleRes.Grain -= reqObj.From[3]
	roleRes.Grain += reqObj.To[3]
	roleRes.SyncExecute()
}
