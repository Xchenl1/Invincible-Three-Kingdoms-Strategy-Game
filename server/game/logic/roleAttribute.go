package logic

import (
	"encoding/json"
	"github.com/go-xorm/xorm"
	"log"
	"sgserver/db"
	"sgserver/net"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"sync"
)

// DefaultRoleAttrService 主要是联盟、征收次数、以及收藏位置
var DefaultRoleAttrService = &RoleAttrService{
	attrs: make(map[int]*data.RoleAttribute),
}

type RoleAttrService struct {
	mutex sync.RWMutex
	attrs map[int]*data.RoleAttribute
}

func (r *RoleAttrService) Load() {
	//加载
	t := make(map[int]*data.RoleAttribute)
	err := db.Engine.Find(t)
	if err != nil {
		log.Println(err)
	}
	//获取联盟id
	for _, v := range t {
		r.attrs[v.RId] = v
	}
	l := CoalitionService.ListCoalition()
	for _, c := range l {
		for _, rid := range c.MemberArray {
			attr, ok := r.attrs[rid]
			if ok {
				attr.UnionId = c.Id
			}
		}
	}
}

// TryCreate 创建玩家属性信息
func (r *RoleAttrService) TryCreate(rid int, req *net.WsMsgReq) error {
	rr := &data.RoleAttribute{}
	ok, err := db.Engine.Table(rr).Where("rid=?", rid).Get(rr)
	if err != nil {
		log.Println("玩家属性查询出错", err)
		return err
	}
	if !ok {
		//查询没有 进行初始化创建
		rr.RId = rid
		rr.ParentId = 0
		rr.UnionId = 0
		var err error
		if session := req.Context.Get("dbSession"); session != nil {
			_, err = session.(*xorm.Session).Table(rr).Insert(rr)
		} else {
			_, err = db.Engine.Table(rr).Insert(rr)
		}
		if err != nil {
			log.Println("玩家属性插入出错", err)
			return err
		}
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.attrs[rid] = rr
	}

	return nil
}

func (r *RoleAttrService) GetPosTags(rid int) []model.PosTag {
	r.mutex.RLock()
	rr, ok := r.attrs[rid]
	r.mutex.RUnlock()
	postTags := make([]model.PosTag, 0)
	if ok {
		err := json.Unmarshal([]byte(rr.PosTags), &postTags)
		if err != nil {
			log.Println("标记格式错误", err)
		}
	}
	return postTags
}

func (r *RoleAttrService) Get(rid int) *data.RoleAttribute {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	ra, ok := r.attrs[rid]
	if ok {
		return ra
	}
	return nil
}

func (r *RoleAttrService) GetUnion(rid int) int {
	return r.attrs[rid].UnionId
}
