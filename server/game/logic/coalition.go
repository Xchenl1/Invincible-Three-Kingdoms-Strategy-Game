package logic

import (
	"log"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/server/common"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"sync"
)

var CoalitionService = &coalitionService{
	unions: make(map[int]*data.Coalition),
}

type coalitionService struct {
	mutex  sync.RWMutex
	unions map[int]*data.Coalition
}

func (c *coalitionService) Load() {
	rr := make([]*data.Coalition, 0)
	err := db.Engine.Table(new(data.Coalition)).Where("state=?", data.UnionRunning).Find(&rr)
	if err != nil {
		log.Println("coalitionService load error", err)
	}
	for _, v := range rr {
		c.unions[v.Id] = v
	}
}

func (c *coalitionService) List() ([]model.Union, error) {
	r := make([]model.Union, 0)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	for _, coalition := range c.unions {
		union := coalition.ToModel().(model.Union)
		//盟主和副盟主信息
		main := make([]model.Major, 0)
		if role := DefaultRoleService.Get(coalition.Chairman); role != nil {
			m := model.Major{Name: role.NickName, RId: role.RId, Title: model.UnionChairman}
			main = append(main, m)
		}
		if role := DefaultRoleService.Get(coalition.ViceChairman); role != nil {
			m := model.Major{Name: role.NickName, RId: role.RId, Title: model.UnionChairman}
			main = append(main, m)
		}
		union.Major = main
		r = append(r, union)
	}
	return r, nil
}

func (c *coalitionService) ListCoalition() []*data.Coalition {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	uns := make([]*data.Coalition, 0)
	for _, v := range c.unions {
		uns = append(uns, v)
	}
	return uns
}

func (c *coalitionService) Get(id int) (model.Union, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	coalition, ok := c.unions[id]
	if ok {
		union := coalition.ToModel().(model.Union)
		//盟主和副盟主信息
		main := make([]model.Major, 0)
		if role := DefaultRoleService.Get(coalition.Chairman); role != nil {
			m := model.Major{Name: role.NickName, RId: role.RId, Title: model.UnionChairman}
			main = append(main, m)
		}
		if role := DefaultRoleService.Get(coalition.ViceChairman); role != nil {
			m := model.Major{Name: role.NickName, RId: role.RId, Title: model.UnionChairman}
			main = append(main, m)
		}
		union.Major = main
		return union, nil
	}
	return model.Union{}, nil
}
func (c *coalitionService) GetCoalition(id int) *data.Coalition {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	coa, ok := c.unions[id]
	if ok {
		return coa
	}
	return nil
}

func (c *coalitionService) GetListApply(unionId int, state int) ([]model.ApplyItem, error) {
	applys := make([]data.CoalitionApply, 0)
	err := db.Engine.Table(new(data.CoalitionApply)).
		Where("union_id=? and state=?", unionId, state).
		Find(&applys)
	if err != nil {
		log.Println("coalitionService GetListApply find error", err)
		return nil, common.New(constant.DBError, "数据库错误")
	}
	ais := make([]model.ApplyItem, 0)
	for _, v := range applys {
		var ai model.ApplyItem
		ai.Id = v.Id
		role := DefaultRoleService.Get(v.RId)
		ai.NickName = role.NickName
		ai.RId = role.RId
		ais = append(ais, ai)
	}
	return ais, nil
}
