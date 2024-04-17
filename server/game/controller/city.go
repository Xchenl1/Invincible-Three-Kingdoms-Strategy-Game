package controller

import (
	"github.com/mitchellh/mapstructure"
	"sgserver/constant"
	"sgserver/net"
	"sgserver/server/game/logic"
	"sgserver/server/game/middleware"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
)

var CityController = &cityController{}

type cityController struct {
}

func (c *cityController) Router(router *net.Router) {
	g := router.Group("city")
	g.Use(middleware.Log())
	g.AddRouter("facilities", c.facilities, middleware.CheckRole())
	g.AddRouter("upFacility", c.upFacility, middleware.CheckRole())
}

func (c *cityController) facilities(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	// 城池 id 查询城池信息
	//
	reqObj := &model.FacilitiesReq{}
	rspObj := &model.FacilitiesRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = constant.OK

	// 拿到角色
	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)

	// 查询城池
	city, ok := logic.Default.Get(rspObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	// 查询城池设施
	fac := logic.CityFacilityService.GetFacility(role.RId, reqObj.CityId)
	if fac == nil {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	rspObj.CityId = reqObj.CityId
	rspObj.Facilities = make([]model.Facility, len(fac))
	for index, v := range fac {
		rspObj.Facilities[index].Type = v.Type
		rspObj.Facilities[index].Name = v.Name
		rspObj.Facilities[index].Level = v.GetLevel()
		rspObj.Facilities[index].UpTime = v.UpTime
	}
}

func (c *cityController) upFacility(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.UpFacilityReq{}
	rspObj := &model.UpFacilityRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = constant.OK
	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)

	//获取城池设施
	city, ok := logic.Default.Get(rspObj.CityId)
	if !ok {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	//获取需要升级设施
	fac := logic.CityFacilityService.GetFacility(role.RId, rspObj.CityId)
	if fac == nil {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	out, errCode := logic.CityFacilityService.UpFacility(role.RId, reqObj.CityId, reqObj.FType)
	rsp.Body.Code = errCode
	if errCode == constant.OK {
		rspObj.Facility.Level = out.GetLevel()
		rspObj.Facility.Type = out.Type
		rspObj.Facility.Name = out.Name
		rspObj.Facility.UpTime = out.UpTime

		// 查询角色的资源
		res := logic.RoleResService.Get(role.RId)
		if res != nil {
			rspObj.RoleRes = res.ToModel().(model.RoleRes)
		}
	}
}
