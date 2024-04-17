package logic

import (
	"log"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/server/common"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/global"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"sgserver/utils"
	"sync"
)

var DefaultRoleBuildService = &RoleBuildService{
	posRB:  make(map[int]*data.MapRoleBuild),
	roleRB: make(map[int][]*data.MapRoleBuild),
}

type RoleBuildService struct {
	mutex  sync.RWMutex
	posRB  map[int]*data.MapRoleBuild
	roleRB map[int][]*data.MapRoleBuild
}

func (r *RoleBuildService) Load() {
	//加载系统建筑到数据库中
	total, err := db.Engine.
		Where("type=? or type=?", gameConfig.MapBuildSysCity, gameConfig.MapBuildSysFortress).
		Count(new(data.MapRoleBuild))
	if err != nil {
		panic(err)
	}
	sysBuild := gameConfig.MapRes.SysBuild
	if total != int64(len(sysBuild)) {
		//对不上，需要将系统建筑存入数据库
		//先删除 后插入
		db.Engine.
			Where("type=? or type=?", gameConfig.MapBuildSysCity, gameConfig.MapBuildSysFortress).
			Delete(new(data.MapRoleBuild))
		for _, v := range sysBuild {
			build := data.MapRoleBuild{
				RId:   0,
				Type:  v.Type,
				Level: v.Level,
				X:     v.X,
				Y:     v.Y,
			}
			build.Init()
			db.Engine.InsertOne(&build)
		}
	}
	//查找所有的角色建筑
	dbRb := make(map[int]*data.MapRoleBuild)
	db.Engine.Find(dbRb)
	//将其转换为 角色id-建筑 位置-建筑
	for _, v := range dbRb {
		v.Init()
		pos := global.ToPosition(v.X, v.Y)
		r.posRB[pos] = v
		_, ok := r.roleRB[v.RId]
		if !ok {
			r.roleRB[v.RId] = make([]*data.MapRoleBuild, 0)
		} else {
			r.roleRB[v.RId] = append(r.roleRB[v.RId], v)
		}
	}
}

func (r *RoleBuildService) GetBuilds(rid int) ([]model.MapRoleBuild, error) {
	mrs := make([]data.MapRoleBuild, 0)
	mr := &data.MapRoleBuild{}
	err := db.Engine.Table(mr).Where("rid=?", rid).Find(&mrs)
	if err != nil {
		log.Println("建筑查询出错", err)
		return nil, common.New(constant.DBError, "建筑查询出错")
	}
	modelMrs := make([]model.MapRoleBuild, 0)
	for _, v := range mrs {
		modelMrs = append(modelMrs, v.ToModel().(model.MapRoleBuild))
	}
	return modelMrs, nil
}

func (r *RoleBuildService) ScanBlock(req *model.ScanBlockReq) ([]model.MapRoleBuild, error) {
	x := req.X
	y := req.Y
	length := req.Length
	rb := make([]model.MapRoleBuild, 0)

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
			if ok && (v.RId != 0 || v.IsSysCity() || v.IsSysFortress()) {
				rb = append(rb, v.ToModel().(model.MapRoleBuild))
			}
		}
	}
	return rb, nil
}

func (r *RoleBuildService) GetYield(rid int) data.Yield {
	var y data.Yield
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	builds, ok := r.roleRB[rid]
	if ok {
		for _, b := range builds {
			y.Iron += b.Iron
			y.Wood += b.Wood
			y.Grain += b.Grain
			y.Stone += b.Grain
		}
	}
	return y
}
