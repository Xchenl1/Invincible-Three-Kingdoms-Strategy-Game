package logic

import (
	"log"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/net"
	"sgserver/server/common"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
	"sgserver/utils"
	"time"
)

var DefaultRoleService = &RoleService{}

type RoleService struct {
}

func (r *RoleService) EnterServer(uid int, rsp *model.EnterServerRsp, req *net.WsMsgReq) error {
	//获取角色 根据用户id 查找角色
	role := &data.Role{}
	// 开启事务
	session := db.Engine.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		log.Println("事务出错", err)
		return common.New(constant.DBError, "查询角色出错")
	}

	req.Context.Set("dbSession", session)

	// 查询角色是否存在
	ok, err := db.Engine.Table(role).Where("uid=?", uid).Get(role)
	if err != nil {
		log.Println("查询角色出错", err)
		return common.New(constant.DBError, "查询角色出错")
	}
	if !ok {
		return common.New(constant.RoleNotExist, "角色不存在")
	}

	// 查询该用户的资源是否存在
	rid := role.RId
	rsp.Role = role.ToModel().(model.Role)
	roleRes := &data.RoleRes{}
	ok, err = db.Engine.Table(roleRes).Where("rid=?", rid).Get(roleRes)
	if err != nil {
		log.Println("查询角色资源出错", err)
		return common.New(constant.DBError, "查询角色资源出错")
	}
	if !ok {
		//资源不存在  加载初始资源
		roleRes = &data.RoleRes{RId: role.RId,
			Wood:   gameConfig.Base.Role.Wood,
			Iron:   gameConfig.Base.Role.Iron,
			Stone:  gameConfig.Base.Role.Stone,
			Grain:  gameConfig.Base.Role.Grain,
			Gold:   gameConfig.Base.Role.Gold,
			Decree: gameConfig.Base.Role.Decree}
		_, err := session.Table(roleRes).Insert(roleRes)
		if err != nil {
			log.Println("插入角色失败", err)
			return common.New(constant.DBError, "数据库出错")
		}
	}
	// 强制转换
	rsp.RoleRes = roleRes.ToModel().(model.RoleRes)

	// 重新生成 token
	rsp.Token, err = utils.Award(rid)
	if err != nil {
		log.Println("生成token出错", err)
		return common.New(constant.SessionInvalid, "生成token出错")
	}
	rsp.Time = time.Now().UnixNano() / 1e6

	req.Conn.SetProperty("role", role)

	//获取角色属性信息 没有就创建
	if err := DefaultRoleAttrService.TryCreate(rid, req); err != nil {
		session.Rollback()
		return common.New(constant.DBError, "生成玩家属性出错")
	}
	//获取角色城市信息 没有就初始化一个并存储
	if err := Default.InitCity(role, req); err != nil {
		session.Rollback()
		return common.New(constant.DBError, "生成玩家属性出错")
	}
	err = session.Commit()
	if err != nil {
		log.Println("事务提交出错", err)
		return common.New(constant.DBError, "事务提交出错")
	}
	return nil
}

func (r *RoleService) GetRoleRes(rid int) (model.RoleRes, error) {
	roleRes := &data.RoleRes{}
	ok, err := db.Engine.Table(roleRes).Where("rid=?", rid).Get(roleRes)
	if err != nil {
		log.Println("查询角色资源出错", err)
		return model.RoleRes{}, common.New(constant.DBError, "查询角色资源出错")
	}
	if ok {
		return roleRes.ToModel().(model.RoleRes), nil
	}
	return model.RoleRes{}, nil
}

func (r *RoleService) Get(rid int) *data.Role {
	role := &data.Role{}
	ok, err := db.Engine.Table(role).Where("rid=?", rid).Get(role)
	if err != nil {
		log.Println("查询角色出错", err)
		return nil
	}
	if !ok {
		return nil
	}
	return role
}
