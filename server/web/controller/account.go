package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

var DefaultAccountController = &AccountController{}

type AccountController struct {
}

func (a *AccountController) Register(ctx *gin.Context) {
	fmt.Println("--register")
}
