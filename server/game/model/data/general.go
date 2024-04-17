package data

import (
	"log"
	"sgserver/db"
	"sgserver/server/game/model"
	"time"
)

const (
	GeneralNormal      = 0 //正常
	GeneralComposeStar = 1 //星级合成
	GeneralConvert     = 2 //转换
)
const SkillLimit = 3

type General struct {
	Id            int             `xorm:"id pk autoincr"`
	RId           int             `xorm:"rid"`
	CfgId         int             `xorm:"cfgId"`
	PhysicalPower int             `xorm:"physical_power"`
	Level         int8            `xorm:"level"`
	Exp           int             `xorm:"exp"`
	Order         int8            `xorm:"order"`
	CityId        int             `xorm:"cityId"`
	CreatedAt     time.Time       `xorm:"created_at"`
	CurArms       int             `xorm:"arms"`
	HasPrPoint    int             `xorm:"has_pr_point"`
	UsePrPoint    int             `xorm:"use_pr_point"`
	AttackDis     int             `xorm:"attack_distance"`
	ForceAdded    int             `xorm:"force_added"`
	StrategyAdded int             `xorm:"strategy_added"`
	DefenseAdded  int             `xorm:"defense_added"`
	SpeedAdded    int             `xorm:"speed_added"`
	DestroyAdded  int             `xorm:"destroy_added"`
	StarLv        int8            `xorm:"star_lv"`
	Star          int8            `xorm:"star"`
	ParentId      int             `xorm:"parentId"`
	Skills        string          `xorm:"skills"`
	SkillsArray   []*model.GSkill `xorm:"-"`
	State         int8            `xorm:"state"`
}

var GeneralDao = &generalDao{
	genChan: make(chan *General, 100),
}

type generalDao struct {
	genChan chan *General
}

func (g generalDao) running() {
	for {
		select {
		case gen := <-g.genChan:
			if gen.Id < 0 {
				_, err := db.Engine.Table(gen).ID(gen.Id).Cols("").Update(gen)
				if err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}

func (g *General) TableName() string {
	return "general"
}

func (g *General) ToModel() interface{} {
	p := model.General{}
	p.CityId = g.CityId
	p.Order = g.Order
	p.PhysicalPower = g.PhysicalPower
	p.Id = g.Id
	p.CfgId = g.CfgId
	p.Level = g.Level
	p.Exp = g.Exp
	p.CurArms = g.CurArms
	p.HasPrPoint = g.HasPrPoint
	p.UsePrPoint = g.UsePrPoint
	p.AttackDis = g.AttackDis
	p.ForceAdded = g.ForceAdded
	p.StrategyAdded = g.StrategyAdded
	p.DefenseAdded = g.DefenseAdded
	p.SpeedAdded = g.SpeedAdded
	p.DestroyAdded = g.DestroyAdded
	p.StarLv = g.StarLv
	p.Star = g.Star
	p.State = g.State
	p.ParentId = g.ParentId
	p.Skills = g.SkillsArray
	return p
}

func (g *General) SyncExecute() {
	GeneralDao.genChan <- g
}
func init() {
	go GeneralDao.running()
}
