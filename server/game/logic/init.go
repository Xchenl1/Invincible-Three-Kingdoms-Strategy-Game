package logic

import "sgserver/server/game/model/data"

func BeforeInit() {
	//接口赋值
	data.GetYield = RoleResService.GetYield
	data.GetUnion = DefaultRoleAttrService.GetUnion
}
