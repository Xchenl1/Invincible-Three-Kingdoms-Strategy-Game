package gameConfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sgserver/server/game/global"
)

type NationalMap struct {
	MId   int  `xorm:"mid"`
	X     int  `xorm:"x"`
	Y     int  `xorm:"y"`
	Type  int8 `xorm:"type"`
	Level int8 `xorm:"level"`
}

type mapRes struct {
	Confs    map[int]NationalMap
	SysBuild map[int]NationalMap
}

type mapData struct {
	Width  int     `json:"w"`
	Height int     `json:"h"`
	List   [][]int `json:"list"`
}

const (
	MapBuildSysFortress = 50 //系统要塞
	MapBuildSysCity     = 51 //系统城市
	MapBuildFortress    = 56 //玩家要塞
)

var MapRes = &mapRes{
	Confs:    make(map[int]NationalMap),
	SysBuild: make(map[int]NationalMap),
}

const mapFile = "/conf/game/map.json"

func (m *mapRes) Load() {
	//获取当前文件路径
	currentDir, _ := os.Getwd()
	//配置文件位置
	cf := currentDir + mapFile
	//打包后 程序参数加入配置文件路径
	if len(os.Args) > 1 {
		if path := os.Args[1]; path != "" {
			cf = path + mapFile
		}
	}
	data, err := ioutil.ReadFile(cf)
	if err != nil {
		log.Println("地图读取失败")
		panic(err)
	}
	mapData := &mapData{}
	err = json.Unmarshal(data, mapData)
	if err != nil {
		log.Println("地图格式定义失败")
		panic(err)
	}
	global.MapWith = mapData.Width
	global.MapHeight = mapData.Height

	log.Println("map len", len(mapData.List))
	for index, v := range mapData.List {
		t := int8(v[0])
		l := int8(v[1])
		nm := NationalMap{
			X:     index % global.MapWith,
			Y:     index / global.MapHeight,
			MId:   index,
			Type:  t,
			Level: l,
		}
		m.Confs[index] = nm
		if t == MapBuildSysFortress || t == MapBuildSysCity {
			m.SysBuild[index] = nm
		}
	}

}
