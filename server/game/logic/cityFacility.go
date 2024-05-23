package logic

import (
	"encoding/json"
	"github.com/go-xorm/xorm"
	"log"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/net"
	"sgserver/server/common"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/model/data"
	"sync"
	"time"
)

var CityFacilityService = &cityFacilityService{
	facilities:      make(map[int]*data.CityFacility),
	facilitiesByRId: make(map[int][]*data.CityFacility),
}

type cityFacilityService struct {
	mutex           sync.RWMutex
	facilities      map[int]*data.CityFacility
	facilitiesByRId map[int][]*data.CityFacility //key:rid
}

func (c *cityFacilityService) TryCreate(cid, rid int, req *net.WsMsgReq) error {

	cf := &data.CityFacility{}
	ok, err := db.Engine.Table(cf).Where("cityId=?", rid).Get(cf)
	if err != nil {
		log.Println("查询城市设施出错", err)
		return common.New(constant.DBError, "数据库错误")
	}
	if ok {
		return nil
	}
	list := gameConfig.FacilityConf.List
	fs := make([]data.Facility, len(list))
	for index, v := range list {
		fac := data.Facility{
			Name:         v.Name,
			PrivateLevel: 0,
			Type:         v.Type,
			UpTime:       0,
		}
		fs[index] = fac
	}
	fsJson, err := json.Marshal(fs)
	if err != nil {
		return common.New(constant.DBError, "转json出错")
	}

	cf.RId = rid
	cf.CityId = cid
	cf.Facilities = string(fsJson)
	if session := req.Context.Get("dbSession"); session != nil {
		_, err = session.(*xorm.Session).Table(cf).Insert(cf)
	} else {
		_, err = db.Engine.Table(cf).Insert(cf)
	}

	if err != nil {
		log.Println("插入城市设施出错", err)
		return common.New(constant.DBError, "数据库错误")
	}
	return nil
}

func (c *cityFacilityService) GetYield(rid int) data.Yield {
	cfs, err := c.GetByRId(rid)
	var y data.Yield
	if err != nil {
		for _, cf := range cfs {
			for _, f := range cf.Facility() {
				if f.GetLevel() > 0 {
					values := gameConfig.FacilityConf.GetValues(f.Type, f.GetLevel())
					additions := gameConfig.FacilityConf.GetAdditions(f.Type)
					for i, aType := range additions {
						if aType == gameConfig.TypeWood {
							y.Wood += values[i]
						} else if aType == gameConfig.TypeGrain {
							y.Grain += values[i]
						} else if aType == gameConfig.TypeIron {
							y.Iron += values[i]
						} else if aType == gameConfig.TypeStone {
							y.Stone += values[i]
						} else if aType == gameConfig.TypeTax {
							y.Gold += values[i]
						}
					}
				}
			}
		}
		log.Println("cityFacilityService GetYield err", err)
	}
	return y
}

func (c *cityFacilityService) GetByRId(rid int) ([]*data.CityFacility, error) {
	cf := make([]*data.CityFacility, 0)
	err := db.Engine.Table(new(data.CityFacility)).Where("rid=?", rid).Find(&cf)
	if err != nil {
		log.Println(err)
		return cf, common.New(constant.DBError, "数据库错误")
	}
	return cf, nil
}

func (c *cityFacilityService) Get(rid int, cid int) *data.CityFacility {
	f := &data.CityFacility{}
	ok, err := db.Engine.Table(new(data.CityFacility)).Where("rid=? and cityId=?", rid, cid).Get(f)
	if err != nil {
		log.Println(err)
		return nil
	}
	if ok {
		return f
	}
	return nil
}

func (c *cityFacilityService) Load() {
	facs := make([]*data.CityFacility, 0)
	err := db.Engine.Find(&facs)
	if err != nil {
		log.Println(" load city_facility table error")
	}
	for _, v := range facs {
		c.facilities[v.CityId] = v
	}

	for _, cityFacility := range c.facilities {
		rid := cityFacility.RId
		_, ok := c.facilitiesByRId[rid]
		if ok == false {
			c.facilitiesByRId[rid] = make([]*data.CityFacility, 0)
		}
		c.facilitiesByRId[rid] = append(c.facilitiesByRId[rid], cityFacility)
	}
}

