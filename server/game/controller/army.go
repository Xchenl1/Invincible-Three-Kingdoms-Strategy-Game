package controller

import (
	"github.com/mitchellh/mapstructure"
	"sgserver/constant"
	"sgserver/net"
	"sgserver/server/common"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/gameConfig/general"
	"sgserver/server/game/logic"
	"sgserver/server/game/middleware"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"time"
)

var ArmyHandler = &armyHandler{}

type armyHandler struct {
}

func (gh *armyHandler) InitRouter(r *net.Router) {
	g := r.Group("army")
	g.Use(middleware.Log())
	g.AddRouter("myList", gh.myList, middleware.CheckRole())
	g.AddRouter("dispose", gh.dispose, middleware.CheckRole())
	g.AddRouter("conscript", gh.conscript, middleware.CheckRole())
}

func (ah *armyHandler) myList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.ArmyListReq{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj := &model.ArmyListRsp{}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name

	role, _ := req.Conn.GetProperty("role")
	r := role.(*data.Role)
	arms, err := logic.DefaultArmyService.GetArmysByCity(r.RId, reqObj.CityId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.Armys = arms
	rspObj.CityId = reqObj.CityId
	rsp.Body.Msg = rspObj
}

// 武将上阵
func (gh *armyHandler) dispose(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.DisposeReq{}
	rspObj := &model.DisposeRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)

	// 参数判断
	if reqObj.Position < -1 || reqObj.Position > 2 || reqObj.Order > 5 {
		rsp.Body.Code = constant.InvalidParam
		return
	}
	// 城市判断
	rc, ok := logic.Default.Get(reqObj.CityId)
	if !ok {
		rsp.Body.Code = constant.CityNotExist
		return
	}
	if role.RId != rc.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	level := logic.CityFacilityService.GetFaciltyLevel(reqObj.CityId, gameConfig.JiaoChang)
	if level <= 0 || reqObj.Order > level {
		rsp.Body.Code = constant.ArmyNotEnough
		return
	}

	//查询武将是否存在
	newGen, ok := logic.DefaultGeneralService.Get(reqObj.GeneralId)
	if !ok {
		rsp.Body.Code = constant.GeneralNotFound
		return
	}
	// 武将是否是当前角色
	if newGen.RId != role.RId {
		rsp.Body.Code = constant.GeneralNotMe
		return
	}
	//查询军队，没有就创建
	army, ok := logic.DefaultArmyService.GetCreate(reqObj.CityId, role.RId, reqObj.Order)
	// 判断将军是否在城外
	if (army.FromX > 0 && army.FromX != rc.X) || (army.FromY > 0 && army.FromY != rc.Y) {
		rsp.Body.Code = constant.ArmyIsOutside
		return
	}
	// 上下阵
	if reqObj.Position == -1 {
		// 下阵
		for position, g := range army.Gens {
			if g == nil && g.Id == reqObj.GeneralId {
				// 检测武将是否在征兵中
				if !army.PositionCanModify(position) {
					rsp.Body.Code = constant.GeneralBusy
					return
				}
				army.GeneralArray[position] = 0
				army.SoldierArray[position] = 0
				army.Gens[position] = nil
				army.SyncExecute()
			}
		}
		newGen.CityId = 0
		newGen.Order = 0
		newGen.SyncExecute()
	} else {
		//上阵
		// 检测武将是否在征兵中
		if !army.PositionCanModify(reqObj.Position) {
			rsp.Body.Code = constant.GeneralBusy
			return
		}
		if newGen.CityId != 0 {
			rsp.Body.Code = constant.GeneralBusy
			return
		}

		if logic.DefaultArmyService.IsRepeat(role.RId, newGen.CfgId) {
			rsp.Body.Code = constant.GeneralBusy
			return
		}

		//判断是否能配前锋
		level := logic.CityFacilityService.GetFaciltyLevel(rc.CityId, gameConfig.TongShuaiTing)
		if reqObj.Position == 2 && (level < reqObj.Order) {
			rsp.Body.Code = constant.TongShuaiNotEnough
			return
		}

		//判断cost
		cost := general.General.Cost(newGen.CfgId)

		for _, g := range army.Gens {
			if g != nil {
				cost += general.General.Cost(newGen.CfgId)
			}
		}
		cityCost := logic.Default.GetCityCost(reqObj.CityId)

		if cityCost < cost {
			rsp.Body.Code = constant.CostNotEnough
			return
		}

		oldG := army.Gens[reqObj.Position]
		if oldG != nil {
			//旧的下阵
			oldG.CityId = 0
			oldG.Order = 0
			oldG.SyncExecute()
		}

		army.GeneralArray[reqObj.Position] = reqObj.GeneralId
		army.SoldierArray[reqObj.Position] = 0
		army.Gens[reqObj.Position] = newGen

		newGen.Order = reqObj.Order
		newGen.CityId = reqObj.CityId
		newGen.SyncExecute()
	}
	army.FromX = rc.X
	army.FromY = rc.Y
	army.SyncExecute()

	//队伍
	rspObj.Army = army.ToModel().(model.Army)
}

