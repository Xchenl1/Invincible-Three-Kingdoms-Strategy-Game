package logic

import (
	"encoding/json"
	"log"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/server/common"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/gameConfig/general"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"time"
)

var DefaultGeneralService = &GeneralService{}

type GeneralService struct {
}

func (g *GeneralService) GetGenerals(rid int) ([]model.General, error) {
	mrs := make([]*data.General, 0)
	mr := &data.General{}
	err := db.Engine.Table(mr).Where("rid=?", rid).Find(&mrs)
	if err != nil {
		log.Println("武将查询出错", err)
		return nil, common.New(constant.DBError, "武将查询出错")
	}
	if len(mrs) <= 0 {
		//随机三个武将 做为初始武将
		var count = 0
		for {
			if count >= 3 {
				break
			}
			cfgId := general.General.Rand()
			if cfgId != 0 {
				gen, err := g.NewGeneral(cfgId, rid, 1)
				if err != nil {
					log.Println("生成武将出错", err)
					continue
				}
				mrs = append(mrs, gen)
				count++
			}
		}

	}
	modelMrs := make([]model.General, 0)
	for _, v := range mrs {
		modelMrs = append(modelMrs, v.ToModel().(model.General))
	}
	return modelMrs, nil
}

const (
	GeneralNormal      = 0 //正常
	GeneralComposeStar = 1 //星级合成
	GeneralConvert     = 2 //转换
)

func (g *GeneralService) NewGeneral(cfgId int, rid int, level int8) (*data.General, error) {
	cfg := general.General.GMap[cfgId]
	sa := make([]*model.GSkill, 3)
	ss, _ := json.Marshal(sa)
	ge := &data.General{
		PhysicalPower: gameConfig.Base.General.PhysicalPowerLimit,
		RId:           rid,
		CfgId:         cfg.CfgId,
		Order:         0,
		CityId:        0,
		Level:         level,
		CreatedAt:     time.Now(),
		CurArms:       cfg.Arms[0],
		HasPrPoint:    0,
		UsePrPoint:    0,
		AttackDis:     0,
		ForceAdded:    0,
		StrategyAdded: 0,
		DefenseAdded:  0,
		SpeedAdded:    0,
		DestroyAdded:  0,
		Star:          cfg.Star,
		StarLv:        0,
		ParentId:      0,
		SkillsArray:   sa,
		Skills:        string(ss),
		State:         GeneralNormal,
	}
	_, err := db.Engine.Table(ge).Insert(ge)
	if err != nil {
		return nil, err
	}
	return ge, nil
}
func (g *GeneralService) Draw(rid int, nums int) []model.General {
	mrs := make([]*data.General, 0)
	for i := 0; i < nums; i++ {
		cfgId := general.General.Rand()
		gen, _ := g.NewGeneral(cfgId, rid, 1)
		mrs = append(mrs, gen)
	}
	modelMrs := make([]model.General, 0)
	for _, v := range mrs {
		modelMrs = append(modelMrs, v.ToModel().(model.General))
	}
	return modelMrs
}

func (g *GeneralService) Get(id int) (*data.General, bool) {
	mr := &data.General{}
	ok, err := db.Engine.Table(mr).Where("id=?", id).Get(mr)
	if err != nil {
		log.Println("武将查询出错", err)
		return nil, false
	}
	if ok {
		return mr, true
	}
	return nil, false
}
