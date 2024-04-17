package controller

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sgserver/constant"
	"sgserver/server/common"
	"sgserver/server/web/logic"
	"sgserver/server/web/model"
)

var DefaultAccountController = &AccountController{}

type AccountController struct {
}

func (a *AccountController) Register(ctx *gin.Context) {
	rq := &model.RegisterReq{}
	err := ctx.ShouldBind(rq)
	if err != nil {
		log.Println("参数不合法！")
		ctx.JSON(http.StatusOK, common.Error(constant.InvalidParam, "参数不合法！"))
		return
	}
	err = logic.DefaultAccountLogic.Register(rq)
	if err != nil {
		log.Println("注册业务出错！")
		ctx.JSON(http.StatusOK, common.Error(err.(*common.MyError).Code(), err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, common.Success(constant.OK, nil))
}