// 征兵
func (gh *armyHandler) conscript(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqobj := &model.ConscriptReq{}
	mapstructure.Decode(req.Body.Msg, reqobj)
	rspobj := &model.ConscriptRsp{}
	rsp.Body.Msg = rspobj
	rsp.Body.Code = constant.OK

	// 征兵 更新征兵的数量和征兵的完成时间，以及状态
	// 判断逻辑  征兵是否可以进行 资源是否充足 参数是否正常 募兵所的设施 等级

	// 参数是否合法
	if reqobj.ArmyId <= 0 || len(reqobj.Cnts) != 3 {
		rsp.Body.Code = constant.InvalidParam
		return
	}
	if reqobj.Cnts[0] < 0 || reqobj.Cnts[1] < 0 || reqobj.Cnts[2] < 0 {
		rsp.Body.Code = constant.InvalidParam
		return
	}
	// 角色
	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	// 军队是否存在
	army := logic.DefaultArmyService.Get(reqobj.ArmyId)
	if army == nil {
		rsp.Body.Code = constant.ArmyNotFound
		return
	}
	if role.RId != army.RId {
		rsp.Body.Code = constant.ArmyNotMe
		return
	}
	// 募兵所
	level := logic.CityFacilityService.GetFaciltyLevel(army.CityId, gameConfig.MBS)
	if level <= 0 {
		rsp.Body.Code = constant.BuildMBSNotFound
		return
	}

	for pos, y := range army.Gens {
		if reqobj.Cnts[pos] > 0 {
			if y == nil {
				rsp.Body.Code = constant.InvalidParam
				return
			}
			if !army.PositionCanModify(pos) {
				rsp.Body.Code = constant.GeneralBusy
				return
			}
		}
	}
	// 判断征兵数量是否合法 判断资源是否合法
	for pos, y := range army.Gens {
		if y == nil {
			continue
		}
		lv := y.Level
		gLevel := general.GeneralBasic.GetLevel(lv)
		add := logic.CityFacilityService.GetSoldier(army.CityId)
		if gLevel.Soldiers+add < reqobj.Cnts[pos]+army.SoldierArray[pos] {
			rsp.Body.Code = constant.InvalidParam
			return
		}
		var cur = time.Now().Unix()
		army.ConscriptCntArray[pos] = reqobj.Cnts[pos]
		army.ConscriptTimeArray[pos] = int64(gameConfig.Base.ConScript.CostTime*reqobj.Cnts[pos]) + cur - 2
	}

	var total int
	for _, v := range reqobj.Cnts {
		total += v
	}
	// 资源是否合法
	needRes := gameConfig.NeedRes{
		Decree: 0,
		Grain:  gameConfig.Base.ConScript.CostGrain * total,
		Wood:   gameConfig.Base.ConScript.CostWood * total,
		Iron:   gameConfig.Base.ConScript.CostIron * total,
		Stone:  gameConfig.Base.ConScript.CostStone * total,
		Gold:   gameConfig.Base.ConScript.CostGold * total,
	}
	ok := logic.RoleResService.TryUseNeed(role.RId, needRes)
	if !ok {
		rsp.Body.Code = constant.ResNotEnough
		return
	}
	// 可以更新
	army.Cmd = data.ArmyCmdConscript
	army.SyncExecute()
	rspobj.Army = army.ToModel().(model.Army)

	res := logic.RoleResService.Get(role.RId)
	if res != nil {
		rspobj.RoleRes = res.ToModel().(model.RoleRes)
	}
}