func (c *cityFacilityService) UpFacility(rid, cid int, fType int8) (*data.Facility, int) {
	facs := c.GetFacility1(rid, cid)
	result := &data.Facility{}
	for _, fac := range facs {

		if fac.Type == fType {

			if !fac.CanLV() {
				return nil, constant.UpError
			}

			maxlev := gameConfig.FacilityConf.MaxLevel(fType)
			if fac.GetLevel() >= maxlev {
				return nil, constant.UpError
			}

			need := gameConfig.FacilityConf.Need(fType, fac.GetLevel()+1)
			ok := RoleResService.TryUseNeed(rid, need)
			if !ok {
				return nil, constant.UpError
			}
			fac.UpTime = time.Now().Unix()
			result = fac

		}
	}
	marshal, _ := json.Marshal(facs)
	cfac := c.Get(rid, cid)
	cfac.Facilities = string(marshal)
	cfac.SyncExecute()
	return result, constant.OK
}

func (c *cityFacilityService) GetFacility(rid int, cid int) []data.Facility {
	cf := &data.CityFacility{}
	ok, err := db.Engine.Table(new(data.CityFacility)).Where("rid=? and cityId=?", rid, cid).Get(cf)
	if err != nil {
		log.Println(err)
		return nil
	}
	if ok {
		return cf.Facility()
	}
	return nil
}

func (c *cityFacilityService) GetFacility1(rid int, cid int) []*data.Facility {
	cf := &data.CityFacility{}
	ok, err := db.Engine.Table(new(data.CityFacility)).Where("rid=? and cityId=?", rid, cid).Get(cf)
	if err != nil {
		log.Println(err)
		return nil
	}
	if ok {
		return cf.Facility1()
	}
	return nil
}

func (c *cityFacilityService) GetByCid(cid int) *data.CityFacility {
	f := &data.CityFacility{}
	ok, err := db.Engine.Table(new(data.CityFacility)).Where("cityId=?", cid).Get(f)
	if err != nil {
		log.Println(err)
		return nil
	}
	if ok {
		return f
	}
	return nil
}

func (c *cityFacilityService) GetFaciltyLevel(cid int, fType int8) int8 {
	cf := c.GetByCid(cid)
	if cf == nil {
		return 0
	}
	facs := cf.Facility1()
	for _, v := range facs {
		if v.Type == fType {
			return v.GetLevel()
		}
	}
	return 0
}

func (c *cityFacilityService) GetCost(cid int) int8 {
	cf := c.GetByCid(cid)
	facility := cf.Facility()
	var cost int
	for _, f := range facility {
		if f.GetLevel() > 0 {
			values := gameConfig.FacilityConf.GetValues(f.Type, f.GetLevel())
			additions := gameConfig.FacilityConf.GetAdditions(f.Type)
			for i, aType := range additions {
				if aType == gameConfig.TypeWood {
					cost += values[i]
				}
			}
		}
	}
	return int8(cost)
}

func (c *cityFacilityService) GetCapacity(rid int) int {
	cfs, err := c.GetByRId(rid)
	var cap int
	if err != nil {
		for _, cf := range cfs {
			for _, f := range cf.Facility() {
				if f.GetLevel() > 0 {
					values := gameConfig.FacilityConf.GetValues(f.Type, f.GetLevel())
					additions := gameConfig.FacilityConf.GetAdditions(f.Type)
					for i, aType := range additions {
						if aType == gameConfig.TypeWarehouseLimit {
							cap += values[i]
						}
					}
				}
			}
		}
		log.Println("cityFacilityService GetYield err", err)
	}
	return cap
}

func (c *cityFacilityService) GetSoldier(cid int) int {
	cf := c.GetByCid(cid)
	facility := cf.Facility()
	var total int
	for _, f := range facility {
		if f.GetLevel() > 0 {
			values := gameConfig.FacilityConf.GetValues(f.Type, f.GetLevel())
			additions := gameConfig.FacilityConf.GetAdditions(f.Type)
			for i, aType := range additions {
				if aType == gameConfig.TypeSoldierLimit {
					total += values[i]
				}
			}
		}
	}
	return total
}
