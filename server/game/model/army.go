package model

type ArmyListReq struct {
	CityId int `json:"cityId"`
}

type ArmyListRsp struct {
	CityId int    `json:"cityId"`
	Armys  []Army `json:"armys"`
}

type DisposeReq struct {
	CityId    int  `json:"cityId"`    //城市id
	GeneralId int  `json:"generalId"` //将领id
	Order     int8 `json:"order"`     //第几队，1-5队
	Position  int  `json:"position"`  //位置，-1到2,-1是解除该武将上阵状态
}

type DisposeRsp struct {
	Army Army `json:"army"`
}

type ConscriptReq struct {
	ArmyId int   `json:"armyId"` // 队伍id
	Cnts   []int `json:"cnts"`   // 征兵人数[20,20,0]
}

type ConscriptRsp struct {
	Army    Army    `json:"army"`
	RoleRes RoleRes `json:"role_res"`
}
