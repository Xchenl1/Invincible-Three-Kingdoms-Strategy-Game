package data

import (
	"encoding/json"
	"log"
	"sgserver/db"
	"sgserver/server/game/gameConfig"
	"time"
)

var CityFacilityDao = &cityFacilityDao{
	cfChan: make(chan *CityFacility),
}

type cityFacilityDao struct {
	cfChan chan *CityFacility
}

type Facility struct {
	Name         string `json:"name"`
	PrivateLevel int8   `json:"level"` //等级，外部读的时候不能直接读，要用GetLevel
	Type         int8   `json:"type"`
	UpTime       int64  `json:"up_time"` //升级的时间戳，0表示该等级已经升级完成了
}

func (f Facility) GetLevel() int8 {
	if f.UpTime > 0 {
		cur := time.Now().Unix()
		cost := gameConfig.FacilityConf.CostTime(f.Type, f.PrivateLevel+1)
		if cur >= f.UpTime+int64(cost) {
			f.PrivateLevel += 1
			f.UpTime = 0
		}
	}
	return f.PrivateLevel
}
func (f *Facility) GetLevel1() {
	if f.UpTime > 0 {
		cur := time.Now().Unix()
		cost := gameConfig.FacilityConf.CostTime(f.Type, f.PrivateLevel+1)
		if cur >= f.UpTime+int64(cost) {
			f.PrivateLevel += 1
			f.UpTime = 0
		}
	}
}

type CityFacility struct {
	Id         int    `xorm:"id pk autoincr"`
	RId        int    `xorm:"rid"`
	CityId     int    `xorm:"cityId"`
	Facilities string `xorm:"facilities"`
}

func (c *CityFacility) TableName() string {
	return "city_facility"
}

func (c *CityFacility) Facility() []Facility {
	facilities := make([]Facility, 0)
	json.Unmarshal([]byte(c.Facilities), &facilities)
	return facilities
}

func (c *CityFacility) SyncExecute() {
	CityFacilityDao.cfChan <- c
}

func (f *Facility) CanLV() bool {
	f.GetLevel1()
	return f.UpTime == 0
}

func (cf *cityFacilityDao) running() {
	for true {
		select {
		case c := <-cf.cfChan:
			if c.Id > 0 {
				_, err := db.Engine.Table(c).ID(c.Id).Cols("facilities").Update(c)
				if err != nil {
					log.Println("db error", err)
				}
			} else {
				log.Println("update CityFacility fail, because id <= 0")
			}
		}
	}
}

func init() {
	go CityFacilityDao.running()
}
func (c *CityFacility) Facility1() []*Facility {
	facilities := make([]*Facility, 0)
	json.Unmarshal([]byte(c.Facilities), &facilities)
	return facilities
}
