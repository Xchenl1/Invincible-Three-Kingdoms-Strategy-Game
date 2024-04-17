package logic

import (
	"log"
	"sgserver/constant"
	"sgserver/db"
	"sgserver/server/common"
	"sgserver/server/game/model"
	"sgserver/server/game/model/data"
)

var DefaultWarService = &WarService{}

type WarService struct {
}

func (w *WarService) GetWarReports(rid int) ([]model.WarReport, error) {
	mrs := make([]data.WarReport, 0)
	mr := &data.WarReport{}
	err := db.Engine.Table(mr).
		Where("a_rid=? or d_rid=?", rid, rid).
		Desc("ctime").
		Limit(30, 0).
		Find(&mrs)
	if err != nil {
		log.Println("战报查询出错", err)
		return nil, common.New(constant.DBError, "战报查询出错")
	}
	modelMrs := make([]model.WarReport, 0)
	for _, v := range mrs {
		modelMrs = append(modelMrs, v.ToModel().(model.WarReport))
	}
	return modelMrs, nil
}
