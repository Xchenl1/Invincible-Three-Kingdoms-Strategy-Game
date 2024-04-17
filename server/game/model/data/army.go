package data

import (
	"encoding/json"
	"fmt"
	"github.com/go-xorm/xorm"
	"sgserver/db"
	"sgserver/server/game/model"
	"time"
)

const (
	ArmyCmdIdle        = 0 //空闲
	ArmyCmdAttack      = 1 //攻击
	ArmyCmdDefend      = 2 //驻守
	ArmyCmdReclamation = 3 //屯垦
	ArmyCmdBack        = 4 //撤退
	ArmyCmdConscript   = 5 //征兵
	ArmyCmdTransfer    = 6 //调动
)

const (
	ArmyStop    = 0
	ArmyRunning = 1
)

var ArmyDao = &armyDao{
	aChan: make(chan *Army, 100),
}

type armyDao struct {
	aChan chan *Army
}

func (a *armyDao) running() {
	for {
		select {
		case army := <-a.aChan:
			if army.Id > 0 {
				db.Engine.Table(army).ID(army.Id).Cols(
					"soldiers", "generals", "conscript_times",
					"conscript_cnts", "cmd", "from_x", "from_y", "to_x",
					"to_y", "start", "end").Update(army)
			}
		}
	}
}

func init() {
	go ArmyDao.running()
}

// 军队
type Army struct {
	Id                 int64      `xorm:"id pk autoincr"`
	RId                int        `xorm:"rid"`
	CityId             int        `xorm:"cityId"`
	Order              int8       `xorm:"order"`
	Generals           string     `xorm:"generals"`
	Soldiers           string     `xorm:"soldiers"`
	ConscriptTimes     string     `xorm:"conscript_times"` //征兵结束时间，json数组
	ConscriptCnts      string     `xorm:"conscript_cnts"`  //征兵数量，json数组
	Cmd                int8       `xorm:"cmd"`
	FromX              int        `xorm:"from_x"`
	FromY              int        `xorm:"from_y"`
	ToX                int        `xorm:"to_x"`
	ToY                int        `xorm:"to_y"`
	Start              time.Time  `json:"-"xorm:"start"`
	End                time.Time  `json:"-"xorm:"end"`
	State              int8       `xorm:"-"` //状态:0:running,1:stop
	GeneralArray       []int      `json:"-" xorm:"-"`
	SoldierArray       []int      `json:"-" xorm:"-"`
	ConscriptTimeArray []int64    `json:"-" xorm:"-"`
	ConscriptCntArray  []int      `json:"-" xorm:"-"`
	Gens               []*General `json:"-" xorm:"-"`
	CellX              int        `json:"-" xorm:"-"`
	CellY              int        `json:"-" xorm:"-"`
}

func (a *Army) BeforeUpdate() {
	a.beforeModify()
}

func (a *Army) beforeModify() {
	data, _ := json.Marshal(a.GeneralArray)
	a.Generals = string(data)

	data, _ = json.Marshal(a.SoldierArray)
	a.Soldiers = string(data)

	data, _ = json.Marshal(a.ConscriptTimeArray)
	a.ConscriptTimes = string(data)

	data, _ = json.Marshal(a.ConscriptCntArray)
	a.ConscriptCnts = string(data)
}

func (a *Army) BeforeInsert() {
	a.beforeModify()
}

func (a *Army) TableName() string {
	return "army"
}

func (a *Army) AfterSet(name string, cell xorm.Cell) {
	if name == "generals" {
		a.GeneralArray = []int{0, 0, 0}
		if cell != nil {
			gs, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(gs, &a.GeneralArray)
				fmt.Println(a.GeneralArray)
			}
		}
	} else if name == "soldiers" {
		a.SoldierArray = []int{0, 0, 0}
		if cell != nil {
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &a.SoldierArray)
				fmt.Println(a.SoldierArray)
			}
		}
	} else if name == "conscript_times" {
		a.ConscriptTimeArray = []int64{0, 0, 0}
		if cell != nil {
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &a.ConscriptTimeArray)
				fmt.Println(a.ConscriptTimeArray)
			}
		}
	} else if name == "conscript_cnts" {
		a.ConscriptCntArray = []int{0, 0, 0}
		if cell != nil {
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &a.ConscriptCntArray)
				fmt.Println(a.ConscriptCntArray)
			}
		}
	}
}

func (a *Army) ToModel() interface{} {
	p := model.Army{}
	p.CityId = a.CityId
	p.Id = int(a.Id)
	p.UnionId = 0
	p.Order = a.Order
	p.Generals = a.GeneralArray
	p.Soldiers = a.SoldierArray
	p.ConTimes = a.ConscriptTimeArray
	p.ConCnts = a.ConscriptCntArray
	p.Cmd = a.Cmd
	p.State = a.State
	p.FromX = a.FromX
	p.FromY = a.FromY
	p.ToX = a.ToX
	p.ToY = a.ToY
	p.Start = a.Start.Unix()
	p.End = a.End.Unix()
	return p
}
func (a *Army) PositionCanModify(pos int) bool {
	if pos >= 3 || pos < 0 {
		return false
	}

	if a.Cmd == ArmyCmdIdle {
		return true
	} else if a.Cmd == ArmyCmdConscript {
		endTime := a.ConscriptTimeArray[pos]
		return endTime == 0
	} else {
		return false
	}
}

func (a *Army) SyncExecute() {
	ArmyDao.aChan <- a
}
