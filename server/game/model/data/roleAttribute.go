package data

import (
	"log"
	"sgserver/db"
	"sgserver/server/game/model"
	"time"
)

var RoleAttrDao = &roleAttrDao{
	raChan: make(chan *RoleAttribute, 100),
}

type roleAttrDao struct {
	raChan chan *RoleAttribute
}

func init() {
	go RoleAttrDao.running()
}
func (d roleAttrDao) Push(r *RoleAttribute) {
	d.raChan <- r
}

func (r *roleAttrDao) running() {
	for {
		select {
		case ra := <-r.raChan:
			if ra.Id > 0 {
				_, err := db.Engine.Table(ra).ID(ra.Id).Cols(
					"parent_id", "collect_times", "last_collect_time", "pos_tags").Update(ra)
				if err != nil {
					log.Println("roleAttrDao update error", err)
				}
			}
		}
	}
}

type RoleAttribute struct {
	Id              int            `xorm:"id pk autoincr"`
	RId             int            `xorm:"rid"`
	UnionId         int            `xorm:"-"`                 //联盟id
	ParentId        int            `xorm:"parent_id"`         //上级id（被沦陷）
	CollectTimes    int8           `xorm:"collect_times"`     //征收次数
	LastCollectTime time.Time      `xorm:"last_collect_time"` //最后征收的时间
	PosTags         string         `xorm:"pos_tags"`          //位置标记
	PosTagArray     []model.PosTag `xorm:"-"`
}

func (r *RoleAttribute) TableName() string {
	return "role_attribute"
}

func (r *RoleAttribute) SyncExecute() {
	RoleAttrDao.Push(r)
}
