package logic

import (
	"github.com/go-xorm/xorm"
	"log"
	"math/rand"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/net"
	"sgserver/server/common"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/global"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"sgserver/utils"
	"sync"
	"time"
)

var Default = &RoleCityService{
	dbRB:   make(map[int]*data.MapRoleCity),
	posRB:  make(map[int]*data.MapRoleCity),
	roleRB: make(map[int][]*data.MapRoleCity),
}

type RoleCityService struct {
	mutex  sync.RWMutex
	dbRB   map[int]*data.MapRoleCity
	posRB  map[int]*data.MapRoleCity
	roleRB map[int][]*data.MapRoleCity
}

func (r *RoleCityService) GetMainCity(id int) *data.MapRoleCity {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	rsc, ok := r.roleRB[id] //有问题
	if ok {
		for _, v := range rsc {
			if v.IsMain == 0 {
				return v
			}
		}
	}
	return nil
}

func (r *RoleCityService) Load() {
	//dbCity := make(map[int]*data.MapRoleCity)
	err := db.Engine.Find(r.dbRB)
	if err != nil {
		log.Println("RoleCityService load role_city table error")
		return
	}

	//转成posCity、roleCity
	for _, v := range r.dbRB {
		posId := global.ToPosition(v.X, v.Y)
		r.posRB[posId] = v
		_, ok := r.roleRB[v.RId]
		if ok == false {
			r.roleRB[v.RId] = make([]*data.MapRoleCity, 0)
		}
		r.roleRB[v.RId] = append(r.roleRB[v.RId], v)
	}
	//耐久度计算 后续做

}

// 初始化城池
func (r *RoleCityService) InitCity(role *data.Role, req *net.WsMsgReq) error {
	rc := &data.MapRoleCity{}
	ok, err := db.Engine.Table(rc).Where("rid=?", role.RId).Get(rc)
	if err != nil {
		log.Println("查询角色城市出错", err)
		return common.New(constant.DBError, "查询角色城市出错")
	}
	if !ok {
		for {
			//没有城池 初始化 条件系统城市5格内不能有玩家城池
			x := rand.Intn(global.MapWith)
			y := rand.Intn(global.MapHeight)
			//判断是否符合创建条件
			if r.IsCanBuild(x, y) {
				//建的肯定是主城
				rc.RId = role.RId
				rc.Y = y
				rc.X = x
				rc.CreatedAt = time.Now()
				rc.Name = role.NickName
				rc.CurDurable = gameConfig.Base.City.Durable

				// 实现 map_role_city 表数据插入 建立 rid 与 cityid 关系
				if session := req.Context.Get("dbSession"); session != nil {
					_, err = session.(*xorm.Session).Table(rc).Insert(rc)
				} else {
					_, err = db.Engine.Table(rc).Insert(rc)
				}
				if err != nil {
					log.Println("插入玩家城市出错", err)
					return common.New(constant.DBError, "插入玩家城市出错")
				}

				posId := global.ToPosition(x, y)
				r.posRB[posId] = rc
				_, ok = r.roleRB[role.RId]
				if !ok {
					r.roleRB[role.RId] = make([]*data.MapRoleCity, 0)
				} else {
					r.roleRB[role.RId] = append(r.roleRB[role.RId], rc)
				}
				r.dbRB[rc.CityId] = rc
				//生成城市设施
				if err := CityFacilityService.TryCreate(rc.CityId, rc.RId, req); err != nil {
					log.Println("插入城池设施出错", err)
					return common.New(err.(*common.MyError).Code(), err.Error())
				}
				break
			}
		}

	}
	return nil
}

func (r *RoleCityService) IsCanBuild(x int, y int) bool {
	confs := gameConfig.MapRes.Confs
	pIndex := global.ToPosition(x, y)
	_, ok := confs[pIndex]
	if !ok {
		return false
	}

	//城池 1范围内 不能超过边界
	if x+1 >= global.MapWith || y+1 >= global.MapHeight || y-1 < 0 || x-1 < 0 {
		return false
	}

	sysBuild := gameConfig.MapRes.SysBuild
	for _, v := range sysBuild {
		if v.Type == gameConfig.MapBuildSysCity {
			//5格内不能有玩家城池
			if x >= v.X-5 &&
				x <= v.X+5 &&
				y >= v.Y-5 &&
				y <= v.Y+5 {
				return false
			}
		}
	}
	//玩家城池的5格内 也不能创建城池
	for i := x - 5; i <= x+5; i++ {
		for j := y - 5; j <= y+5; j++ {
			posId := global.ToPosition(i, j)
			_, ok := r.posRB[posId]
			if ok {
				return false
			}
		}
	}
	return true
}

func (r *RoleCityService) GetCitys(rid int) ([]model.MapRoleCity, error) {
	mrs := make([]data.MapRoleCity, 0)
	mr := &data.MapRoleCity{}
	err := db.Engine.Table(mr).Where("rid=?", rid).Find(&mrs)
	if err != nil {
		log.Println("城池查询出错", err)
		return nil, common.New(constant.DBError, "城池查询出错")
	}
	modelMrs := make([]model.MapRoleCity, 0)
	for _, v := range mrs {
		modelMrs = append(modelMrs, v.ToModel().(model.MapRoleCity))
	}
	return modelMrs, nil
}

func (r *RoleCityService) ScanBlock(req *model.ScanBlockReq) ([]model.MapRoleCity, error) {
	rb := make([]model.MapRoleCity, 0)
	x := req.X
	y := req.Y
	length := req.Length
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return rb, nil
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	maxX := utils.MinInt(global.MapWith, x+length-1)
	maxY := utils.MinInt(global.MapHeight, y+length-1)

	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			posId := global.ToPosition(i, j)
			v, ok := r.posRB[posId]
			if ok {
				rb = append(rb, v.ToModel().(model.MapRoleCity))
			}
		}
	}
	return rb, nil
}

func (r *RoleCityService) Get(id int) (*data.MapRoleCity, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	city, ok := r.dbRB[id] //有问题
	if ok {
		return city, ok
	}
	return nil, ok
}

func (r *RoleCityService) GetCityCost(cid int) int8 {
	return CityFacilityService.GetCost(cid) + gameConfig.Base.City.Cost
}
