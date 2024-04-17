package logic

import (
	"log"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/server/common"
	"sgserver/server/game/global"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"sgserver/utils"
	"sync"
)

var DefaultArmyService = &ArmyService{
	passByPosArmys: make(map[int]map[int]*data.Army),
}

type ArmyService struct {
	passBy         sync.RWMutex
	passByPosArmys map[int]map[int]*data.Army //玩家路过位置的军队 key:posId,armyId
}

func (g *ArmyService) GetArmys(rid int) ([]model.Army, error) {
	mrs := make([]data.Army, 0)
	mr := &data.Army{}
	err := db.Engine.Table(mr).Where("rid=?", rid).Find(&mrs)
	if err != nil {
		log.Println("军队查询出错", err)
		return nil, common.New(constant.DBError, "军队查询出错")
	}
	modelMrs := make([]model.Army, 0)
	for _, v := range mrs {
		modelMrs = append(modelMrs, v.ToModel().(model.Army))
	}
	return modelMrs, nil
}

func (g *ArmyService) GetArmysByCity(rid, cid int) ([]model.Army, error) {
	mrs := make([]data.Army, 0)
	mr := &data.Army{}
	err := db.Engine.Table(mr).Where("rid=? and cityId=?", rid, cid).Find(&mrs)
	if err != nil {
		log.Println("军队查询出错", err)
		return nil, common.New(constant.DBError, "军队查询出错")
	}
	modelMrs := make([]model.Army, 0)
	for _, v := range mrs {
		modelMrs = append(modelMrs, v.ToModel().(model.Army))
	}
	return modelMrs, nil
}

func (a *ArmyService) ScanBlock(roleId int, req *model.ScanBlockReq) ([]model.Army, error) {
	x := req.X
	y := req.Y
	length := req.Length
	out := make([]model.Army, 0)
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return out, nil
	}

	maxX := utils.MinInt(global.MapWith, x+length-1)
	maxY := utils.MinInt(global.MapHeight, y+length-1)

	a.passBy.RLock()
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {

			posId := global.ToPosition(i, j)
			armys, ok := a.passByPosArmys[posId]
			if ok {
				//是否在视野范围内
				is := armyIsInView(roleId, i, j)
				if is == false {
					continue
				}
				for _, army := range armys {
					out = append(out, army.ToModel().(model.Army))
				}
			}
		}
	}
	a.passBy.RUnlock()
	return out, nil
}

func (g *ArmyService) GetArmysByCityAndOrder(rid, cid int, order int8) (*data.Army, bool) {
	armys := make([]*data.Army, 0)
	mr := &data.Army{}
	err := db.Engine.Table(mr).Where("rid=? and cityId=?", rid, cid).Find(&armys)
	if err != nil {
		log.Println("军队查询出错", err)
		return nil, false
	}
	for _, v := range armys {
		if v.Order == order {
			g.updateGenerals(v)
			return v, true
		}
	}

	return nil, false
}

func (a *ArmyService) updateGenerals(armys ...*data.Army) {
	for _, army := range armys {
		army.Gens = make([]*data.General, 0)
		for _, gid := range army.GeneralArray {
			if gid == 0 {
				army.Gens = append(army.Gens, nil)
			} else {
				//查询武将
				g, _ := DefaultGeneralService.Get(gid)
				army.Gens = append(army.Gens, g)
			}
		}
	}
}

func (a *ArmyService) GetCreate(cid int, rid int, order int8) (*data.Army, bool) {
	army, ok := a.GetArmysByCityAndOrder(rid, cid, order)
	if ok {
		return army, true
	}
	//需要创建
	army = &data.Army{RId: rid,
		Order:              order,
		CityId:             cid,
		Generals:           `[0,0,0]`,
		Soldiers:           `[0,0,0]`,
		GeneralArray:       []int{0, 0, 0},
		SoldierArray:       []int{0, 0, 0},
		ConscriptCnts:      `[0,0,0]`,
		ConscriptTimes:     `[0,0,0]`,
		ConscriptCntArray:  []int{0, 0, 0},
		ConscriptTimeArray: []int64{0, 0, 0},
	}
	a.updateGenerals(army)

	_, err := db.Engine.Table(army).Insert(army)
	if err != nil {
		log.Println("armyServer GetCreate err", err)
		return nil, false
	}
	return army, true
}

func (g *ArmyService) GetDbArmys(rid int) ([]*data.Army, error) {
	mrs := make([]*data.Army, 0)
	mr := &data.Army{}
	err := db.Engine.Table(mr).Where("rid=?", rid).Find(&mrs)
	if err != nil {
		log.Println("军队查询出错", err)
		return nil, common.New(constant.DBError, "军队查询出错")
	}
	modelMrs := make([]model.Army, 0)
	for _, v := range mrs {
		modelMrs = append(modelMrs, v.ToModel().(model.Army))
	}
	return mrs, nil
}

func (g *ArmyService) IsRepeat(rid int, cfgId int) bool {
	armys, err := g.GetDbArmys(rid)
	if err != nil {
		return false
	}
	for _, v := range armys {
		for _, gid := range v.GeneralArray {
			if gid == cfgId {
				return true
			}
		}
	}
	return false
}

func armyIsInView(rid, x, y int) bool {
	//简单点 先设为true
	return true
}
