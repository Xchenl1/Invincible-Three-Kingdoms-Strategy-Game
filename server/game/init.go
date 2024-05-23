package game

import (
	"sgserver/db"
	"sgserver/net"
	"sgserver/server/game/controller"
	"sgserver/server/game/gameConfig"
	"sgserver/server/game/gameConfig/general"
	"sgserver/server/game/logic"
)

var Router = &net.Router{}

func Init() {
	db.TestDB()
	//加载初始资源
	gameConfig.Base.Load()
	//加载地图配置
	gameConfig.MapBuildConf.Load()
	//加载地图单元格配置
	gameConfig.MapRes.Load()
	//加载城池设施的相关配置
	gameConfig.FacilityConf.Load()

	logic.CityFacilityService.Load()
	//加载武将的相关配置
	general.General.Load()
	//
	general.GeneralBasic.Load()
	//加载技能配置信息
	gameConfig.Skill.Load()

	//加载所有建筑信息
	logic.DefaultRoleBuildService.Load()
	//加载所有城池信息
	logic.Default.Load()
	//加载联盟的初始化信息
	logic.CoalitionService.Load()
	//加载所有的角色属性
	logic.DefaultRoleAttrService.Load()
	logic.BeforeInit()
	// 加载

	initRouter()
}

func initRouter() {
	controller.DefaultRoleHandler.InitRouter(Router)
	controller.DefaultNationMapHandler.InitRouter(Router)
	controller.GeneralHandler.InitRouter(Router)
	controller.ArmyHandler.InitRouter(Router)
	controller.WarHandler.InitRouter(Router)
	controller.SkillHandler.InitRouter(Router)
	controller.InteriorController.Router(Router)
	controller.CoalitionController.Router(Router)
	controller.CityController.Router(Router)
}
