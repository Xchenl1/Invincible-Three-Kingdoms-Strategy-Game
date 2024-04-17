package data

import (
	"log"
	"sgserver/db"
	"sgserver/server/game/model"
)

type RoleRes struct {
	Id     int `xorm:"id pk autoincr"`
	RId    int `xorm:"rid"`
	Wood   int `xorm:"wood"`
	Iron   int `xorm:"iron"`
	Stone  int `xorm:"stone"`
	Grain  int `xorm:"grain"`
	Gold   int `xorm:"gold"`
	Decree int `xorm:"decree"` //令牌
}

type Yield struct {
	Wood  int
	Iron  int
	Stone int
	Grain int
	Gold  int
}

var RoleResDao = &roleResDao{
	resChan: make(chan *RoleRes, 100),
}

type roleResDao struct {
	resChan chan *RoleRes
}

func init() {
	go RoleResDao.running()
}

func (r *roleResDao) running() {
	for {
		select {
		case rr := <-r.resChan:
			if rr.Id > 0 {
				_, err := db.Engine.Table(rr).ID(rr.Id).Cols("wood", "iron", "stone",
					"grain", "gold", "decree").Update(rr)
				if err != nil {
					log.Println("roleResDao update error", err)
				}
			}
		}
	}
}
func (r *RoleRes) TableName() string {
	return "role_res"
}

func (r *RoleRes) ToModel() interface{} {
	p := model.RoleRes{}
	p.Gold = r.Gold
	p.Grain = r.Grain
	p.Stone = r.Stone
	p.Iron = r.Iron
	p.Wood = r.Wood
	p.Decree = r.Decree

	yield := GetYield(r.RId)
	p.GoldYield = yield.Gold
	p.GrainYield = yield.Grain
	p.StoneYield = yield.Stone
	p.IronYield = yield.Iron
	p.WoodYield = yield.Wood
	p.DepotCapacity = 10000
	return p
}

func (r *RoleRes) SyncExecute() {
	//通知修改
	RoleResDao.resChan <- r
}
