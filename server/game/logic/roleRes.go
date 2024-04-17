package logic

import (
	"log"
	"sgserver/db"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/model/data"
	"time"
)

var RoleResService = &roleResService{
	rolesRes: make(map[int]*data.RoleRes),
}

type roleResService struct {
	rolesRes map[int]*data.RoleRes
}

func (r *roleResService) Load() {
	rr := make([]*data.RoleRes, 0)
	err := db.Engine.Find(&rr)
	if err != nil {
		log.Println(" load role_res table error")
	}

	for _, v := range rr {
		r.rolesRes[v.RId] = v
	}

	go r.produce()
}

func (r *roleResService) Get(rid int) *data.RoleRes {
	ra := &data.RoleRes{}
	ok, err := db.Engine.Table(ra).Where("rid=?", rid).Get(ra)
	if err != nil {
		log.Println(err)
		return nil
	}
	if ok {
		return ra
	}
	return nil
}

func (r *roleResService) GetYield(rid int) data.Yield {
	//产量 建筑 城市固定收益 = 最终的产量

	rbYield := DefaultRoleBuildService.GetYield(rid)
	rcYield := CityFacilityService.GetYield(rid)

	var y data.Yield

	y.Gold = rbYield.Gold + rcYield.Gold + gameConfig.Base.Role.GoldYield
	y.Stone = rbYield.Stone + rcYield.Stone + gameConfig.Base.Role.StoneYield
	y.Iron = rbYield.Iron + rcYield.Iron + gameConfig.Base.Role.IronYield
	y.Grain = rbYield.Grain + rcYield.Grain + gameConfig.Base.Role.GrainYield
	y.Wood = rbYield.Wood + rcYield.Wood + gameConfig.Base.Role.WoodYield

	return y
}

func (r *roleResService) IsEnoughGold(rid int, cost int) bool {
	rr := r.Get(rid)
	return rr.Gold >= cost
}

func (r *roleResService) CostGold(rid int, cost int) {
	rr := r.Get(rid)
	if rr.Gold >= cost {
		rr.Gold -= cost
		rr.SyncExecute()
	}
}

func (r *roleResService) TryUseNeed(rid int, need gameConfig.NeedRes) bool {

	rr := r.Get(rid)
	if need.Decree <= rr.Decree && need.Grain <= rr.Grain &&
		need.Stone <= rr.Stone && need.Wood <= rr.Wood &&
		need.Iron <= rr.Iron && need.Gold <= rr.Gold {
		rr.Decree -= need.Decree
		rr.Iron -= need.Iron
		rr.Wood -= need.Wood
		rr.Stone -= need.Stone
		rr.Grain -= need.Grain
		rr.Gold -= need.Gold

		rr.SyncExecute()
		return true
	} else {
		return false
	}
}

func (r *roleResService) produce() {
	for {
		// 获取产量 隔一段时间获取一次
		recovertime := gameConfig.Base.Role.RecoveryTime
		time.Sleep(time.Second * time.Duration(recovertime))

		for _, v := range r.rolesRes {
			capacity := GetDepotCapacity(v.RId)
			yield := r.GetYield(v.RId)
			if v.Wood < capacity {
				// 按照实际需求来
				v.Wood += yield.Wood / 6
			}
			if v.Stone < capacity {
				// 按照实际需求来
				v.Stone += yield.Wood / 6
			}
			if v.Iron < capacity {
				// 按照实际需求来
				v.Iron += yield.Wood / 6
			}
			if v.Gold < capacity {
				// 按照实际需求来
				v.Gold += yield.Wood / 6
			}
			if v.Grain < capacity {
				// 按照实际需求来
				v.Grain += yield.Wood / 6
			}

		}
	}
}

func GetDepotCapacity(rid int) int {
	return CityFacilityService.GetCapacity(rid) + gameConfig.Base.Role.DepotCapacity
}
